{ pkgs, buildGoModule }:
buildGoModule {
  name = "mpd-mpris";
  src = ./..;
  vendorSha256 = "sha256-HCDJrp9WFB1z+FnYpOI5e/AojtdnpN2ZNtgGVaH/v/Q=";
}
