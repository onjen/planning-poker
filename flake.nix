{
  description = "Planning Poker";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
  };

  outputs = inputs: {
    packages = builtins.mapAttrs (system: pkgs: {
      planning-poker = pkgs.buildGo127Module {
        pname = "planning-poker";
        version = "0.1.0";

        src = pkgs.lib.fileset.toSource {
          root = ./.;
          fileset = pkgs.lib.fileset.unions [
            (pkgs.lib.fileset.fileFilter (file: file.hasExt "go") ./.)
            ./go.mod
            ./go.sum
            ./ui/html
            ./ui/static
          ];
        };

        vendorHash = "sha256-5Kr4PvsEMcRxILh/iGRZyoX4I583s4R8rh2CuNhOKNM=";

        meta.mainProgram = "planning-poker";
      };

      default = inputs.self.packages.${system}.planning-poker;
    }) inputs.nixpkgs.legacyPackages;

    devShells = builtins.mapAttrs (system: pkgs: {
      golang = pkgs.mkShell {
        packages = with pkgs; [ go_1_27 ];
        shellHook = ''
          ${pkgs.hello}/bin/hello
      '';
      };

      default = inputs.self.devShells.${system}.golang;
    }) inputs.nixpkgs.legacyPackages;
  };
}
