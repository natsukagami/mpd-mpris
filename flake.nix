{
  inputs.nixpkgs.url = github:nixOS/nixpkgs;
  inputs.flake-utils.url = github:numtide/flake-utils;

  outputs = { nixpkgs, self, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        rec {
          packages.default = pkgs.callPackage ./nix/package.nix { };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls ];
          };
        }) //
    {
      overlays.default = final: prev: {
        mpd-mpris = final.pkgs.callPackage ./nix/package.nix { };
      };
      nixosModules.default = import ./nix/module.nix;
      homeManagerModules.default = import ./nix/module.nix;
    };
}
