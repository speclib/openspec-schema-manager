{
  description = "Discover, inspect, install, author and compose OpenSpec workflow schemas";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

      version = nixpkgs.lib.removeSuffix "\n" (builtins.readFile ./VERSION);
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          ossm = pkgs.callPackage ./package.nix {
            inherit version;
            buildGoModule = pkgs.buildGo126Module;
          };
        in
        {
          inherit ossm;
          default = ossm;
        });

      checks = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          build = self.packages.${system}.ossm;

          # scripts/coverage-gate.sh is the single source of truth for what
          # passing means. This derivation calls it and restates none of it.
          gate = pkgs.buildGo126Module {
            pname = "ossm-gate";
            inherit version;
            src = ./.;
            vendorHash = "sha256-Eqdr8gBEStdsx7Ht97P7IzBacvmrE/Fcowm/LEYHVkg=";

            nativeBuildInputs = [ pkgs.bash pkgs.git ];

            buildPhase = "true";

            doCheck = true;
            checkPhase = ''
              runHook preCheck
              export HOME="$TMPDIR"
              bash scripts/coverage-gate.sh
              runHook postCheck
            '';

            installPhase = ''
              mkdir -p $out
              echo "gate passed" > $out/result
            '';
          };
        });

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_26
              gopls
              git
              jujutsu
              openspec
            ];
          };
        });
    };
}
