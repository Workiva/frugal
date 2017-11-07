#!/usr/bin/env bash

set -ex

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

if [ ! -e "$FRUGAL_HOME/lib/go/glide.lock" ]; then
    cd $FRUGAL_HOME/lib/go && glide install
fi

cd $FRUGAL_HOME

# Create Go binaries
rm -rf test/integration/go/bin/*
cd test/integration/go
glide install
go build -o bin/testclient src/bin/testclient/main.go
go build -o bin/testserver src/bin/testserver/main.go
