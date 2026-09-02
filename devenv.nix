{ pkgs, ... }:

{
  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/packages/
  packages = [
    pkgs.mdbook # builds/serves the manual in docs/
    pkgs.git
  ];

  # https://devenv.sh/scripts/
  # These replace the old justfile recipes. Run them from inside `devenv shell`;
  # devenv lists them with their descriptions when the shell starts.
  scripts.build = {
    description = "Build snowman for the current platform";
    exec = ''cd "$DEVENV_ROOT" && go build -o snowman'';
  };

  scripts.build-all = {
    description = "Build snowman for all officially supported platforms";
    exec = ''cd "$DEVENV_ROOT" && ./release.bash'';
  };

  scripts.run = {
    description = "Run snowman from source, e.g. `run build`";
    exec = ''cd "$DEVENV_ROOT" && go run main.go "$@"'';
  };

  scripts.docs = {
    description = "Serve the manual locally with live reload";
    exec = ''cd "$DEVENV_ROOT/docs" && mdbook serve'';
  };

  # https://devenv.sh/tests/
  # `devenv test` replaces the old `just test` recipe.
  enterTest = ''
    cd "$DEVENV_ROOT" && go test -v ./...
  '';

  # https://devenv.sh/git-hooks/
  git-hooks.hooks = {
    gofmt.enable = true;
  };
}
