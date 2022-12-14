package mpris

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/natsukagami/mpd-mpris/mpd"
)

// TrackIDFormat is the formatter string for a track ID.
const TrackIDFormat = "/org/mpd/Tracks/%d"

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.TrackList` interface.

// TrackList is a DBus object satisfying the `org.mpris.MediaPlayer2.TrackList` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/TrackList_Interface.html
type TrackList struct {
	*Instance
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

// MapFromSong returns a MetadataMap from the Song struct in mpd.
func MapFromSong(s mpd.Song) MetadataMap {
	if s.ID == -1 {
		// No song
		return MetadataMap{
			"mpris:trackid": dbus.ObjectPath("/org/mpris/MediaPlayer2/TrackList/NoTrack"),
		}
	}

	m := &MetadataMap{
		"mpris:trackid": dbus.ObjectPath(fmt.Sprintf(TrackIDFormat, s.ID)),
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
