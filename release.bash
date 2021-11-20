#!/usr/bin/env bash

platforms=("windows/amd64" "linux/386" "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

echo -n "Please enter the path prefix to trim from the executables: "
read trimpath

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
