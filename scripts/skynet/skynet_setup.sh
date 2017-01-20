#!/usr/bin/env bash

set -exo pipefail

mkdir -p /go/src/github.com/Workiva/

# Move godeps to gopath for both library and binary
cp -r /testing/Godeps/_workspace/* $GOPATH/
cp -r /testing/lib/go/Godeps/_workspace/* $GOPATH/


# Symlink frugal to gopath - this allows skynet-cli editing for interactive/directmount
ln -s /testing/ /go/src/github.com/Workiva/frugal

# Install frugal
if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    mkdir -p $GOPATH/bin
    cp ${SKYNET_APPLICATION_FRUGAL_RELEASE} $GOPATH/bin/frugal
    mkdir $GOPATH/src/github.com/Workiva/frugal/test/integration/log
    chmod 755 ${GOPATH}/bin/frugal
else
    cd $GOPATH/src/github.com/Workiva/frugal
    godep go install
fi


# Start gnatsd
gnatsd &

