{ pkgs, config, lib, ... }:
with lib; let
  cfg = config.services.mpd-mpris;

  instanceOpts = {
    host = mkOption {
      type = types.str;
      description = "The host to connect to. Defaults to localhost";
      default = "";
    };
    network = mkOption {
      type = types.str;
      description = "The network type to dial (see https://golang.org/pkg/net/#Dial). Defaults to tcp, but automatically detect unix socket if localhost is used";
      default = "";
    };
    port = mkOption {
      type = types.nullOr types.numbers.positive;
      description = "The port to connect to. Only work if network is 'tcp'. Defaults to 6600";
      default = null;
    };
    passwordFile = mkOption {
      type = types.str;
      description = "Path to the file containing the password for the MPD server.";
      default = "";
    };
  };
in
{
  options.services.mpd-mpris = {
    enable = mkEnableOption "Enable mpd-mpris, an MPRIS protocol implement server for the MPD music player";
    package = mkPackageOption pkgs "mpd-mpris" { };
    enableDefaultInstance = mkOption {
      type = types.bool;
      description = "Enable the default instance";
      # default = cfg.enable && (builtins.length cfg.instances == 0);
      default = true;
    };
    instances = mkOption {
      type = types.attrsOf (types.submodule { options = instanceOpts; });
      description = "Configure multiple instances of mpd-mpris, by name";
      default = { };
    };
  } // instanceOpts;

  config = mkIf cfg.enable
    (
      let
        defaultInst = "default";
        portAssert = name: opts: {
          assertion = opts.port == null || opts.network == "" || opts.network == "tcp";
          message = "in instance ${name}: port can only be specified when network is 'tcp' (specified: '${opts.network}')";
        };
        mkService = name: opts: {
          "mpd-mpris${if name == defaultInst then "" else "-${name}"}" = {
            Install = { WantedBy = [ "default.target" ]; };

            Unit = {
              Description = "An MPRIS protocol implementation for the MPD music player (${name} instance).";
              After = lists.optional (opts.host == "") "mpd.service";
            };

            Service = {
              Type = "dbus";
              Restart = "on-failure";
              RestartSec = "5s";
              ExecStart = strings.concatStringsSep " "
                [
                  "${cfg.package}/bin/mpd-mpris"
                  (strings.optionalString (opts.host != "") "-host ${opts.host}")
                  (strings.optionalString (opts.network != "") "-network ${opts.network}")
                  (strings.optionalString (opts.passwordFile != "") "-pwd-file ${opts.network}")
                  (strings.optionalString (opts.port != null) "-port ${opts.port}")
                  (if name == defaultInst then "-no-instance" else "-instance-name ${name}")
                ];
              BusName = "org.mpris.MediaPlayer2.mpd" + (if name == defaultInst then "" else ".${name}");
            };
          };
        };
      in

      {
        assertions = [ (portAssert defaultInst cfg) ] ++ builtins.attrValues (builtins.mapAttrs portAssert cfg.instances);

        systemd.user.services = mkMerge
          (
            lists.optional cfg.enableDefaultInstance (mkService defaultInst cfg)
            ++ builtins.attrValues (builtins.mapAttrs mkService cfg.instances)
          );
      }
    );
}

