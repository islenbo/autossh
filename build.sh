#!/bin/bash

PROJECT="autossh"
VERSION="v1.1.0"
BUILD=`date +%FT%T%z`

function build() {
    os=$1
    arch=$2
    alias_name=$3
    package="${PROJECT}-${alias_name}-${arch}_${VERSION}"

    echo "build ${package} ..."
    mkdir -p "./releases/${package}"
    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -o "./releases/${package}/autossh" -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}" src/main/main.go
    cp ./config.example.json "./releases/${package}/config.json"
    chmod +x ./install
    cp ./install "./releases/${package}/install"
    cd ./releases/
    zip -r "./${package}.zip" "./${package}"
    echo "clean ${package}"
    rm -rf "./${package}"
    cd ../
}

# OS X Mac
build darwin amd64 macOS

# Linux
build linux amd64 linux
build linux 386 linux
build linux arm linux
