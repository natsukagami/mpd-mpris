package mpd

import (
	"context"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/pkg/errors"
)

// Watcher is our implementation of the watcher.
// It automatically subscribes to MPRIS-related events and
// `Poll` can be used to wait for any event.
type Watcher struct {
	*mpd.Watcher
}

var (
	// See https://mpd.readthedocs.io/en/latest/protocol.html#command-idle
	eventsToSubscribe = []string{
		"playlist", // the queue (i.e. the current playlist) has been modified
		"player",   // the player has been started, stopped or seeked or tags of the currently playing song have changed (e.g. received from stream)
		"mixer",    // the volume has been changed
		"options",  // options like repeat, random, crossfade, replay gain
	}
)

// NewWatcher creates a new watcher with the given parameters.
func NewWatcher(net, addr, passwd string) (*Watcher, error) {
	w, err := mpd.NewWatcher(net, addr, passwd, eventsToSubscribe...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Watcher{w}, nil
}

// Poll waits for the next event, or errors out.
func (w *Watcher) Poll(ctx context.Context) error {
	select {
	case <-w.Event:
		return nil
	case err := <-w.Error:
		return errors.Wrap(err, "polling for events")
	case <-ctx.Done():
		return context.Canceled
	}
}
