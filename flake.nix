{
  description = "lazybox";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    gomod2nix,
    gitignore,
  }: let
    allSystems = [
      "x86_64-linux" # 64-bit Intel/AMD Linux
      "aarch64-linux" # 64-bit ARM Linux
      "x86_64-darwin" # 64-bit Intel macOS
      "aarch64-darwin" # 64-bit ARM macOS
    ];
    forAllSystems = f:
      nixpkgs.lib.genAttrs allSystems (system:
        f {
          inherit system;
          pkgs = import nixpkgs {inherit system;};
        });
  in {
    packages = forAllSystems ({
      system,
      pkgs,
      ...
    }: let
      buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
    in rec {
      default = lazybox;

      lazybox = buildGoApplication {
        name = "lazybox";
        src = gitignore.lib.gitignoreSource ./.;
        # Update to latest Go version when https://nixpk.gs/pr-tracker.html?pr=324123 is backported to release-24.05.
        go = pkgs.go;
        # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
        pwd = ./.;
        subPackages = ["."];
        CGO_ENABLED = 0;
        flags = [
          "-trimpath"
        ];
        ldflags = [
          "-s"
          "-w"
          "-extldflags -static"
        ];
      };
    });

    # `nix develop` provides a shell containing development tools.
    devShell = forAllSystems ({
      system,
      pkgs,
    }:
      pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_23
          gomod2nix.legacyPackages.${system}.gomod2nix
        ];
      });

    # This flake outputs an overlay that can be used to add lazybox and
    # lazybox-docs to nixpkgs as per https://lazybox.guide/quick-start/installation/#nix
    #
    # Example usage:
    #
    # nixpkgs.overlays = [
    #   inputs.lazybox.overlays.default
    # ];
    overlays.default = final: prev: {
      lazybox = self.packages.${final.stdenv.system}.lazybox;
      lazybox-docs = self.packages.${final.stdenv.system}.lazybox-docs;
    };
  };
}
