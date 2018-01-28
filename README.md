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

```bash
# mpd-mpris
Usage of mpd-mpris:
  -host string
        The MPD host. (default "localhost")
  -port int
        The MPD port (default 6600)
  -pwd string
        The MPD connection password. Leave empty for none.
```

Will block for requests and log them down so you may want
to run and forget.

## Implementation Status

* [x] Root Running
* [x] Player control
* [ ] Track list
* [ ] Playlist support

## License

MIT
