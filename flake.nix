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
      default = pkgs.buildEnv {
        name = "xbox";
        paths = [diff2];
      };

      diff2 = buildGoApplication {
        name = "diff2";
        src = gitignore.lib.gitignoreSource ./diff2;
        go = pkgs.go_1_23;
        pwd = ./diff2;
        CGO_ENABLED = 0;
        modules = ./diff2/gomod2nix.toml;
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

    overlays.default = final: prev: {
      diff2 = self.packages.${final.stdenv.system}.diff2;
    };
  };
}
