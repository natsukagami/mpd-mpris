package mpris

import (
	"context"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5/introspect"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/natsukagami/mpd-mpris/mpd"
	"github.com/pkg/errors"
)

// Instance is an instance of mpd-mpris.
// It contains a connection to the MPD server and the DBus connection.
type Instance struct {
	mpd   *mpd.Client
	dbus  *dbus.Conn
	props *prop.Properties

	// interface implementations
	root   *MediaPlayer2
	player *Player

	name string
}

// Close ends the connection.
func (ins *Instance) Close() error {
	if ins.mpd == nil {
		return nil // already closed
	}
	if err := ins.dbus.Close(); err != nil {
		return errors.WithStack(err)
	}
	if err := ins.mpd.Close(); err != nil {
		return err
	}
	ins.mpd = nil
	return nil
}

// Name returns the name of the instance.
func (ins *Instance) Name() string {
	return ins.name
}

// NewInstance creates a new instance that takes care of the specified mpd.
func NewInstance(mpd *mpd.Client, opts ...Option) (ins *Instance, err error) {
	ins = &Instance{
		mpd: mpd,

		name: fmt.Sprintf("org.mpris.MediaPlayer2.mpd.instance%d", os.Getpid()),
	}
	if ins.dbus, err = dbus.SessionBus(); err != nil {
		return nil, errors.WithStack(err)
	}
	// Apply options
	for _, opt := range opts {
		opt(ins)
	}

	ins.root = &MediaPlayer2{Instance: ins}
	ins.player = &Player{Instance: ins}

	ins.player.createStatus()

	ins.props, err = prop.Export(ins.dbus, "/org/mpris/MediaPlayer2", map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2":        ins.root.properties(),
		"org.mpris.MediaPlayer2.Player": ins.player.props,
	})
	return
}

// Start starts the instance. Blocking, so you should fire and forget ;)
func (ins *Instance) Start(ctx context.Context) error {
	ins.dbus.Export(ins.root, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	ins.dbus.Export(ins.player, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")
	ins.dbus.Export(introspect.NewIntrospectable(ins.IntrospectNode()), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")

	reply, err := ins.dbus.RequestName(ins.Name(), dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.WithStack(err)
	}

	// Set up a periodic updaters
	go ins.mpd.Keepalive(ctx)
	go ins.player.pollSeek(ctx)

	// Set up a status updater
	for {
		if err := ins.mpd.Poll(ctx); errors.Is(err, context.Canceled) {
			return nil
		} else if err != nil {
			return errors.Wrap(err, "cannot poll mpd")
		}
		if err := ins.player.update(); err != nil {
			return err
		}
	}
}
