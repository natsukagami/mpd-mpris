package mpris

import (
	"fmt"
	"sort"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"

	"github.com/natsukagami/mpd-mpris/mpd"
)

const (
	// TrackIDFormat is the formatter string for a track ID.
	TrackIDFormat = trackIDPrefix + "%d"

	trackIDPrefix = "/org/mpd/Tracks/"
)

// NoTrack represents a TrackID of "no track selected".
const NoTrack dbus.ObjectPath = "/org/mpris/MediaPlayer2/TrackList/NoTrack"

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.TrackList` interface.

// Tracklist is an implementation of the TrackList interface for `Instance`.
// https://specifications.freedesktop.org/mpris-spec/latest/TrackList_Interface.html
type Tracklist struct {
	ins   *Instance
	props map[string]*prop.Prop

	lastPlaylistVersion int
	songs               []mpd.Song
}

func (tl *Tracklist) update(status *mpd.Status) error {
	if status.PlaylistVersion == tl.lastPlaylistVersion {
		return nil // no change
	}
	changes, err := tl.ins.mpd.PlaylistChanges(tl.lastPlaylistVersion)
	if err != nil {
		return err
	}

	if len(changes) < status.PlaylistLength {
		tl.applyChanges(status, changes)
		return nil
	}

	songs, err := tl.getPlaylist()
	if err != nil {
		return err
	}

	tl.ins.setProp("TrackList", "Tracks", dbus.MakeVariant(songsToURIs(songs)))
	// emit `TrackListReplaced`
	// https://specifications.freedesktop.org/mpris-spec/latest/Track_List_Interface.html#Signal:TrackListReplaced
	tl.ins.emit("TrackList", "TrackListReplaced", songs, idToURI(status.Song))

	tl.lastPlaylistVersion = status.PlaylistVersion
	tl.songs = songs

	return nil
}

// mpd's change list model (`plchanges`, see https://mpd.readthedocs.io/en/latest/protocol.html#command-plchanges)
// is a bit complicated to change to sequential insert/remove model.
func (tl *Tracklist) applyChanges(status *mpd.Status, cs []mpd.Song) {
	sort.Slice(cs, func(i, j int) bool { return cs[i].Pos < cs[j].Pos })
	// Change represents a playlist change.
	type Change struct {
		ChangedPos int
		Song       *mpd.Song
	}
	toInsert := make([]Change, 0)
	toRemove := make([]int, 0)

	offset := 0
	for _, _change := range cs {
		change := _change
		i := change.Pos
		ptr := i + offset
		// no change if not in cs
		if ptr < len(tl.songs) && tl.songs[ptr].ID == change.ID {
			continue
		}
		// opt: if tl.ids[oldPtr+1] == change, we can remove instead
		if ptr+1 < len(tl.songs) && tl.songs[ptr+1].ID == change.ID {
			toRemove = append(toRemove, tl.songs[ptr].ID)
			offset += 1
			continue
		}
		// we have to insert
		toInsert = append(toInsert, Change{ChangedPos: i - 1, Song: &change})
		offset -= 1
	}

	// send all removals first
	for ptr := status.PlaylistLength + offset; ptr < len(tl.songs); ptr++ {
		tl.ins.emit("TrackList", "TrackRemoved", idToURI(tl.songs[ptr].ID))
	}
	for _, r := range toRemove {
		tl.ins.emit("TrackList", "TrackRemoved", idToURI(r))
	}
	// mutate songs
	for _, change := range cs {
		if change.Pos == len(tl.songs) {
			tl.songs = append(tl.songs, change)
		} else {
			tl.songs[change.Pos] = change
		}
	}
	tl.songs = tl.songs[:status.PlaylistLength]
	// send insertions left to right
	for _, c := range toInsert {
		id := -1
		if c.ChangedPos != -1 {
			id = tl.songs[c.ChangedPos].ID
		}
		tl.ins.emit("TrackList", "TrackAdded", MapFromSong(*c.Song), idToURI(id))
	}

	// update prop
	go tl.ins.setProp("TrackList", "Tracks", dbus.MakeVariant(songsToURIs(tl.songs)))

	tl.lastPlaylistVersion = status.PlaylistVersion
}

// getPlaylist returns the whole playlist. Also stores it in the tracklist.
func (tl *Tracklist) getPlaylist() ([]mpd.Song, error) {
	songs, err := tl.ins.mpd.PlaylistInfo(-1, -1)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func songsToURIs(songs []mpd.Song) []dbus.ObjectPath {
	songsMeta := make([]dbus.ObjectPath, 0, len(songs))
	for _, s := range songs {
		songsMeta = append(songsMeta, idToURI(s.ID))
	}
	return songsMeta
}

// Creates and populate a new tracklist.
func newTracklist(ins *Instance) (*Tracklist, error) {
	tl := &Tracklist{ins: ins}

	// Populate it with the current mpd playlist
	status, err := ins.mpd.Status()
	if err != nil {
		return nil, err
	}

	// load the current playlist info
	songs, err := tl.getPlaylist()
	if err != nil {
		return nil, err
	}

	tl.lastPlaylistVersion = status.PlaylistVersion
	tl.songs = songs

	// set the props
	tl.props = map[string]*prop.Prop{
		"CanEditTracks": {
			Value:    true,
			Writable: true,
			Emit:     prop.EmitTrue,
			Callback: nil,
		},
		"Tracks": {
			Value:    songsToURIs(songs),
			Writable: true,
			Emit:     prop.EmitInvalidates,
			Callback: nil,
		},
	}

	return tl, nil
}

// URI is an unique resource identifier.
// https://specifications.freedesktop.org/mpris-spec/latest/Track_List_Interface.html#Simple-Type:Uri
type URI string

// MetadataMap is a mapping from metadata attribute names to values.
// https://specifications.freedesktop.org/mpris-spec/latest/Track_List_Interface.html#Mapping:Metadata_Map
type MetadataMap map[string]interface{}

func (m *MetadataMap) nonEmptyString(field, value string) {
	if value != "" {
		(*m)[field] = value
	}
}

func (m *MetadataMap) nonEmptySlice(field string, values []string) {
	toAdd := []string{}
	for _, v := range values {
		if v != "" {
			toAdd = append(toAdd, v)
		}
	}
	if len(toAdd) > 0 {
		(*m)[field] = toAdd
	}
}

func idToURI(id int) dbus.ObjectPath {
	if id == -1 {
		return NoTrack
	}
	return dbus.ObjectPath(fmt.Sprintf(TrackIDFormat, id))
}

// MapFromSong returns a MetadataMap from the Song struct in mpd.
func MapFromSong(s mpd.Song) MetadataMap {
	if s.ID == -1 {
		// No song
		return MetadataMap{
			"mpris:trackid": NoTrack,
		}
	}

	m := &MetadataMap{
		"mpris:trackid": idToURI(s.ID),
		"mpris:length":  s.Duration / time.Microsecond,
	}

	m.nonEmptyString("xesam:album", s.Album)
	m.nonEmptyString("xesam:title", s.Title)
	m.nonEmptyString("xesam:url", s.Filepath)
	m.nonEmptyString("xesam:contentCreated", s.Date)
	m.nonEmptySlice("xesam:albumArtist", []string{s.AlbumArtist})
	m.nonEmptySlice("xesam:artist", []string{s.Artist})
	m.nonEmptySlice("xesam:genre", []string{s.Genre})

	if artURI, ok := s.AlbumArtURI(); ok {
		(*m)["mpris:artUrl"] = artURI
	}

	if s.Track != 0 {
		(*m)["xesam:trackNumber"] = s.Track
	}

	return *m
}
