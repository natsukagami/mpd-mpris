package mpris

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/prop"
	"github.com/pkg/errors"
)

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.Player` interface.

// Player is a DBus object satisfying the `org.mpris.MediaPlayer2.Player` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
type Player struct {
	*Instance
}

// TrackID is the Unique track identifier.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Track_Id
type TrackID string

// PlaybackRate is a playback rate.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Playback_Rate
type PlaybackRate float64

// TimeInUs is time in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Time_In_Us
type TimeInUs int64

// UsFromDuration returns the type from a time.Duration
func UsFromDuration(t time.Duration) TimeInUs {
	return TimeInUs(t / time.Microsecond)
}

// Duration returns the type in time.Duration
func (t TimeInUs) Duration() time.Duration { return time.Duration(t) * time.Microsecond }

// PlaybackStatus is a playback state.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
type PlaybackStatus string

// Defined PlaybackStatuses.
const (
	PlaybackStatusPlaying PlaybackStatus = "Playing"
	PlaybackStatusPaused  PlaybackStatus = "Paused"
	PlaybackStatusStopped PlaybackStatus = "Stopped"
)

// LoopStatus is a repeat / loop status.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Loop_Status
type LoopStatus = string

// Defined LoopStatuses
const (
	LoopStatusNone     LoopStatus = "None"
	LoopStatusTrack    LoopStatus = "Track"
	LoopStatusPlaylist LoopStatus = "Playlist"
)

// ============================================================================

func transform(err error) *dbus.Error {
	if err != nil {
		return dbus.MakeFailedError(errors.WithStack(err))
	}
	return nil
}

func (p *Player) updateStatus() *dbus.Error {
	status, err := p.mpd.Status()
	if err != nil {
		return transform(err)
	}
	var playStatus PlaybackStatus
	switch status.State {
	case "play":
		playStatus = PlaybackStatusPlaying
	case "pause":
		playStatus = PlaybackStatusPaused
	case "stop":
		playStatus = PlaybackStatusStopped
	}
	var loopStatus LoopStatus
	switch {
	case !status.Repeat:
		loopStatus = LoopStatusNone
	case !status.Single:
		loopStatus = LoopStatusPlaylist
	default:
		loopStatus = LoopStatusTrack
	}
	song, err := p.mpd.CurrentSong()
	if err != nil {
		return transform(err)
	}
	if err := p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(MapFromSong(song))); err != nil {
		return err
	}
	if err := p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant(playStatus)); err != nil {
		return err
	}
	if oldLoop, err := p.props.Get("org.mpris.MediaPlayer2.Player", "LoopStatus"); err == nil && oldLoop.Value().(string) != string(loopStatus) {
		if err := p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "LoopStatus", dbus.MakeVariant(string(loopStatus))); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if oldShuffle, err := p.props.Get("org.mpris.MediaPlayer2.Player", "Shuffle"); err == nil && oldShuffle.Value().(bool) != status.Random {
		if err := p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "Shuffle", dbus.MakeVariant(status.Random)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	newVolume := math.Max(0, float64(status.Volume)/100.0)
	if oldVolume, err := p.props.Get("org.mpris.MediaPlayer2.Player", "Volume"); err == nil && oldVolume.Value().(float64) != newVolume {

		if err := p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "Volume", dbus.MakeVariant(newVolume)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return p.Instance.props.Set("org.mpris.MediaPlayer2.Player", "Position", dbus.MakeVariant(UsFromDuration(status.Seek)))
}

// OnLoopStatus handles LoopStatus change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:LoopStatus
func (p *Player) OnLoopStatus(c *prop.Change) *dbus.Error {
	loop := LoopStatus(c.Value.(string))
	log.Printf("LoopStatus changed to %v\n", loop)
	switch loop {
	case LoopStatusNone:
		if err := p.mpd.Single(false); err != nil {
			return transform(err)
		}
		if err := p.mpd.Repeat(false); err != nil {
			return transform(err)
		}
	case LoopStatusPlaylist:
		if err := p.mpd.Single(false); err != nil {
			return transform(err)
		}
		if err := p.mpd.Repeat(true); err != nil {
			return transform(err)
		}
	case LoopStatusTrack:
		if err := p.mpd.Single(true); err != nil {
			return transform(err)
		}
		if err := p.mpd.Repeat(true); err != nil {
			return transform(err)
		}
	default:
		return transform(errors.New("Invalid loop " + string(loop)))
	}
	return nil
}

// OnVolume handles volume changes.
func (p *Player) OnVolume(c *prop.Change) *dbus.Error {
	val := int(c.Value.(float64) * 100)
	log.Printf("Volume changed to %v\n", val)
	if val < 0 {
		val = 0
	}
	return transform(p.mpd.SetVolume(val))
}

// OnShuffle handles Shuffle change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:Shuffle
func (p *Player) OnShuffle(c *prop.Change) *dbus.Error {
	log.Printf("Shuffle changed to %v\n", c.Value.(bool))
	return transform(p.mpd.Random(c.Value.(bool)))
}

func (p *Player) properties() map[string]*prop.Prop {
	status, err := p.mpd.Status()
	if err != nil {
		log.Fatalf("%+v", err)
		panic(err)
	}
	var playStatus PlaybackStatus
	switch status.State {
	case "play":
		playStatus = PlaybackStatusPlaying
	case "pause":
		playStatus = PlaybackStatusPaused
	default:
		playStatus = PlaybackStatusStopped
	}
	var loopStatus LoopStatus
	switch {
	case !status.Repeat:
		loopStatus = LoopStatusNone
	case !status.Single:
		loopStatus = LoopStatusPlaylist
	default:
		loopStatus = LoopStatusTrack
	}
	song, err := p.mpd.CurrentSong()
	if err != nil {
		panic(err)
	}

	// Set up a position updater
	go func() {
		tick := time.NewTicker(time.Second / 10)
		defer tick.Stop()
		for range tick.C {
			if err := p.updateStatus(); err != nil {
				log.Printf("%v\n", err)
			}
		}
	}()

	return map[string]*prop.Prop{
		"PlaybackStatus": newProp(playStatus, true, true, nil),
		"LoopStatus":     newProp(loopStatus, true, true, p.OnLoopStatus),
		"Rate":           newProp(1.0, true, true, nil),
		"Shuffle":        newProp(status.Random, true, true, p.OnShuffle),
		"Metadata":       newProp(MapFromSong(song), true, true, nil),
		"Volume":         newProp(math.Max(0, float64(status.Volume)/100.0), true, true, p.OnVolume),
		"Position": &prop.Prop{
			Value:    UsFromDuration(status.Seek),
			Writable: true,
			Emit:     prop.EmitFalse,
			Callback: nil,
		},
		"MinimumRate":   newProp(1.0, false, true, nil),
		"MaximumRate":   newProp(1.0, false, true, nil),
		"CanGoNext":     newProp(true, false, true, nil),
		"CanGoPrevious": newProp(true, false, true, nil),
		"CanPlay":       newProp(true, false, true, nil),
		"CanPause":      newProp(true, false, true, nil),
		"CanSeek":       newProp(true, false, true, nil),
		"CanControl":    newProp(true, false, true, nil),
	}
}

// ============================================================================

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	log.Printf("Next requested\n")
	if err := transform(p.Instance.mpd.Next()); err != nil {
		return err
	}
	return p.updateStatus()
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	log.Printf("Previous requested\n")
	if err := transform(p.Instance.mpd.Previous()); err != nil {
		return err
	}
	return p.updateStatus()
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	log.Printf("Pause requested\n")
	if err := transform(p.Instance.mpd.Pause(true)); err != nil {
		return err
	}
	return p.updateStatus()
}

// Play starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	log.Printf("Play requested\n")
	if err := transform(p.Instance.mpd.Play(-1)); err != nil {
		return err
	}
	return p.updateStatus()
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	log.Printf("Stop requested\n")
	if err := transform(p.Instance.mpd.Stop()); err != nil {
		return err
	}
	return p.updateStatus()
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	log.Printf("Play/Pause requested. Switching context...\n")
	status, err := p.mpd.Status()
	if err != nil {
		return transform(err)
	}
	if status.State == "play" {
		return p.Pause()
	}
	return p.Play()
}

// Seek seeks forward in the current track by the specified number of microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p *Player) Seek(x TimeInUs) *dbus.Error {
	log.Printf("Seek(%v) requested\n", x.Duration())
	status, err := p.mpd.Status()
	if err != nil {
		return transform(err)
	}
	song, err := p.mpd.CurrentSong()
	if err != nil {
		return transform(err)
	}
	if status.Seek+x.Duration() < 0 {
		return p.SetPosition(TrackID(fmt.Sprintf(TrackIDFormat, status.Song)), 0)
	}
	if status.Seek+x.Duration() > song.Duration {
		return p.Next()
	}
	return p.SetPosition(TrackID(fmt.Sprintf(TrackIDFormat, status.Song)), UsFromDuration(status.Seek+x.Duration()))
}

// SetPosition sets the current track position in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:SetPosition
func (p *Player) SetPosition(o TrackID, x TimeInUs) *dbus.Error {
	log.Printf("SetPosition(%v, %v) requested\n", o, x.Duration())
	var id int
	if _, err := fmt.Sscanf(string(o), TrackIDFormat, &id); err != nil {
		return transform(err)
	}
	if err := p.mpd.SeekID(id, int(x.Duration()/time.Second)); err != nil {
		return transform(err)
	}
	if err := p.updateStatus(); err != nil {
		return err
	}
	// Unnatural seek, create signal
	return transform(p.dbus.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player.Seeked", x))
}
