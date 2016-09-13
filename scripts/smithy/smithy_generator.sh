#!/usr/bin/env bash
set -e

# Run the generator tests
cd $FRUGAL_HOME
godep go build -o frugal
godep go test -race ./test
mv frugal $SMITHY_ROOT
rm -rf ./test/out
