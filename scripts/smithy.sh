#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail
# Set -e so that we fail if an error is hit.
set -e

ROOT=$PWD

# Retrieve the thrift binary
mkdir -p $ROOT/bin
wget -O $ROOT/bin/thrift https://github.com/stevenosborne-wf/thrift/releases/download/0.9.3-wk-2/thrift-0.9.3-wk-2-linux-amd64 
chmod 0755 $ROOT/bin/thrift
export PATH=$PATH:$ROOT/bin

# Compile the java library code
cd $ROOT/lib/java && mvn verify
mv target/frugal-*.jar $ROOT

# Compile the go library code
cd $ROOT/lib/go
go get -d ./go .
go build

# Compile the python library code
cd $ROOT/lib/python
make deps
make unit

# Run the generator tests
cd $ROOT
go get -d ./compiler .
go build -o frugal
go test ./test
rm -rf ./test/out
