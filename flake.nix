{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
  };

  outputs = inputs: {
    packages = builtins.mapAttrs (system: pkgs: {
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
