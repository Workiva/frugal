#!/usr/bin/env bash

set -exo pipefail

./scripts/skynet/skynet_setup.sh

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal
cd ${FRUGAL_HOME}

# Remove any leftover log files (necessary for skynet-cli)
rm -rf test/integration/log/*

# rm any existing generated code (necessary for skynet-cli)
rm -rf test/integration/go/gen/*
rm -rf test/integration/java/frugal-integration-test/gen-java/*
rm -rf test/integration/python/tornado/gen_py_tornado/*
rm -rf test/integration/python/ascynio/gen_py_asyncio/*
rm -rf test/integration/python/vanilla/gen_py/*
rm -rf test/integration/dart/gen-dart/*

frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
frugal --gen py:tornado -r --out='test/integration/python/tornado/gen_py_tornado' test/integration/frugalTest.frugal
frugal --gen py:asyncio -r --out='test/integration/python/asyncio/gen_py_asyncio' test/integration/frugalTest.frugal
frugal --gen py -r -out='test/integration/python/tornado/gen-py' test/integration/frugalTest.frugal
frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal

# Set everything up in parallel (code generation is fast enough to not require in parallel)
go run scripts/skynet/cross/cross_setup.go

# Run cross tests - want to report any failures, so don't allow command to exit
# without cleaning up
cd ${FRUGAL_HOME}

if go run test/integration/test.go test/integration/tests.json; then
    /testing/scripts/skynet/test_cleanup.sh
else
    /testing/scripts/skynet/test_cleanup.sh
    exit 1
fi