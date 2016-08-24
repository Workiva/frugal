#!/usr/bin/env bash

set -exo pipefail

./scripts/skynet/skynet_setup.sh

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal
cd $FRUGAL_HOME

# Allow identical operation whether generating with or without thrift
if [ $# -eq 1 ] && [ "$1" == "-gen_with_thrift" ]; then
    gen_with_thrift=true
else
    gen_with_thrift=false
fi

# RM and Generate Go Code
rm -rf test/integration/go/gen/*
if [ "$gen_with_thrift" = true ]; then
    frugal --gen go:package_prefix=github.com/Workiva/frugal/,gen_with_frugal=false -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
else
    frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
fi

# Create Go binaries
rm -rf test/integration/go/bin/*
godep go build -o test/integration/go/bin/testclient test/integration/go/src/bin/testclient/main.go
godep go build -o test/integration/go/bin/testserver test/integration/go/src/bin/testserver/main.go

# RM and Generate Dart Code
rm -rf test/integration/dart/gen-dart/*
if [ "$gen_with_thrift" = true ]; then
    frugal --gen dart:gen_with_frugal=false -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
else
    frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
fi

cd $FRUGAL_HOME/test/integration/dart/test_client
pub get

# Try pub get and ignore failures - it will fail on any release
cd $FRUGAL_HOME/test/integration/dart/gen-dart/frugal_test
if pub get ; then
    echo 'pub get returned no error'
else
    echo 'Pub get returned an error we ignored'
fi

# get frugal version to use with manually placing package in pub-cache
frugal_version=$(frugal --version | awk '{print $3}')

# we need to manually install our package to match with the version of frugal
# so delete existing package (if above pub get succeeded) and override with the
# current version if not
rm -rf  ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
mkdir -p ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
cp -r $FRUGAL_HOME/lib/dart/* ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
pub get --offline


# RM and Generate Java Code
cd $FRUGAL_HOME
rm -rf test/integration/java/frugal-integration-test/gen-java/*
if [ "$gen_with_thrift" = true ]; then
    frugal --gen java:gen_with_frugal=false -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
else
    frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
fi

# Build and install java frugal library
cd $FRUGAL_HOME/lib/java
mvn clean verify
mv target/frugal-*.jar $FRUGAL_HOME/test/integration/java/frugal-integration-test/frugal.jar
cd $FRUGAL_HOME/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U
# Compile java tests
mvn clean compile -U


# Run cross tests - want to report any failures, so don't allow command to exit
# without cleaning up
cd $FRUGAL_HOME
if python test/integration/test.py --retry-count=0 ; then
    /testing/scripts/skynet/test_cleanup.sh
else
    /testing/scripts/skynet/test_cleanup.sh
    exit 1
fi
