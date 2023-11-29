#!/usr/bin/env bash
set -e

# Run the generator tests
cd $FRUGAL_HOME
CGO_ENABLED=0 GOOS=linux go build -o frugal
go test -race ./...
rm -rf ./compiler/testdata/out
