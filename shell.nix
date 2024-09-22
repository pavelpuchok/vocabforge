{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  )
, mkGoEnv ? pkgs.mkGoEnv
, gomod2nix ? pkgs.gomod2nix
}:

let
  goEnv = mkGoEnv { pwd = ./.; };
in
pkgs.mkShell {
  packages = [
    goEnv
    gomod2nix
    pkgs.go
    pkgs.golangci-lint
    pkgs.pre-commit
    pkgs.docker
    pkgs.docker-compose
  ];

  shellHook = ''
      export VOCABFORGE_MONGO_USER=supertuperuser
      export VOCABFORGE_MONGO_PASS=supercoolpass
      export VOCABFORGE_MONGO_URI=mongodb://supertuperuser:supercoolpass@localhost:27017/
    '';
}
