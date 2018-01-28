package mpd

import (
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
	"github.com/pkg/errors"
)

// Status represents mpd's current status.
type Status struct {
	Volume         int
	Repeat         bool
	Random         bool
	Single         bool
	Consume        bool
	PlaylistLength int
	State          string
	Song           int
	Seek           time.Duration
	NextSong       int
	Attrs          mpd.Attrs
}

func parseBoolFrom01(str string) (bool, error) {
	switch str {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, errors.New("Invalid value `" + str + "`, expected `0`/`1`")
	}
}

// StatusFromAttrs returns a Status struct from the given attrs.
func StatusFromAttrs(attr mpd.Attrs) (s Status, err error) {
	if s.Volume, err = strconv.Atoi(attr["volume"]); err != nil {
		return s, errors.WithStack(err)
	}
	if s.Repeat, err = parseBoolFrom01(attr["repeat"]); err != nil {
		return s, err
	}
	if s.Random, err = parseBoolFrom01(attr["random"]); err != nil {
		return s, err
	}
	if s.Single, err = parseBoolFrom01(attr["single"]); err != nil {
		return s, err
	}
	if s.Consume, err = parseBoolFrom01(attr["consume"]); err != nil {
		return s, err
	}
	if x, err := strconv.ParseFloat(attr["elapsed"], 64); err != nil {
		s.Seek = 0
	} else {
		s.Seek = time.Duration(x * float64(time.Second))
	}
	if s.PlaylistLength, err = strconv.Atoi(attr["playlistlength"]); err != nil {
		return s, errors.WithStack(err)
	}
	s.State = attr["state"]
	if s.Song, err = strconv.Atoi(attr["songid"]); err != nil {
		s.Song = -1
	}
	if s.NextSong, err = strconv.Atoi(attr["nextsongid"]); err != nil {
		s.NextSong = -1
	}
	s.Attrs = attr
	return
}
