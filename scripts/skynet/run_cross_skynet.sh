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
rm -rf test/integration/python/tornado/gen_py_tornado/*
rm -rf test/integration/python/ascynio/gen_py_asyncio/*
rm -rf test/integration/dart/gen-dart/*

# Generate code
if [ "$gen_with_thrift" = true ]; then
    frugal --gen go:package_prefix=github.com/Workiva/frugal/,gen_with_frugal=false -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
    frugal --gen java:gen_with_frugal=false -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
    frugal --gen py:tornado,gen_with_frugal=false -r --out='test/integration/python/tornado/gen_py_tornado' test/integration/frugalTest.frugal
    frugal --gen py:asyncio,gen_with_frugal=false -r --out='test/integration/python/asyncio/gen_py_asyncio' test/integration/frugalTest.frugal
    frugal --gen dart:gen_with_frugal=false -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
else
    frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal
    frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal
    frugal --gen py:tornado -r --out='test/integration/python/tornado/gen_py_tornado' test/integration/frugalTest.frugal
    frugal --gen py:asyncio -r --out='test/integration/python/asyncio/gen_py_asyncio' test/integration/frugalTest.frugal
    frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal
fi

# Create Go binaries
rm -rf test/integration/go/bin/*
godep go build -o test/integration/go/bin/testclient test/integration/go/src/bin/testclient/main.go
godep go build -o test/integration/go/bin/testserver test/integration/go/src/bin/testserver/main.go

# Python Dependencies
cd ${FRUGAL_HOME}/lib/python
pip install -e ".[tornado]"
pip install -r requirements_dev_tornado.txt

# Asyncio
python3.5 /usr/bin/pip3 install -e ".[asyncio]"
python3.5 /usr/bin/pip3 install -q -r requirements_dev_asyncio.txt

# Dart Dependencies
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

# Build and install java frugal library
cd ${FRUGAL_HOME}/lib/java
mvn clean verify -q
mv target/frugal-*.jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/frugal.jar
cd ${FRUGAL_HOME}/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U -q
# Compile java tests
mvn clean compile -U -q

# Run cross tests - want to report any failures, so don't allow command to exit
# without cleaning up
cd ${FRUGAL_HOME}
if python test/integration/test.py ; then
    /testing/scripts/skynet/test_cleanup.sh
else
    /testing/scripts/skynet/test_cleanup.sh
    exit 1
fi
