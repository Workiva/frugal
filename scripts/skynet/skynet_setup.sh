#!/usr/bin/env bash

set -exo pipefail

which glide > /dev/null || {
    curl https://glide.sh/get | sh
}

mkdir -p /go/src/github.com/Workiva/

# Symlink frugal to gopath - this allows skynet-cli editing for interactive/directmount
ln -s /testing/ /go/src/github.com/Workiva/frugal

# Install frugal
if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    mkdir -p $GOPATH/bin
    cp $SKYNET_APPLICATION_FRUGAL_RELEASE $GOPATH/bin/frugal
    mkdir $GOPATH/src/github.com/Workiva/frugal/test/integration/log
    chmod 755 $GOPATH/bin/frugal
else
    cd $GOPATH/src/github.com/Workiva/frugal && go install
fi


# Start gnatsd
gnatsd &

