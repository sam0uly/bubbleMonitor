{
  description = "bubbleMonitor — a TUI system monitor";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forEachSystem = nixpkgs.lib.genAttrs supportedSystems;
      version = self.shortRev or self.dirtyShortRev or "dev";

      hasNvidia = builtins.pathExists /dev/nvidia0;

      hasAmd =
        let
          drmDir = /sys/class/drm;
          cards =
            if builtins.pathExists drmDir then
              builtins.filter (n: builtins.match "card[0-9]+" n != null) (
                builtins.attrNames (builtins.readDir drmDir)
              )
            else
              [ ];
          isAmdCard =
            card:
            let
              vendor = drmDir + "/${card}/device/vendor";
            in
            builtins.pathExists vendor
            && (builtins.tryEval (builtins.readFile vendor)).success
            && (builtins.readFile vendor) == "0x1002\n";
        in
        builtins.any isAmdCard cards;
    in
    {
      packages = forEachSystem (
        system:
        let
          # amdsmi is x86_64-linux only in nixpkgs; gate AMD on system too.
          canAmd = hasAmd && system == "x86_64-linux";

          pkgs = import nixpkgs {
            inherit system;
            config.allowUnfree = true;
            config.cudaSupport = hasNvidia;
            config.rocmSupport = canAmd;
          };

          cgoCflags = pkgs.lib.concatStringsSep " " (
            pkgs.lib.optionals hasNvidia [ "-I${pkgs.cudaPackages.cuda_nvml_dev}/include" ]
            ++ pkgs.lib.optionals canAmd [ "-I${pkgs.rocmPackages.amdsmi}/include" ]
            ++ pkgs.lib.optionals canAmd [
              "-DUNKNOWN=AMDSMI_PROCESSOR_TYPE_UNKNOWN"
              "-DAMD_GPU=AMDSMI_PROCESSOR_TYPE_AMD_GPU"
              "-DAMD_CPU=AMDSMI_PROCESSOR_TYPE_AMD_CPU"
              "-DNON_AMD_GPU=AMDSMI_PROCESSOR_TYPE_NON_AMD_GPU"
              "-DNON_AMD_CPU=AMDSMI_PROCESSOR_TYPE_NON_AMD_CPU"
              "-DAMD_CPU_CORE=AMDSMI_PROCESSOR_TYPE_AMD_CPU_CORE"
              "-DAMD_APU=AMDSMI_PROCESSOR_TYPE_AMD_APU"
            ]
          );
        in
        {
          default = pkgs.buildGoModule {
            pname = "bubbleMonitor";
            inherit version;
            src = ./.;

            subPackages = [ "cmd/bub" ];

            # Re-run with `lib.fakeSha256` and copy the error's hash on go.mod change.
            vendorHash = "sha256-xdCzlxaqjbPmbLczClFSY7mvY1phGI5uRzoP1R7XMC4=";

            nativeBuildInputs = [ pkgs.pkg-config ];

            buildInputs =
              pkgs.lib.optionals hasNvidia [ pkgs.cudaPackages.cuda_nvml_dev ]
              ++ pkgs.lib.optionals canAmd [ pkgs.rocmPackages.amdsmi ];

            tags = pkgs.lib.optional (!hasNvidia) "nonnvml" ++ pkgs.lib.optional (!canAmd) "noamd";

            env.CGO_ENABLED = "1";
            env.CGO_CFLAGS = cgoCflags;

            ldflags = [
              "-s"
              "-w"
              "-X main.buildVersion=${version}"
              "-X main.buildCommit=${self.rev or self.shortRev or "unknown"}"
              "-X main.buildDate=${toString self.lastModified or "unknown"}"
            ];

            meta = with pkgs.lib; {
              description = "TUI system monitor";
              license = licenses.mit;
              platforms = supportedSystems;
              mainProgram = "bub";
            };
          };
        }
      );

      apps = forEachSystem (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/bub";
        };
      });

      devShells = forEachSystem (
        system:
        let
          canAmd = hasAmd && system == "x86_64-linux";

          pkgs = import nixpkgs {
            inherit system;
            config.allowUnfree = true;
            config.cudaSupport = hasNvidia;
            config.rocmSupport = canAmd;
          };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              go-tools
              gnumake
              pkg-config
              cudaPackages.cuda_nvml_dev
              # Provides `aarch64-unknown-linux-gnu-gcc` for goreleaser linux/arm64.
              pkgsCross.aarch64-multiplatform.buildPackages.gcc
            ];

            shellHook = ''
              export CGO_ENABLED=1
              export CGO_CFLAGS="-I${pkgs.cudaPackages.cuda_nvml_dev}/include"
              echo "bubbleMonitor dev shell ready (cgo enabled, NVML headers in scope)."
            '';
          };
        }
      );

      overlays.default = final: prev: {
        bubbleMonitor = self.packages.${prev.system}.default;
      };
    };
}
