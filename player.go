package mpris

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/natsukagami/mpd-mpris/mpd"
	"github.com/pkg/errors"
)

// This file implements a struct that satisfies the `org.mpris.MediaPlayer2.Player` interface.

// Player is a DBus object satisfying the `org.mpris.MediaPlayer2.Player` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
type Player struct {
	*Instance

	status Status
	props  map[string]*prop.Prop
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

func PlaybackStatusFromMPD(status string) (PlaybackStatus, error) {
	switch status {
	case "play":
		return PlaybackStatusPlaying, nil
	case "pause":
		return PlaybackStatusPaused, nil
	case "stop":
		return PlaybackStatusStopped, nil
	}
	return "", errors.Errorf("unknown playback status: %s", status)
}

// LoopStatus is a repeat / loop status.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Loop_Status
type LoopStatus = string

// Defined LoopStatuses
const (
	LoopStatusNone     LoopStatus = "None"
	LoopStatusTrack    LoopStatus = "Track"
	LoopStatusPlaylist LoopStatus = "Playlist"
)

// Status holds the internal status, as well as the corresponding props of the player's status.
type Status struct {
	mu sync.Mutex
	// Internal Status part
	PlaybackStatus PlaybackStatus
	LoopStatus     LoopStatus
	Shuffle        bool
	Volume         float64
	CurrentSong    mpd.Song
}

// ============================================================================

func (p *Player) setProp(iface, name string, value dbus.Variant) {
	if err := p.Instance.props.Set(iface, name, value); err != nil {
		log.Printf("Setting %s %s failed: %+v\n", iface, name, errors.WithStack(err))
	}
}

// Update performs an update on the status.
func (s *Status) Update(p *Player) *dbus.Error {
	s.mu.Lock()
	defer s.mu.Unlock()

	status, err := p.mpd.Status()
	if err != nil {
		return p.transformErr(err)
	}

	// Playback Status
	playbackStatus, err := PlaybackStatusFromMPD(status.State)
	if err != nil {
		return p.transformErr(err)
	}
	if s.PlaybackStatus != playbackStatus {
		s.PlaybackStatus = playbackStatus
		go p.setProp("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant(playbackStatus))
	}
	// Loop status
	var loopStatus LoopStatus
	switch {
	case !status.Repeat:
		loopStatus = LoopStatusNone
	case !status.Single:
		loopStatus = LoopStatusPlaylist
	default:
		loopStatus = LoopStatusTrack
	}
	if loopStatus != s.LoopStatus {
		s.LoopStatus = loopStatus
		go p.setProp("org.mpris.MediaPlayer2.Player", "LoopStatus", dbus.MakeVariant(string(loopStatus)))
	}

	// Shuffle
	if status.Random != s.Shuffle {
		s.Shuffle = status.Random
		go p.setProp("org.mpris.MediaPlayer2.Player", "Shuffle", dbus.MakeVariant(status.Random))
	}

	// Current song metadata
	song, err := p.mpd.CurrentSong()
	if err != nil {
		return p.transformErr(err)
	}
	if !song.SameAs(&s.CurrentSong) {
		go p.setProp("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(MapFromSong(song)))
	}

	// Volume
	newVolume := math.Max(0, float64(status.Volume)/100.0)
	if newVolume != s.Volume {
		s.Volume = newVolume
		go p.setProp("org.mpris.MediaPlayer2.Player", "Volume", dbus.MakeVariant(newVolume))
	}

	go p.setProp("org.mpris.MediaPlayer2.Player", "Position", dbus.MakeVariant(UsFromDuration(status.Seek)))
	return nil
}

func notImplemented(c *prop.Change) *dbus.Error {
	return dbus.MakeFailedError(errors.New("Not implemented"))
}

// OnLoopStatus handles LoopStatus change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:LoopStatus
func (p *Player) OnLoopStatus(c *prop.Change) *dbus.Error {
	loop := LoopStatus(c.Value.(string))
	log.Printf("LoopStatus changed to %v\n", loop)
	p.status.mu.Lock()
	defer p.status.mu.Unlock()
	p.status.LoopStatus = loop
	switch loop {
	case LoopStatusNone:
		if err := p.mpd.Single(false); err != nil {
			return p.transformErr(err)
		}
		if err := p.mpd.Repeat(false); err != nil {
			return p.transformErr(err)
		}
	case LoopStatusPlaylist:
		if err := p.mpd.Single(false); err != nil {
			return p.transformErr(err)
		}
		if err := p.mpd.Repeat(true); err != nil {
			return p.transformErr(err)
		}
	case LoopStatusTrack:
		if err := p.mpd.Single(true); err != nil {
			return p.transformErr(err)
		}
		if err := p.mpd.Repeat(true); err != nil {
			return p.transformErr(err)
		}
	default:
		return p.transformErr(errors.New("Invalid loop " + string(loop)))
	}
	return nil
}

// OnVolume handles volume changes.
func (p *Player) OnVolume(c *prop.Change) *dbus.Error {
	val := int(c.Value.(float64) * 100)
	log.Printf("Volume changed to %v\n", val)
	p.status.mu.Lock()
	defer p.status.mu.Unlock()
	p.status.Volume = c.Value.(float64)
	if val < 0 {
		val = 0
	}
	return p.transformErr(p.mpd.SetVolume(val))
}

// OnShuffle handles Shuffle change.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Property:Shuffle
func (p *Player) OnShuffle(c *prop.Change) *dbus.Error {
	log.Printf("Shuffle changed to %v\n", c.Value.(bool))
	p.status.mu.Lock()
	defer p.status.mu.Unlock()
	p.status.Shuffle = c.Value.(bool)
	return p.transformErr(p.mpd.Random(c.Value.(bool)))
}

func (p *Player) createStatus() {
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

	volume := math.Max(0, float64(status.Volume)/100.0)

	p.status = Status{
		PlaybackStatus: playStatus,
		LoopStatus:     loopStatus,
		Shuffle:        status.Random,
		Volume:         volume,
		CurrentSong:    song,
	}

	// Set up a position updater
	go func() {
		tick := time.NewTicker(time.Second / 10)
		defer tick.Stop()
		for range tick.C {
			if err := p.status.Update(p); err != nil {
				log.Printf("%+v\n", err)
			}
		}
	}()

	p.props = map[string]*prop.Prop{
		"PlaybackStatus": newProp(playStatus, nil),
		"LoopStatus":     newProp(loopStatus, p.OnLoopStatus),
		"Rate":           newProp(1.0, notImplemented),
		"Shuffle":        newProp(status.Random, p.OnShuffle),
		"Metadata":       newProp(MapFromSong(song), nil),
		"Volume":         newProp(volume, p.OnVolume),
		"Position": {
			Value:    UsFromDuration(status.Seek),
			Writable: true,
			Emit:     prop.EmitFalse,
			Callback: nil,
		},
		"MinimumRate":   newProp(1.0, nil),
		"MaximumRate":   newProp(1.0, nil),
		"CanGoNext":     newProp(true, nil),
		"CanGoPrevious": newProp(true, nil),
		"CanPlay":       newProp(true, nil),
		"CanPause":      newProp(true, nil),
		"CanSeek":       newProp(status.Seekable, nil),
		"CanControl":    newProp(true, nil),
	}
}

// ============================================================================

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	log.Printf("Next requested\n")
	if err := p.transformErr(p.Instance.mpd.Next()); err != nil {
		return err
	}
	return p.status.Update(p)
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	log.Printf("Previous requested\n")
	if err := p.transformErr(p.Instance.mpd.Previous()); err != nil {
		return err
	}
	return p.status.Update(p)
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	log.Printf("Pause requested\n")
	if err := p.transformErr(p.Instance.mpd.Pause(true)); err != nil {
		return err
	}
	return p.status.Update(p)
}

// Play starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	log.Printf("Play requested\n")
	if err := p.transformErr(p.Instance.mpd.Play(-1)); err != nil {
		return err
	}
	return p.status.Update(p)
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	log.Printf("Stop requested\n")
	if err := p.transformErr(p.Instance.mpd.Stop()); err != nil {
		return err
	}
	return p.status.Update(p)
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	log.Printf("Play/Pause requested. Switching context...\n")
	status, err := p.mpd.Status()
	if err != nil {
		return p.transformErr(err)
	}
	if status.State == "play" {
		return p.Pause()
	}
	return p.Play()
}

// Seek seeks forward in the current track by the specified number of microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Seek
func (p *Player) Seek(x TimeInUs) *dbus.Error {
	status, err := p.mpd.Status()
	if err != nil {
		return p.transformErr(err)
	}

	if !status.Seekable {
		return nil // Quit silently
	}

	log.Printf("Seek(%v) requested\n", x.Duration())
	song, err := p.mpd.CurrentSong()
	if err != nil {
		return p.transformErr(err)
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
	status, err := p.mpd.Status()
	if err != nil {
		return p.transformErr(err)
	}

	if !status.Seekable {
		return nil // Quit silently
	}

	log.Printf("SetPosition(%v, %v) requested\n", o, x.Duration())
	var id int
	if _, err := fmt.Sscanf(string(o), TrackIDFormat, &id); err != nil {
		return p.transformErr(err)
	}
	if err := p.mpd.SeekID(id, int(x.Duration()/time.Second)); err != nil {
		return p.transformErr(err)
	}
	if err := p.status.Update(p); err != nil {
		return err
	}
	// Unnatural seek, create signal
	return p.transformErr(p.dbus.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player.Seeked", x))
}
