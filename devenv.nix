{ pkgs, lib, config, inputs, ... }:

{
  # https://devenv.sh/basics/
  env.GREET = "devenv";

  # https://devenv.sh/packages/
  packages = with pkgs; [ git bashInteractive sqlc go-migrate sqlite ];


  # https://devenv.sh/languages/
  # languages.rust.enable = true;

  # https://devenv.sh/processes/
  # processes.cargo-watch.exec = "cargo-watch";

  # https://devenv.sh/services/
  # services.postgres.enable = true;

  # https://devenv.sh/scripts/
  scripts.vf-create-migration.exec = ''
    migrate create -ext sql -dir db/migrations -seq $1
  '';

  scripts.vf-create-dev-db.exec = ''
    rm -rf .temp/db.sqlite3
    sqlite3 .temp/db.sqlite3 "VACUUM;"
  '';

  enterShell = ''
    export SHELL=${pkgs.bashInteractive}/bin/bash
    export VF_TELEGRAM_TOKEN=$(cat $XDG_RUNTIME_DIR/vf_tg_bot_token)
    export VF_OPENAI_TOKEN=$(cat $XDG_RUNTIME_DIR/vf_openai_token)
    export VF_SQL_DB_PATH="$(pwd)/.temp/db.sqlite3"

    git --version
    go version
    echo "sqlc version $(sqlc version)"

    mkdir -p .temp
  '';


  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  # enterTest = ''
  #   echo "Running tests"
  #   git --version | grep --color=auto "${pkgs.git.version}"
  # '';

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
