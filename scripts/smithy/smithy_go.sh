#!/usr/bin/env bash
set -e
# Get godep
which godep > /dev/null || {
    go get github.com/tools/godep
}

# Compile library code
cd $FRUGAL_HOME/lib/go
godep go build

# Run the tests
godep go test -race

# Build artifact
cd $FRUGAL_HOME
tar -czf $SMITHY_ROOT/goLib.tar.gz ./lib/go