#!/usr/bin/env bash

set -eo pipefail

go get bitbucket.org/tebeka/go2xunit

# Set the outfile
outfile=$PWD/gotest.out

echo "Running go integration tests"
go test -race -v github.com/Workiva/frugal/test/integration/go | tee $outfile

# Convert the out file to xml
go2xunit -input $outfile -output /testing/reports/go_integration_tests.xml
