#!/usr/bin/env bash

set -exo pipefail

./scripts/skynet/skynet_setup.sh

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal
cd ${FRUGAL_HOME}

# Remove any leftover log files (necessary for skynet-cli)
rm -rf test/integration/log/*

# Allow identical operation whether generating with or without thrift
if [ $# -eq 1 ] && [ "$1" == "-gen_with_thrift" ]; then
    gen_with_thrift=true
else
    gen_with_thrift=false
fi

# rm any existing generated code (necessary for skynet-cli)
rm -rf test/integration/go/gen/*
rm -rf test/integration/java/frugal-integration-test/gen-java/*
rm -rf test/integration/python/gen_py_tornado/*
rm -rf test/integration/dart/gen-dart/*

# Generate code
if [ "$gen_with_thrift" = true ]; then
    frugal --gen go:package_prefix=github.com/Workiva/frugal/,gen_with_frugal=false -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
    frugal --gen java:gen_with_frugal=false -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
    frugal --gen py:tornado,gen_with_frugal=false -r --out='test/integration/python/gen_py_tornado' test/integration/frugalTest.frugal
    frugal --gen dart:gen_with_frugal=false -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
else
    frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
    frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
    frugal --gen py:tornado -r --out='test/integration/python/gen_py_tornado' test/integration/frugalTest.frugal
    frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
fi

# Set everything up in parallel (code generation is fast enough to not require in parallel)
go run scripts/skynet/cross/cross_setup.go


# Run cross tests - want to report any failures, so don't allow command to exit
# without cleaning up
cd ${FRUGAL_HOME}
if python test/integration/test.py ; then
    /testing/scripts/skynet/test_cleanup.sh
else
    /testing/scripts/skynet/test_cleanup.sh
    exit 1
fi
