package mpd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/pkg/errors"
)

var (
	mpdTemp string // Temp folder location
)

func init() {
	tmp := filepath.Join(os.TempDir(), "mpd_mpris")
	if _, err := os.Stat(tmp); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("Cannot stat temp folder, not supporting album arts...", err)
	} else if err == nil {
		log.Println("Cleaning previously existed temp folder")
		if err := os.RemoveAll(tmp); err != nil {
			log.Println("Cannot clean old temp folder, not supporting album arts...", err)
			return
		}
	}
	if err := os.MkdirAll(tmp, 0777); err != nil {
		log.Println("Cannot create temp folder, not supporting album arts...:", err)
		return
	}
	mpdTemp = tmp
}

// getAlbumArtPath stats and then load the album art if they exists.
func getAlbumArtPath(id int) (path string, alreadyExists bool) {
	if mpdTemp == "" {
		return
	}
	path = filepath.Join(mpdTemp, fmt.Sprintf("albumart_%d", id))
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, false
	} else if err != nil {
		log.Println("Cannot stat album art file, skipping: ", path, err)
		return "", false
	}
	return path, true
}

// Song represents a music file with metadata.
type Song struct {
	File
	ID int // The song's ID (within the playlist)

	albumArt string // The path to the song's album art. Empty if there is none.
}

// SameAs checks if both songs are the same.
func (s *Song) SameAs(other *Song) bool {
	if other == nil || s == nil {
		return s == nil && other == nil
	}
	return s.ID == other.ID && s.Path() == other.Path()
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

	albumArtURI, exists := getAlbumArtPath(s.ID)
	if albumArtURI != "" {
		if !exists {
			// Write the album art to it
			art, err := c.getAlbumArt(s.Path())
			if err != nil {
				log.Printf("error getting artwork for '%v': %v", s.Filepath, err)
				goto doneAlbumArt
			}
			if err := ioutil.WriteFile(albumArtURI, art, 0x644); err != nil {
				log.Printf("error getting artwork for '%v': %v", s.Filepath, err)
				goto doneAlbumArt
			}
		}
		s.albumArt = albumArtURI
	}
doneAlbumArt:

	return
}

// Get a song's album art, first by trying readpicture, then try albumart.
func (c *Client) getAlbumArt(uri string) ([]byte, error) {
	if art, err := c.readPicture(uri); err == nil {
		return art, nil
	}
	return c.AlbumArt(uri)
}

// readPicture retrieves an album artwork image for a song with the given URI using MPD's readpicture command.
// Pretty much the same as `c.AlbumArt`.
func (c *Client) readPicture(uri string) ([]byte, error) {
	offset := 0
	var data []byte
	for {
		// Read the data in chunks
		chunk, size, err := c.Command("readpicture %s %d", uri, offset).Binary()
		if err != nil {
			return nil, err
		}

		// Accumulate the data
		data = append(data, chunk...)
		offset = len(data)
		if offset >= size {
			break
		}
	}
	return data, nil
}

// AlbumArtURI returns the URI to the album art, if it is available.
func (s Song) AlbumArtURI() (string, bool) {
	if s.albumArt == "" {
		return "", false
	}
	// Should I do something better here?
	return "file://" + s.albumArt, true
}
