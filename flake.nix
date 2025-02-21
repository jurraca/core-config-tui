{
  description = "Bitcoin Core configuration generator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.default = pkgs.buildGoModule {
        pname = "bitcoin-core-config";
        version = "0.1.0";
        src = ./.;

        vendorHash = "sha256-bTV7RQ1An26kDSTGQf1lm5Jai2yGuV6NfZMMiO/isZs=";
      };

      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
        ];
      };
    });
}
