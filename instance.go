package mpris

import (
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

	name string
}

// Close ends the connection.
func (ins *Instance) Close() {
	ins.dbus.Close()
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

	mp2 := &MediaPlayer2{Instance: ins}
	ins.dbus.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")

	player := &Player{Instance: ins}
	ins.dbus.Export(player, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	ins.dbus.Export(introspect.NewIntrospectable(ins.IntrospectNode()), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")

	ins.props = prop.New(ins.dbus, "/org/mpris/MediaPlayer2", map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2":        mp2.properties(),
		"org.mpris.MediaPlayer2.Player": player.properties(),
	})

	reply, err := ins.dbus.RequestName(ins.Name(), dbus.NameFlagReplaceExisting)

	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.WithStack(err)
	}
	return
}
