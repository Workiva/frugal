#!/usr/bin/env bash

set -ex

ROOT=$PWD
cd ./test/integration/dart/test_client
mkdir local_cache
PUB_CACHE=./local_cache/ pub get
cp $ROOT/test/integration/dart/gen-dart/frugal_test/pubspec.yaml local_cache/pubspec.yaml
tar -czf local_cache.tar.gz ./local_cache
rm -rf ./local_cache
mv local_cache.tar.gz $ROOT
