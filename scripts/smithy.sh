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
# Compile library code
cd $ROOT/lib/dart
pub get
cp ./pubspec.lock $ROOT
# Run the tests
pub run dart_dev test
pub run dart_dev coverage --no-html
./tool/codecov.sh
pub run dart_dev format --check
pub run dart_dev analyze

# Python
virtualenv -p /usr/bin/python /tmp/frugal
source /tmp/frugal/bin/activate
cd $ROOT/lib/python
make deps
make xunit

# Run the generator tests
cd $ROOT
godep restore
go build -o frugal
go test -race ./test
rm -rf ./test/out
