package mpris

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

// Transform any error into a *dbus.Error.
func (ins *Instance) transformErr(err error) *dbus.Error {
	if err == nil {
		return nil
	}
	// We have to blindly test the mpd connection here. Not a good choice, but meh...
	if err := ins.mpd.Ping(); err != nil {
		panic(fmt.Sprint("connection to mpd is severed: ", err))
	}
	var dbusErr dbus.Error
	if !errors.As(err, &dbusErr) {
		return dbus.MakeFailedError(errors.WithStack(err))
	}
	return &dbusErr
}
