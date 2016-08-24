#!/usr/bin/env bash

set -exo pipefail

# Install thrift
wget -O $GOPATH/bin/thrift https://github.com/stevenosborne-wf/thrift/releases/download/0.9.3-wk-3/thrift-0.9.3-wk-3-linux-amd64
chmod 0755 $GOPATH/bin/thrift		

# Move godeps to gopath for both library and binary
cp -r /testing/Godeps/_workspace/* $GOPATH/
cp -r /testing/lib/go/Godeps/_workspace/* $GOPATH/

# Move frugal itself into gopath
mkdir -p $GOPATH/src/github.com/Workiva
cp -r /testing $GOPATH/src/github.com/Workiva/frugal
cd $GOPATH/src/github.com/Workiva/frugal

# Install frugal
godep go install

# Install and start a nats instance
wget -O gnatsd.tar.gz https://github.com/nats-io/gnatsd/releases/download/v0.7.2/gnatsd-v0.7.2-linux-amd64.tar.gz
tar xzf gnatsd.tar.gz
chmod 0755 gnatsd
./gnatsd &

