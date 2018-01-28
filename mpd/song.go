package mpd

import (
	"strconv"

	"github.com/fhs/gompd/mpd"
)

// Song represents a music file with metadata.
type Song struct {
	File
	ID int // The song's ID (within the playlist)
}

// SongFromAttrs returns a song from the attributes map.
func SongFromAttrs(attr mpd.Attrs) (s Song, err error) {
	if s.ID, err = strconv.Atoi(attr["Id"]); err != nil {
		s.ID = -1
		return s, nil
	}
	if s.File, err = FileFromAttrs(attr); err != nil {
		return
	}
	return
}
