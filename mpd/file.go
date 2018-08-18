package mpd

import (
	"math"
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
	"github.com/pkg/errors"
)

// Item represents an item in the file system.
// It could either be a File, a Directory or a PlaylistFile.
type Item interface {
	// The path to the item, relative to the library's root.
	Path() string
}

// ItemFromAttrs returns an Item from the given Attr struct.
func ItemFromAttrs(attr mpd.Attrs) (Item, error) {
	if _, ok := attr["file"]; ok {
		return FileFromAttrs(attr)
	}
	if _, ok := attr["directory"]; ok {
		return Directory{Attrs: attr}, nil
	}
	if _, ok := attr["playlist"]; ok {
		return PlaylistFile{Attrs: attr}, nil
	}
	return nil, errors.New("Not a valid item")
}

// File represents a music file.
type File struct {
	InLibrary   bool // Specifies whether the file is in the library
	Title       string
	Artist      string
	Genre       string
	Date        string // Has non-standard time format
	Album       string
	AlbumArtist string
	Track       int
	Duration    time.Duration // In miliseconds
	Attrs       mpd.Attrs     // Other attributes
}

// Path returns the path to the file.
func (f File) Path() string {
	return f.Attrs["file"]
}

// FileFromAttrs returns a File from the attributes map.
func FileFromAttrs(attr mpd.Attrs) (s File, err error) {
	if title, ok := attr["Title"]; ok {
		s.Title = title
		s.InLibrary = true
	} else {
		s.InLibrary = false
		return
	}
	s.Artist = attr["Artist"]
	s.Genre = attr["Genre"]
	s.Date = attr["Date"]
	s.Album = attr["Album"]
	s.AlbumArtist = attr["AlbumArtist"]
	if s.Track, err = strconv.Atoi(attr["Track"]); err != nil {
		err = nil
		// No track information
		s.Track = 0
	}
	// Handle duration-less files, set duration to maximum possible
	if d, ok := attr["duration"]; !ok || d == "" {
		s.Duration = time.Duration(math.MaxInt64)
	} else {
		var durationF float64
		if durationF, err = strconv.ParseFloat(d, 64); err != nil {
			err = errors.Wrap(err, "Parse duration")
			return
		}
		s.Duration = time.Duration(durationF * float64(time.Second))
	}
	s.Attrs = attr
	return
}

// Directory represents a directory in the library.
type Directory struct {
	Attrs mpd.Attrs
}

// Path returns the path to the directory.
func (d Directory) Path() string {
	return d.Attrs["directory"]
}

// PlaylistFile represents a Playlist in the library.
// No metadata about the playlist is stored here, only the file's information.
type PlaylistFile struct {
	Attrs mpd.Attrs
}

// Path returns the path to the directory.
func (p PlaylistFile) Path() string {
	return p.Attrs["playlist"]
}
