#!/bin/bash

VERSION="v0.1"
PROJECT="autossh"

function build() {
    os=$1
    arch=$2
    package="${PROJECT}-${os}-${arch}_${VERSION}"

    echo "build ${package} ..."
    mkdir -p "./releases/${package}"
    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -o "./releases/${package}/autossh" main.go
    cp ./servers.example.json "./releases/${package}/servers.json"
    cd ./releases/
    zip -r "./${package}.zip" "./${package}"
    echo "clean ${package}"
    rm -rf "./${package}"
    cd ../
}

# Linux
build linux amd64
build linux 386
build linux arm

# OS X Mac
build darwin amd64
