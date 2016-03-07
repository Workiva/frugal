#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail
# Set -e so that we fail if an error is hit.
set -e

# Get maven
wget http://apache.claz.org/maven/maven-3/3.3.9/binaries/apache-maven-3.3.9-bin.tar.gz
tar xzvf apache-maven-3.3.9-bin.tar.gz
export PATH=$PWD/bin:$PWD/apache-maven-3.3.9/bin:$PATH

ROOT=$PWD

# Compile the java library code
cd $ROOT/lib/java && mvn verify
mv target/frugal-*.jar $ROOT

# Compile the go library code
cd $ROOT/lib/go
go get -d ./go .
go build

# Run the generator tests
cd $ROOT
go get -d ./compiler .
go build -o frugal
go test ./test
rm -rf ./test/out
