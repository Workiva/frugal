#!/usr/bin/env bash

set -ex

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal
cd $FRUGAL_HOME/test/integration/go
go mod tidy

# Create Go binaries
rm -rf test/integration/go/bin/*
go build -o bin/testclient src/bin/testclient/main.go
go build -o bin/testserver src/bin/testserver/main.go
go build -o bin/testpublisher src/bin/testpublisher/main.go
go build -o bin/testsubscriber src/bin/testsubscriber/main.go
