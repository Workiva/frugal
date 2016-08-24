#!/usr/bin/env bash

set -eo pipefail

./scripts/skynet/skynet_setup.sh
export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

# Remove and regenerate go example code
cd $FRUGAL_HOME/example
rm -rf go/gen-go
frugal --gen go:package_prefix=github.com/Workiva/frugal/example/go/gen-go/ -r --out='go/gen-go' event.frugal


echo "Running go integration tests"

# Set the outfile for tests and run tests
outfile=$FRUGAL_HOME/gotest.out

cd $FRUGAL_HOME
godep go test -race -v github.com/Workiva/frugal/test/integration/go | tee ${outfile}

# Convert the out file to xml
go2xunit -input $outfile -output /testing/reports/go_integration_tests.xml