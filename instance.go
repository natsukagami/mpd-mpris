package mpris

import (
	"fmt"
	"os"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/prop"
	"github.com/natsukagami/mpd-mpris/mpd"
	"github.com/pkg/errors"
)

// Instance is an instance of mpd-mpris.
// It contains a connection to the MPD server and the DBus connection.
type Instance struct {
	mpd   *mpd.Client
	dbus  *dbus.Conn
	props *prop.Properties
}

// Close ends the connection.
func (ins *Instance) Close() {
	ins.dbus.Close()
}

// NewInstance creates a new instance that takes care of the specified mpd.
func NewInstance(mpd *mpd.Client) (ins *Instance, err error) {
	// TODO: Unimplemented!
	ins = &Instance{mpd: mpd}
	if ins.dbus, err = dbus.SessionBus(); err != nil {
		return nil, errors.WithStack(err)
	}

	mp2 := &MediaPlayer2{Instance: ins}
	ins.dbus.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")

	player := &Player{Instance: ins}
	ins.dbus.Export(player, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	ins.props = prop.New(ins.dbus, "/org/mpris/MediaPlayer2", map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2":        mp2.properties(),
		"org.mpris.MediaPlayer2.Player": player.properties(),
	})

	reply, err := ins.dbus.RequestName(fmt.Sprintf("org.mpris.MediaPlayer2.mpd.instance%d", os.Getpid()), dbus.NameFlagReplaceExisting)

	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.WithStack(err)
	}

	return
}
