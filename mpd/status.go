package mpd

import (
	"time"

	"github.com/fhs/gompd/mpd"
)

// Status represents mpd's current status.
type Status struct {
	Volume         int
	Repeat         bool
	Random         bool
	Single         bool
	Consume        bool
	PlaylistLength time.Duration
	State          string
	Song           int
	Seek           time.Duration
	NextSong       int
	Attrs          mpd.Attrs

	Seekable bool // Whether we can seek the current song
}

// StatusFromAttrs returns a Status struct from the given attrs.
func StatusFromAttrs(attr mpd.Attrs) (s Status, err error) {
	p := &parseMap{m: attr}

	p.Int("volume", &s.Volume, true)
	p.Bool("repeat", &s.Repeat, true)
	p.Bool("single", &s.Single, true)
	p.Bool("random", &s.Random, true)
	p.Bool("consume", &s.Consume, true)

	{
		var x float64
		p.Float("elapsed", &x, true)
		s.Seek = time.Duration(x * float64(time.Second))

		//? This is a guess, assuming any non-seekable content has 0 duration
		s.Seekable = p.Float("duration", &x, true)
		s.Seekable = s.Seekable && x != 0.0
	}

	{
		var x int
		p.Int("playlistlength", &x, true)
		s.PlaylistLength = time.Duration(x) * time.Second
	}

	p.String("state", &s.State, true)
	if !p.Int("songid", &s.Song, true) {
		s.Song = -1
	}
	if !p.Int("nextsongid", &s.NextSong, true) {
		s.NextSong = -1
	}

	err = p.Err
	s.Attrs = attr
	return s, nil
}
