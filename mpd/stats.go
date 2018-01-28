package mpd

import (
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
	"github.com/pkg/errors"
)

// Stats stores statistics of the mpd instance.
type Stats struct {
	Uptime   time.Duration
	PlayTime time.Duration
	Artists  int
	Albums   int
	Songs    int
	Attrs    mpd.Attrs
}

// StatsFromAttrs returns a Stats struct from the given attrs.
func StatsFromAttrs(attr mpd.Attrs) (s Stats, err error) {
	var x int
	if x, err = strconv.Atoi(attr["uptime"]); err != nil {
		return s, errors.Wrap(err, "Parse uptime")
	}
	s.Uptime = time.Duration(x) * time.Second
	if x, err = strconv.Atoi(attr["playtime"]); err != nil {
		return s, errors.Wrap(err, "Parse playtime")
	}
	s.PlayTime = time.Duration(x) * time.Second
	if x, err = strconv.Atoi(attr["artists"]); err != nil {
		return s, errors.Wrap(err, "Parse artists")
	}
	s.Artists = x
	if x, err = strconv.Atoi(attr["albums"]); err != nil {
		return s, errors.Wrap(err, "Parse albums")
	}
	s.Albums = x
	if x, err = strconv.Atoi(attr["songs"]); err != nil {
		return s, errors.Wrap(err, "Parse songs")
	}
	s.Songs = x
	s.Attrs = attr
	return
}
