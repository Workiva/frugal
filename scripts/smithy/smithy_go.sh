#!/usr/bin/env bash
set -e

# Compile library code
cd $FRUGAL_HOME/lib/go && GO111MODULE=on go mod vendor

# Run the tests
go test -race -coverprofile=$FRUGAL_HOME/gocoverage.txt

# Build artifact
cd $FRUGAL_HOME && tar -czf $FRUGAL_HOME/goLib.tar.gz ./lib/go
