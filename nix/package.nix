{ pkgs, buildGoModule }:
buildGoModule {
  name = "mpd-mpris";
  src = ./..;
  vendorSha256 = "sha256-HCDJrp9WFB1z+FnYpOI5e/AojtdnpN2ZNtgGVaH/v/Q=";

  postInstall = ''
    mkdir -p $out/lib/systemd/user
    substitute mpd-mpris.service $out/lib/systemd/user/mpd-mpris.service \
       --replace "/usr/bin/mpd-mpris" "$out/bin/mpd-mpris"

    mkdir -p $out/etc/xdg/autostart
    substitute mpd-mpris.desktop $out/etc/xdg/autostart/mpd-mpris.desktop
  '';
}
