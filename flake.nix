{
  description = "Blueprint - composable project scaffolding tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      env = {
        CGO_ENABLED = 0;
      };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          commit = self.shortRev or self.dirtyShortRev or "unknown";
          buildDate = self.lastModifiedDate or "unknown";
        in
        {
          default = pkgs.buildGoModule rec {
            pname = "blueprint";
            version = "dev";
            src = pkgs.lib.cleanSource ./.;
            vendorHash = "sha256-DVXnrBLuJo1CF4JTDqVZmxJrZhygwslUVpeKC/UHmTI=";

            inherit env;

            ldflags = [
              "-s"
              "-w"
              "-X github.com/dhanush0x96c/blueprint/internal/version.Version=${version}"
              "-X github.com/dhanush0x96c/blueprint/internal/version.GitCommit=${commit}"
              "-X github.com/dhanush0x96c/blueprint/internal/version.BuildDate=${buildDate}"
            ];

            meta = with pkgs.lib; {
              description = "Blueprint - composable project scaffolding tool";
              homepage = "https://github.com/dhanush0x96c/blueprint";
              license = licenses.mit;
              mainProgram = "blueprint";
            };
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              # Task runner
              just

              # Go toolchain & language tools
              go
              gopls
              golangci-lint
              delve
              gotools

              # Release & changelog tools
              goreleaser
              git-cliff

              # Version control
              jujutsu
              git

              # Demo & terminal recording
              vhs
              bat
              eza
            ];

            shellHook = ''
              export PATH="$(go env GOPATH)/bin:$PATH"
            '';

            inherit env;
          };
        }
      );
    };
}
