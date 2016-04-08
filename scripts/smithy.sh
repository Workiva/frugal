#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail
# Set -e so that we fail if an error is hit.
set -e

ROOT=$PWD
CODECOV_TOKEN='bQ4MgjJ0G2Y73v8JNX6L7yMK9679nbYB'
THRIFT_TAG=0.9.3-wk-2
THRIFT=thrift-$THRIFT_TAG-linux-amd64


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
go get -d -t ./go .
go build
# Run the tests
go test

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
# zip the library
zip -r frugal-dart.zip $ROOT

# Run the generator tests
cd $ROOT
go get -d ./compiler .
go build -o frugal
go test ./test
rm -rf ./test/out
