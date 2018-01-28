package mpris

import (
	"fmt"
	"time"

	"github.com/natsukagami/mpd-mpris/mpd"
)

// TrackIDFormat is the formatter string for a track ID.
const TrackIDFormat = "/org/mpris/MediaPlayer2/TrackList/%d"

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

// MapFromSong returns a MetadataMap from the Song struct in mpd.
func MapFromSong(s mpd.Song) MetadataMap {
	if s.ID == -1 {
		// No song
		return MetadataMap{
			"mpris:trackid": "/org/mpris/MediaPlayer2/TrackList/NoTrack",
		}
	}
	return MetadataMap{
		"mpris:trackid":     fmt.Sprintf(TrackIDFormat, s.ID),
		"mpris:length":      s.Duration / time.Microsecond,
		"xesam:album":       s.Album,
		"xesam:albumArtist": []string{s.AlbumArtist},
		"xesam:artist":      []string{s.Artist},
		"xesam:trackNumber": s.Track,
		"xesam:genre":       []string{s.Genre},
		"xesam:title":       s.Title,
	}
}
