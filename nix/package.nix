{ buildGoModule, lib }:
buildGoModule (attrs: {
  name = "mpd-mpris";
  src = ./..;
  vendorHash = "sha256-V0ogN36WgtWaEBo7DDmXusNuLBYmkB3CRHIqxQa2gDU=";
  # vendorHash = lib.fakeHash;

  postInstall = ''
    mkdir -p $out/lib/systemd/user
    substitute services/mpd-mpris.service $out/lib/systemd/user/mpd-mpris.service \
       --replace-fail "ExecStart=mpd-mpris" "ExecStart=$out/bin/mpd-mpris"

    mkdir -p $out/etc/xdg/autostart
    substitute mpd-mpris.desktop $out/etc/xdg/autostart/mpd-mpris.desktop
  '';
})
