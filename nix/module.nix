self:
{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.services.planning-poker;
in
{
  options.services.planning-poker = {
    enable = lib.mkEnableOption "the Planning Poker server";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.planning-poker;
      defaultText = lib.literalExpression "planning-poker.packages.\${system}.planning-poker";
      description = "The planning-poker package to run.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 4000;
      description = "Port the server listens on.";
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Whether to open {option}`services.planning-poker.port` in the firewall.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.planning-poker = {
      description = "Planning Poker";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        ExecStart = "${lib.getExe cfg.package} -port ${toString cfg.port}";
        Restart = "on-failure";
        RestartSec = 5;

        DynamicUser = true;

        # The service keeps all state in memory and serves its assets from the
        # binary, so it needs nothing but a socket.
        AmbientCapabilities = lib.optional (cfg.port < 1024) "CAP_NET_BIND_SERVICE";
        CapabilityBoundingSet = lib.optional (cfg.port < 1024) "CAP_NET_BIND_SERVICE";
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        PrivateTmp = true;
        ProcSubset = "pid";
        ProtectClock = true;
        ProtectControlGroups = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        ProtectSystem = "strict";
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
        SystemCallFilter = [
          "@system-service"
          "~@privileged"
          "~@resources"
        ];
        UMask = "0077";
      };
    };

    networking.firewall.allowedTCPPorts = lib.mkIf cfg.openFirewall [ cfg.port ];
  };
}
