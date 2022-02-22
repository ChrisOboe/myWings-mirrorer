{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in
      rec {
        packages = flake-utils.lib.flattenTree {
          myWings-mirrorer = let lib = pkgs.lib; in
            pkgs.buildGoModule {
              pname = "myWings-mirrorer";
              version = "0.0.1";

              modSha256 = lib.fakeSha256;
              vendorSha256 = "8RE26hK45rT/PWo4+pAmxm/MQt2n1N1Jh0JpPu2EbDE=";
              src = ./src;

              meta = {
                description = "A tool to automatically download all the files of myWings your user has permissions to.";
                homepage = "https://github.com/ChrisOboe/myWings-mirrorer";
                license = lib.licenses.gpl3;
                maintainers = [ "chris@oboe.email" ];
                platforms = lib.platforms.linux ++ lib.platforms.darwin;
              };
            };
        };

        defaultPackage = packages.myWings-mirrorer;
        defaultApp = packages.myWings-mirrorer;

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
