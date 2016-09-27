#!/usr/bin/env bash

set -e

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal


# Dart Dependencies
cd $FRUGAL_HOME/test/integration/dart/test_client
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

# we need to manually install our package to match with the version of frugal
# so delete existing package (if above pub get succeeded) and override with the
# current version if not
rm -rf  ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
mkdir -p ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
cp -r $FRUGAL_HOME/lib/dart/* ~/.pub-cache/hosted/pub.workiva.org/frugal-${frugal_version}/
pub get --offline