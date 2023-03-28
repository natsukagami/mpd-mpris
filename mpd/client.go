package mpd

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/pkg/errors"
)

// Client represents a MPD client.
// Some of the methods are overriden from the `mpd.Client` struct to provide typings safety.
type Client struct {
	*mpd.Client
	*Watcher
	Address        string
	MusicDirectory string

	lastSongMu sync.Mutex
	lastSong   *Song
}

// KeepaliveTimeout is the time between pings to keep the connection alive.
// As MPD recommends a 30 second timeout for connection (https://mpd.readthedocs.io/en/latest/client.html#environment-variables),
// we keep something similar here.
// However, if we detect `MPD_TIMEOUT`, we set the corresponding timeout.
const KeepaliveTimeoutDefault = 25 * time.Second

func (c *Client) init() error {
	// Find the music directory
	conf, err := c.Command("config").Attrs()
	if err == nil {
		c.MusicDirectory = conf["music_directory"]
	}
	return nil
}

// Dial connects to MPD listening on address addr (e.g. "127.0.0.1:6600") on network network (e.g. "tcp").
func Dial(network, addr string) (*Client, error) {
	return DialAuthenticated(network, addr, "")
}

// DialAuthenticated connects to MPD listening on address addr (e.g. "127.0.0.1:6600") on network network (e.g. "tcp").
// It then authenticates with MPD using the plaintext password password if it's not empty.
func DialAuthenticated(network, addr, password string) (*Client, error) {
	c, err := mpd.DialAuthenticated(network, addr, password)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	w, err := NewWatcher(network, addr, password)
	if err != nil {
		return nil, err
	}
	client := &Client{Client: c, Watcher: w, Address: addr}
	if err := client.init(); err != nil {
		return nil, err
	}
	return client, nil
}

// CurrentSong returns information about the current song in the playlist.
func (c *Client) CurrentSong() (Song, error) {
	a, e := c.Client.CurrentSong()
	if e != nil {
		return Song{}, errors.WithStack(e)
	}
	c.lastSongMu.Lock()
	defer c.lastSongMu.Unlock()
	if c.lastSong != nil && c.lastSong.Path() == a["file"] {
		// Heuristically, we have... the same song...
		return *c.lastSong, nil
	}
	song, err := c.SongFromAttrs(a)
	if err != nil {
		return Song{}, err
	}
	c.lastSong = &song
	return *c.lastSong, nil
}

// Find searches the library for songs and returns attributes for each matching song.
// The args are the raw arguments passed to MPD. For example, to search for
// songs that belong to a specific artist and album:
//
//	Find("artist", "Artist Name", "album", "Album Name")
//
// Searches are case sensitive. Use Search for case insensitive search.
func (c *Client) Find(args ...string) ([]File, error) {
	a, err := c.Client.Find(args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]File, len(a))
	for id, item := range a {
		if arr[id], err = c.FileFromAttrs(item); err != nil {
			return nil, errors.Wrapf(err, "Item %d", id)
		}
	}
	return arr, nil
}

// ListAllInfo returns attributes for songs in the library. Information about any song that is either inside or matches the passed in uri is returned.
// To get information about every song in the library, pass in "/".
func (c *Client) ListAllInfo(uri string) ([]Item, error) {
	a, err := c.Client.ListAllInfo(uri)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]Item, len(a))
	for id, item := range a {
		if arr[id], err = c.ItemFromAttrs(item); err != nil {
			return nil, errors.Wrapf(err, "Item %d", id)
		}
	}
	return arr, nil
}

// ListInfo lists the contents of the directory URI using MPD's lsinfo command.
func (c *Client) ListInfo(uri string) ([]Item, error) {
	a, err := c.Client.ListInfo(uri)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]Item, len(a))
	for id, item := range a {
		if arr[id], err = c.ItemFromAttrs(item); err != nil {
			return nil, errors.Wrapf(err, "Item %d", id)
		}
	}
	return arr, nil
}

// ListPlaylists lists all stored playlists.
func (c *Client) ListPlaylists() ([]PlaylistFile, error) {
	a, err := c.Client.ListPlaylists()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]PlaylistFile, len(a))
	for id, item := range a {
		arr[id] = PlaylistFile{Attrs: item}
	}
	return arr, nil
}

// PlaylistContents returns a list of attributes for songs in the specified stored playlist.
func (c *Client) PlaylistContents(name string) ([]File, error) {
	a, err := c.Client.PlaylistContents(name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]File, len(a))
	for id, item := range a {
		if arr[id], err = c.FileFromAttrs(item); err != nil {
			return nil, errors.Wrapf(err, "Item %d", id)
		}
	}
	return arr, nil
}

// PlaylistInfo returns attributes for songs in the current playlist.
// If both start and end are negative, it does this for all songs in playlist.
// If end is negative but start is positive, it does it for the song at position start.
// If both start and end are positive, it does it for positions in range [start, end).
func (c *Client) PlaylistInfo(start, end int) ([]File, error) {
	a, err := c.Client.PlaylistInfo(start, end)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	arr := make([]File, len(a))
	for id, item := range a {
		if arr[id], err = c.FileFromAttrs(item); err != nil {
			return nil, errors.Wrapf(err, "Item %d", id)
		}
	}
	return arr, nil
}

// Stats displays statistics (number of artists, songs, playtime, etc)
func (c *Client) Stats() (Stats, error) {
	a, e := c.Client.Stats()
	if e != nil {
		return Stats{}, errors.WithStack(e)
	}
	return StatsFromAttrs(a)
}

// Status returns information about the current status of MPD.
func (c *Client) Status() (Status, error) {
	a, e := c.Client.Status()
	if e != nil {
		return Status{}, errors.WithStack(e)
	}
	return StatusFromAttrs(a)
}

// Keepalive keeps the client alive with pings until `ctx` is done.
func (c *Client) Keepalive(ctx context.Context) {
	// get the timeout
	timeout := KeepaliveTimeoutDefault
	if tStr, ok := os.LookupEnv("MPD_TIMEOUT"); ok {
		if t, err := strconv.Atoi(tStr); err == nil {
			timeout = time.Duration(t) * time.Second
			log.Println("Using MPD_TIMEOUT's keepalive clock of %v", timeout)
		}
	}

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Ping(); err != nil {
				log.Fatalf("Connection to mpd is severed: %+v", errors.WithStack(err))
			}
		}
	}
}

// Close closes the client.
func (c *Client) Close() error {
	if err := c.Client.Close(); err != nil {
		return errors.WithStack(err)
	}
	if err := c.Watcher.Close(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
