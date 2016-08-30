#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail
# Set -e so that we fail if an error is hit.
set -e

# Get godep
which godep > /dev/null || {
    go get github.com/tools/godep
}

ROOT=$PWD
CODECOV_TOKEN='bQ4MgjJ0G2Y73v8JNX6L7yMK9679nbYB'
THRIFT_TAG=0.9.3-wk-2
THRIFT=thrift-$THRIFT_TAG-linux-amd64
GORACE="halt_on_error=1"


# Retrieve the thrift binary
mkdir -p $ROOT/bin
curl -L -O https://github.com/stevenosborne-wf/thrift/releases/download/$THRIFT_TAG/$THRIFT
mv $THRIFT $ROOT/bin/thrift
chmod 0755 $ROOT/bin/thrift
export PATH=$PATH:$ROOT/bin

# JAVA
# Compile library code
cd $ROOT/lib/java && mvn clean verify
mv target/frugal-*.jar $ROOT

# GO
# Compile library code
cd $ROOT/lib/go
godep restore
go build
# Run the tests
go test -race

# DART
# Wrap up package for pub
cd $ROOT
tar -C lib/dart -czf frugal.pub.tgz .
# Compile library code
cd $ROOT/lib/dart
pub get
# Run the tests
pub run dart_dev test
pub run dart_dev coverage --no-html
./tool/codecov.sh
pub run dart_dev format --check
pub run dart_dev analyze

# Python
virtualenv -p /usr/bin/python /tmp/frugal
source /tmp/frugal/bin/activate
pip install -U pip
cd $ROOT/lib/python
make deps-tornado
make deps-gae
make xunit-py2
deactivate

virtualenv -p /usr/bin/python3.5 /tmp/frugal-py3
source /tmp/frugal-py3/bin/activate
pip install -U pip
cd $ROOT/lib/python
make deps-asyncio
make xunit-py3
make install
mv dist/frugal-*.tar.gz $ROOT
deactivate

# Run the generator tests
cd $ROOT
godep restore
go build -o frugal
go test -race ./test
rm -rf ./test/out
