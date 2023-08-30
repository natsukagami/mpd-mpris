# mpd-mpris

[![AUR version](https://img.shields.io/aur/version/mpd-mpris)](https://aur.archlinux.org/packages/mpd-mpris)
[![AUR version](https://img.shields.io/aur/version/mpd-mpris-bin)](https://aur.archlinux.org/packages/mpd-mpris-bin)
[![Flake is available](https://img.shields.io/badge/flake-available-blue)](#for-nix-users)
[![Matrix chat](https://img.shields.io/matrix/mpd-mpris:matrix.org)](https://matrix.to/#/#mpd-mpris:matrix.org)


An implementation of the [MPRIS](https://specifications.freedesktop.org/mpris-spec/latest/) protocol for [MPD](http://musicpd.org/).

---

## Installation

Probably needs Go v1.9 or newer.

```bash
go install github.com/natsukagami/mpd-mpris/cmd/mpd-mpris
```

### For Arch Linux users.

Check out the AUR packages [mpd-mpris](https://aur.archlinux.org/packages/mpd-mpris)
and [mpd-mpris-bin](https://aur.archlinux.org/packages/mpd-mpris-bin)
for the manually and pre-built versions respectively.
A systemd user service file is also provided (enable with `systemctl --user enable mpd-mpris --now`).

### For Nix users

The repository provides the `mpd-mpris` package, overlay and a NixOS/`home-manager` module (`services.mpd-mpris`) as a flake.
```
# nix flake show github:natsukagami/mpd-mpris
github:natsukagami/mpd-mpris
├───devShells
│   ├───aarch64-darwin
│   │   └───default: development environment 'nix-shell'
│   ├───aarch64-linux
│   │   └───default: development environment 'nix-shell'
│   ├───x86_64-darwin
│   │   └───default: development environment 'nix-shell'
│   └───x86_64-linux
│       └───default: development environment 'nix-shell'
├───homeManagerModules:
│   └───default: home-manager module
├───nixosModules
│   └───default: NixOS module
├───overlays
│   └───default: Nixpkgs overlay
└───packages
    ├───aarch64-darwin
    │   └───default: package 'mpd-mpris'
    ├───aarch64-linux
    │   └───default: package 'mpd-mpris'
    ├───x86_64-darwin
    │   └───default: package 'mpd-mpris'
    └───x86_64-linux
        └───default: package 'mpd-mpris'
```

The `mpd-mpris` module has the following options:
- `services.mpd-mpris.enable`: Enable the service.
- `services.mpd-mpris.package`: Overrides the package. Defaults to `pkgs.mpd-mpris` (which uses the nixpkgs package without the overlay).
- `services.mpd-mpris.enableDefaultInstance`: The module has a default instance that listens to the local `mpd`, enable this.

Per-instance configurations:
- `services.mpd-mpris.instances.{name}.host`: The host to connect to. (`-host` flag)
- `services.mpd-mpris.instances.{name}.network`: The network type. (`-network` flag)
- `services.mpd-mpris.instances.{name}.port`: The port to connect to. (`-port` flag)
- `services.mpd-mpris.instances.{name}.passwordFile`: The file containing the password to use. (`-pwd-file` flag)

Each instance will create a `mpd-mpris-{name}` service (with the default being `mpd-mpris`), with the MPRIS instance name
`org.mpris.MediaPlayer2.mpd.{name}` (with the default being just `org.mpris.MediaPlayer2.mpd`).
All per-instance configurations are available without the `instances.{name}` infix, and will apply to the default instance.

## Running

```
# mpd-mpris --help
Usage of mpd-mpris:
  -host string
        The MPD host (default localhost)
  -instance-name string
        Set the MPRIS's interface as 'org.mpris.MediaPlayer2.mpd.{instance-name}'
  -network string
        The network used to dial to the mpd server. Check https://golang.org/pkg/net/#Dial for available values (most common are "tcp" and "unix") (default "tcp")
  -no-instance
        Set the MPRIS's interface as 'org.mpris.MediaPlayer2.mpd' instead of 'org.mpris.MediaPlayer2.mpd.instance#'
  -port int
        The MPD port. Only works if network is "tcp". If you use anything else, you should put the port inside addr yourself. (default 6600)
  -pwd string
        The MPD connection password. Leave empty for none.
  -pwd-file string
        Path to the file containing the mpd server password.
```

Will block for requests and log them down so you may want
to run and forget.

## Questions?

Join our Matrix channel at [`#mpd-mpris:matrix.org`](https://matrix.to/#/#mpd-mpris:matrix.org).

## Implementation Status

- [x] Root Running
- [x] Player control
- [ ] Track list
- [ ] Playlist support

## License

MIT
