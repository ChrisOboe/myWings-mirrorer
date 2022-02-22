{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = [
            # ide
            pkgs.kakoune
            pkgs.kak-lsp
            pkgs.gopls

            # dev
            pkgs.go
            pkgs.openapi-generator-cli
          ];
        };
      });
}
