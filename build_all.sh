#!/bin/bash
set -e

BINARY="cf-mysql-plugin"

main() {
    ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

    build_for_platform_and_arch linux amd64
    build_for_platform_and_arch linux 386

    build_for_platform_and_arch darwin amd64

    build_for_platform_and_arch windows amd64
    build_for_platform_and_arch windows 386
}

build_for_platform_and_arch() {
    platform="$1"
    arch="$2"
    destination="output/$platform-$arch"
    binary=`binary_for_platform "$platform"`
    mkdir -p "$destination"
    GOOS="$platform" GOARCH="$arch" go build
    mv "$binary" "$destination"

    pushd "$destination" > /dev/null
    zip ../"$BINARY-$platform-$arch.zip" *
    popd > /dev/null
}

binary_for_platform() {
    platform="$1"
    case "$platform" in
        windows)
            echo "$BINARY.exe"
            ;;
        *)
            echo "$BINARY"
            ;;
    esac
}

main

