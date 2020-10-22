package mpd

import (
	"time"

	"github.com/fhs/gompd/v2/mpd"
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
	Title       string
	Artist      string
	Genre       string
	Date        string // Has non-standard time format
	Album       string
	AlbumArtist string
	Track       int
	Duration    time.Duration
	Attrs       mpd.Attrs // Other attributes
}

// Path returns the path to the file.
func (f File) Path() string {
	return f.Attrs["file"]
}

// FileFromAttrs returns a File from the attributes map.
func FileFromAttrs(attr mpd.Attrs) (s File, err error) {
	p := &parseMap{m: attr}

	if !p.String("Title", &s.Title, true) {
		s.Title = "unknown title"
	}
	// All the following values can be empty
	if !p.String("Artist", &s.Artist, true) {
		s.Artist = "unknown artist"
	}
	p.String("Genre", &s.Genre, true)
	p.String("Date", &s.Date, true)
	p.String("Album", &s.Album, true)
	p.String("AlbumArtist", &s.AlbumArtist, true)

	p.Int("Track", &s.Track, true)
	// Handle duration-less files, set duration to 0 and do not convert it to a
	// metadata field
	durationF := 0.0
	p.Float("duration", &durationF, true)
	s.Duration = time.Duration(durationF * float64(time.Second))

	err = p.Err
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
