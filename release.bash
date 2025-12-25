#!/usr/bin/env bash

# Ensure script is run from project root
if [ ! -f "main.go" ] || [ ! -f "go.mod" ]; then
    echo "Error: This script must be run from the project root directory."
    echo "Please cd to the directory containing main.go and go.mod before running this script."
    exit 1
fi

# Use current directory as trimpath
trimpath=$PWD
echo "Building with trimpath: $trimpath"

platforms=("windows/amd64" "linux/386" "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='snowman-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -gcflags=-trimpath=$trimpath -asmflags=-trimpath=$trimpath -o $output_name main.go
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
