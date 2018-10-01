package mpris

// Option represents a togglable option.
type Option func(*Instance)

// NoInstance registers the instance's name without the instance# part.
func NoInstance() Option {
	return func(ins *Instance) {
		ins.name = "org.mpris.MediaPlayer2.mpd"
	}
}
