package mpd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/fhs/gompd/v2/mpd"
)

var albumArtLock sync.Mutex
var albumArtURI string

func init() {
	mpdTemp := filepath.Join(os.TempDir(), "mpd_mpris")
	if err := os.MkdirAll(mpdTemp, 0x644); err != nil {
		log.Println("Cannot create temp file for album art, we don't support them then!")
		return
	}
	f, err := ioutil.TempFile(mpdTemp, "artwork_")
	if err != nil {
		log.Println("Cannot create temp file for album art, we don't support them then!")
		return
	}
	albumArtURI = f.Name()
	f.Close()
}

// Song represents a music file with metadata.
type Song struct {
	File
	ID int // The song's ID (within the playlist)

	albumArt bool // Whether the song has an album art. The album art will be loaded into memory at AlbumArtURI.
}

// SongFromAttrs returns a song from the attributes map.
func (c *Client) SongFromAttrs(attr mpd.Attrs) (s Song, err error) {
	if s.ID, err = strconv.Atoi(attr["Id"]); err != nil {
		s.ID = -1
		return s, nil
	}
	if s.File, err = c.FileFromAttrs(attr); err != nil {
		return
	}

	// Attempt to load the album art.
	albumArtLock.Lock()
	defer albumArtLock.Unlock()

	// Write the album art to it
	art, err := c.AlbumArt(s.Path())
	if err != nil {
		log.Println(err)
		return s, nil
	}
	if err := ioutil.WriteFile(albumArtURI, art, 0x644); err != nil {
		log.Println(err)
		return s, nil
	}
	s.albumArt = true

	return
}

// AlbumArtURI returns the URI to the album art, if it is available.
func (s Song) AlbumArtURI() (string, bool) {
	if !s.albumArt {
		return "", false
	}
	// Should I do something better here?
	return "file://" + albumArtURI, true
}
