#!/usr/bin/env bash

set -ex

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

# Dart Dependencies
cd $FRUGAL_HOME/test/integration/dart/test_client
rm -rf .packages packages
pub upgrade

# Try pub get and ignore failures - it will fail on any release
cd $FRUGAL_HOME/test/integration/dart/gen-dart/frugal_test
if pub upgrade ; then
    echo 'pub upgrade returned no error'
else
    echo 'Pub upgrade returned an error we ignored'
fi

# get frugal version to use with manually placing package in pub-cache
frugal_version=$(frugal --version | awk '{print $3}')

# we need to manually install frugal to match with the version of we are testing
# so delete existing package (if above pub get succeeded) and override with the
# current version if not. Otherwise, it will use the latest matching release

rm -rf  ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
mkdir -p ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/

pub_extract_target=~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/

if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    tar -xzf ${SKYNET_APPLICATION_FRUGAL_PUB} -C $pub_extract_target
else
    cp -r $FRUGAL_HOME/lib/dart/* $pub_extract_target
fi

pub get --offline