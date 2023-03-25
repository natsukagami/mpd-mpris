{ pkgs, buildGoModule }:
buildGoModule {
  name = "mpd-mpris";
  src = ./..;
  vendorSha256 = "sha256-GmdD/4VYp3KeblNGgltFWHdOnK5qsBa2ygIYOBrH+b0=";
}
