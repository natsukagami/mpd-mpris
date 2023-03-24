# mpd-mpris

An implementation of the [MPRIS](https://specifications.freedesktop.org/mpris-spec/latest/) protocol for [MPD](http://musicpd.org/).

---

## Installation

Probably needs Go v1.9 or newer.

```bash
go install github.com/natsukagami/mpd-mpris/cmd/mpd-mpris
```

Install scripts coming soon.

## Running

```
# mpd-mpris --help
Usage of mpd-mpris:
  -host string
        The MPD host (default localhost)
  -interval duration
        How often to update the current song position. Set to 0 to never update the current song position. (default 1s)
  -network string
        The network used to dial to the mpd server. Check https://golang.org/pkg/net/#Dial for available values (most common are "tcp" and "unix") (default "tcp")
  -no-instance
        Set the MPDris's interface as 'org.mpris.MediaPlayer2.mpd' instead of 'org.mpris.MediaPlayer2.mpd.instance#'
  -port int
        The MPD port. Only works if network is "tcp". If you use anything else, you should put the port inside addr yourself. (default 6600)
  -pwd string
        The MPD connection password. Leave empty for none.
```

Will block for requests and log them down so you may want
to run and forget.

## Implementation Status

- [x] Root Running
- [x] Player control
- [ ] Track list
- [ ] Playlist support

## License

MIT
