#!/usr/bin/env bash
set -ex

cd $GOPATH/src/github.com/Workiva/frugal
FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

# If any failures exist, copy these as an artifact for skynet to load
# as a test artifact
if [ -f ${FRUGAL_HOME}/test/integration/log/unexpected_failures.log ]; then
    cp -r ${FRUGAL_HOME}/test/integration/log/unexpected_failures.log /testing/artifacts/unexpected_failures.log
fi
tar -czf successful_tests.tar.gz test/integration/log
mv successful_tests.tar.gz /testing/artifacts/

pkill gnatsd
