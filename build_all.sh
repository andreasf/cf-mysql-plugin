#!/bin/bash
set -e
set -o pipefail

BINARY="cf-mysql-plugin"

main() {
    ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --compilers=2

    build_for_platform_and_arch linux amd64
    build_for_platform_and_arch linux 386

    build_for_platform_and_arch darwin amd64

    build_for_platform_and_arch windows amd64
    build_for_platform_and_arch windows 386

    print_release_yaml
}

build_for_platform_and_arch() {
    platform="$1"
    arch="$2"
    
    mkdir -p output
    build_filename=`build_filename_for_platform "$platform"`
    release_name="output/`release_name_for_platform $platform $arch`"
    GODEBUG=netdns=cgo GOOS="$platform" GOARCH="$arch" go build
    mv "$build_filename" "$release_name"

    hash_val=`shasum ${release_name} | cut -f 1 -d" "`
    hash_var_name="hash_${platform}_${arch}"
    eval "$hash_var_name=${hash_val}"
}

build_filename_for_platform() {
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

release_name_for_platform() {
    platform="$1"
    case "$platform" in
        windows)
            echo "$BINARY-$arch.exe"
            ;;
        *)
            echo "$BINARY-$platform-$arch"
            ;;
    esac
}

print_release_yaml() {
    prefixed_version=$(git describe --tags)
    version=${prefixed_version:1}
    updated=$(date "+%Y-%m-%dT%H:%M:%S%z")

    cat <<EOF
- name: mysql-plugin
  description: Runs mysql and mysqldump clients against your CF database services. Use it to inspect, dump and restore your DB.
  version: ${version}
  created: 2017-02-02T09:30:00Z
  updated: ${updated}
  company:
  authors:
  - name: Andreas Fleig
    homepage: https://github.com/andreasf
    contact: afleig@pivotal.io
  homepage: https://github.com/andreasf/cf-mysql-plugin
  binaries:
  - platform: osx
    url: https://github.com/andreasf/cf-mysql-plugin/releases/download/v${version}/${BINARY}-darwin-amd64
    checksum: ${hash_darwin_amd64}
  - platform: win64
    url: https://github.com/andreasf/cf-mysql-plugin/releases/download/v${version}/${BINARY}-amd64.exe
    checksum: ${hash_windows_amd64}
  - platform: win32
    url: https://github.com/andreasf/cf-mysql-plugin/releases/download/v${version}/${BINARY}-386.exe
    checksum: ${hash_windows_386}
  - platform: linux32
    url: https://github.com/andreasf/cf-mysql-plugin/releases/download/v${version}/${BINARY}-linux-386
    checksum: ${hash_linux_386}
  - platform: linux64
    url: https://github.com/andreasf/cf-mysql-plugin/releases/download/v${version}/${BINARY}-linux-amd64
    checksum: ${hash_linux_amd64}
EOF
}

main

