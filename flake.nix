{
  description = "telegram bot for reminding roommates about cleaning duties";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      lastModifiedDate =
        self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;
      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          duty-reminder = pkgs.buildGoModule {
            pname = "duty-reminder";
            inherit version;
            src = ./.;
            vendorHash = "sha256-Qk4HXQ4RT7Glsrb/uo2iZEJj8c7SCnsBTY2a0LmD+vw=";
          };
        });

      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools go-tools postgresql ];
          };
        });

      defaultPackage =
        forAllSystems (system: self.packages.${system}.duty-reminder);
    };
}
