#!/usr/bin/env bash

set -ex

gnatsd &
frugalDir=$PWD

# Start with clean log folder
rm -rf test/integration/log/*

# RM and Generate Go Code
rm -rf test/integration/go/gen/*
if [ $# -eq 1 ] && [ "$1" == "-gen_with_frugal" ]; then
    frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
else
    frugal --gen go:package_prefix=github.com/Workiva/frugal/,gen_with_frugal=false -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
fi

# Create Go binaries
rm -rf test/integration/go/bin/*
go build -o test/integration/go/bin/testclient test/integration/go/src/bin/testclient/main.go
go build -o test/integration/go/bin/testserver test/integration/go/src/bin/testserver/main.go

# RM and Generate Dart Code
rm -rf test/integration/dart/gen-dart/
if [ $# -eq 1 ] && [ "$1" == "-gen_with_frugal" ]; then
    frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
else
    frugal --gen dart:gen_with_frugal=false -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
fi

cd test/integration/dart/test_client
pub upgrade
cd ../gen-dart/frugal_test
pub upgrade
cd ${frugalDir}

# RM and Generate Java Code
rm -rf test/integration/java/frugal-integration-test/gen-java/
frugal --gen java:gen_with_frugal=false -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal

cd lib/java
mvn clean verify
mv target/frugal-*.jar ${frugalDir}/test/integration/java/frugal-integration-test/frugal.jar
cd ${frugalDir}/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U
mvn clean compile -U

cd ${frugalDir}
# Run tests
# -v flag for verbose output
# --server for specific server languages (only go supported currently)
# --client for specific client languages (go and dart supported currently)
# Example: python test/integration/test.py --server go --client go
python test/integration/test.py --retry-count=0

# After running this script once, you can just run:     python test/integration/test.py --retry-count=0
# to run all cross language tests. Be sure to re-run the entire script if you make changes between test runs.
pkill gnatsd
