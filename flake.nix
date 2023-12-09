{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devenv.url = "github:cachix/devenv";
    dagger.url = "github:dagger/nix";
    dagger.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.devenv.flakeModule
      ];

      systems = [ "x86_64-linux" "aarch64-darwin" ];

      perSystem = { config, self', inputs', pkgs, system, ... }: rec {
        devenv.shells = {
          default = {
            languages = {
              go.enable = true;
            };

            packages = with pkgs; [
              just
              mage
              golangci-lint
              (bats.withLibraries (p: [ p.bats-support p.bats-assert p.bats-file ]))

              skopeo
              regctl
            ] ++ [
              self'.packages.service-locator-gen
              inputs'.dagger.packages.dagger
            ];

            env = {
              DAGGER_MODULE = "ci";
            };

            # https://github.com/cachix/devenv/issues/528#issuecomment-1556108767
            containers = pkgs.lib.mkForce { };
          };

          ci = devenv.shells.default;
        };

        packages = {
          # TODO: binary name
          service-locator-gen = pkgs.buildGoModule rec {
            pname = "service-locator-gen";
            name = "service-locator-gen";
            # version = "0.8.0";

            src = pkgs.fetchFromGitHub {
              owner = "sagikazarmark";
              repo = "go-service-locator";
              # rev = "v${version}";
              rev = "f6a1274c757172035c57be4dd078cd2cc7ec190c";
              sha256 = "sha256-mmlHm1zJRSpjotoy1vSG/c56fTH5WYYUjM1NKPnk99c=";
            };

            vendorHash = "sha256-/+VGWI73NEyZgKSxe6MP4alO/J58eTwl8HrTLzGFueo=";

            subPackages = [ "." ];
          };
        };
      };
    };
}
