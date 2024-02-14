@_default:
  just --list

# runs all snowman tests
test:
  go test -v ./...

# builds snowman for the current platform
build:
  go build -o snowman

run:
  go run main.go {{COMMAND}}

# builds snowman for all officially supported platforms
build-all:
  ./release.bash
