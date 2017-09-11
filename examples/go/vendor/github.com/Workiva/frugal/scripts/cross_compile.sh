#!/bin/bash
go get github.com/mitchellh/gox
go get github.com/tcnksm/ghr

export APPNAME="frugal"
export OSARCH="linux/386 linux/amd64 linux/arm darwin/amd64"
export DIRS="linux-386 linux-amd64 linux-arm darwin-amd64"
export OUTDIR="pkg"
export VERSION=$(grep 'const Version' compiler/globals/globals.go | awk -F" "  '{ print $4 }' | tr -d '"')

gox -osarch="$OSARCH" -ldflags="-s -w" -output "$OUTDIR/$APPNAME-$VERSION-{{.OS}}-{{.Arch}}"
