@_default:
  just --list

# runs all snowman tests
test:
  go test -v ./...

# builds snowman for the current platform
build:
  go build -o snowman

# runs snowman with the given command and arguments
run *COMMAND:
  go run main.go {{COMMAND}}

# builds snowman for all officially supported platforms
build-all:
  ./release.bash

# starts the documentation server and file watcher
docs:
  cd docs && mdbook serve

