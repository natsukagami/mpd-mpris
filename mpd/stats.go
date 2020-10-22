package mpd

import (
	"time"

	"github.com/fhs/gompd/v2/mpd"
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
	p := &parseMap{m: attr}
	var x int

	p.Int("uptime", &x, false)
	s.Uptime = time.Duration(x) * time.Second

	p.Int("playtime", &x, false)
	s.PlayTime = time.Duration(x) * time.Second

	p.Int("artists", &s.Artists, false)
	p.Int("albums", &s.Albums, false)
	p.Int("songs", &s.Songs, false)

	err = p.Err
	s.Attrs = attr
	return
}
