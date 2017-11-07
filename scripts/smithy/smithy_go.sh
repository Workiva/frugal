#!/usr/bin/env bash
set -e

which glide > /dev/null || {
    curl https://glide.sh/get | sh
}

# Compile library code
cd $FRUGAL_HOME/lib/go && glide install

# Run the tests
go test -race -coverprofile=$FRUGAL_HOME/gocoverage.txt
$FRUGAL_HOME/scripts/smithy/codecov.sh $FRUGAL_HOME/gocoverage.txt golibrary

# Build artifact
cd $FRUGAL_HOME && tar -czf $SMITHY_ROOT/goLib.tar.gz ./lib/go
