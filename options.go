package mpris

import "fmt"

// Option represents a togglable option.
type Option func(*Instance)

// NoInstance registers the instance's name without the instance# part.
func NoInstance() Option {
	return func(ins *Instance) {
		ins.name = "org.mpris.MediaPlayer2.mpd"
	}
}

// InstanceName gives a custom name after "mpd" for the MPRIS instance.
func InstanceName(name string) Option {
	return func(ins *Instance) {
		ins.name = fmt.Sprintf("org.mpris.MediaPlayer2.mpd.%s", name)
	}
}
