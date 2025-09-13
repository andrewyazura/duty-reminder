{
  description = "Telegram bot for reminding roommates about cleaning duties";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      duty-reminder = pkgs.buildGoModule {
        pname = "duty-reminder";
        version = self.shortRev or "dev";
        src = ./.;
        vendorHash = "sha256-Qk4HXQ4RT7Glsrb/uo2iZEJj8c7SCnsBTY2a0LmD+vw=";

        ldflags = [ "-s" "-w" ];
        env.CGO_ENABLED = 0;
        meta.mainProgram = "duty-reminder";
      };
    in {
      packages.${system} = {
        default = duty-reminder;
        duty-reminder = duty-reminder;
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [ go gopls gotools go-tools postgresql ];
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/app";
      };

      nixosModules.default = { lib, config, ... }:
        let cfg = config.services.duty-reminder;
        in {
          options.services.duty-reminder = {
            enable = lib.mkEnableOption "Enable duty-reminder service";

            environment = lib.mkOption {
              type = lib.types.attrsOf lib.types.str;
              default = { };
              description = "Environment variables passed to the service";
            };

            environmentFile = lib.mkOption {
              type = lib.types.nullOr lib.types.path;
              default = null;
              description = "Optional .env file with secrets";
            };
          };

          config = lib.mkIf cfg.enable {
            users.users.duty-reminder = {
              description = "Duty Reminder service user";
              isSystemUser = true;
              group = "duty-reminder";
            };
            users.groups.duty-reminder = { };

            systemd.services.duty-reminder = {
              description = "Duty Reminder service";
              after = [ "network-online.target" ];
              wants = [ "network-online.target" ];
              wantedBy = [ "multi-user.target" ];

              environment = cfg.environment;

              serviceConfig = {
                User = "duty-reminder";
                Group = "duty-reminder";

                ExecStart =
                  "${self.packages.${system}.default}/bin/duty-reminder";

                Type = "simple";
                Restart = "on-failure";
              } // lib.optionalAttrs (cfg.environmentFile != null) {
                EnvironmentFile = cfg.environmentFile;
              };
            };
          };
        };
    };
}
