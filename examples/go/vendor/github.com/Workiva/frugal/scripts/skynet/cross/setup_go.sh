#!/usr/bin/env bash

set -ex

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

cd ${FRUGAL_HOME}

# Create Go binaries
rm -rf test/integration/go/bin/*
godep go build -o test/integration/go/bin/testclient test/integration/go/src/bin/testclient/main.go
godep go build -o test/integration/go/bin/testserver test/integration/go/src/bin/testserver/main.go
