package mpris

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2` interface.

// MediaPlayer2 is a DBus object satisfying the `org.mpris.MediaPlayer2` interface.
type MediaPlayer2 struct {
	*Instance
}

func (m *MediaPlayer2) properties() map[string]*prop.Prop {
	return map[string]*prop.Prop{
		"CanQuit":      newProp(false, nil),                                            // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanQuit
		"CanRaise":     newProp(false, nil),                                            // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanRaise
		"HasTrackList": newProp(true, nil),                                             // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:HasTrackList
		"Identity":     newProp(fmt.Sprintf("MPD on %s", m.Instance.mpd.Address), nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:Identity
		// Empty because we can't add arbitary files in...
		"SupportedUriSchemes": newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedUriSchemes
		"SupportedMimeTypes":  newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedMimeTypes
	}
}

// Raise brings the media player's user interface to the front using any appropriate mechanism available.
// But for MPD, there's no User Interface, this function does nothing.
//
// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Raise
func (m *MediaPlayer2) Raise() *dbus.Error { return nil }

// Quit causes the media player to stop running.
// But for MPD, it's not up to the client to end its existence. Hence this function does nothing.
//
// https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Method:Quit
func (m *MediaPlayer2) Quit() *dbus.Error { return nil }
