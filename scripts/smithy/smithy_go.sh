#!/usr/bin/env bash
set -e

# Compile library code
pushd lib/go
GO111MODULE=on go mod vendor

# Run the tests
go test -race -coverprofile=../../gocoverage.txt

# Build artifact
popd
tar -czf goLib.tar.gz ./lib/go
