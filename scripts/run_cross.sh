#!/usr/bin/env bash

# Intentionally leaving out the -e flag so results are posted to skynet run.
# Non-zero exit code is caught in the skynet.yaml
set -x

frugalDir=$PWD

# Clear old logs
rm -rf test/integration/log/*

# RM and Generate Go Code
rm -rf test/integration/go/gen/*
frugal --gen go:package_prefix=github.com/Workiva/frugal/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal

# Create Go binaries
rm -rf test/integration/go/bin/*
go build -o test/integration/go/bin/testclient test/integration/go/src/bin/testclient/main.go
go build -o test/integration/go/bin/testserver test/integration/go/src/bin/testserver/main.go

# RM and Generate Dart Code
rm -rf test/integration/dart/gen-dart/
frugal --gen dart -r --out='test/integration/dart/gen-dart' test/integration/frugalTest.frugal

# Pub get hackery.  This can be fixed when Skynet has pub credentials.
tar xzf $SKYNET_APPLICATION_FRUGAL_LOCAL_CACHE -C .
rm -rf test/intgration/dart/gen-dart/frugal_test/pubspec.yaml
cp -r local_cache/pubspec.yaml test/integration/dart/gen-dart/frugal_test/pubspec.yaml
cd test/integration/dart/test_client
PUB_CACHE=${frugalDir}/local_cache/ pub upgrade --offline
PUB_CACHE=${frugalDir}/local_cache/ pub get --offline
cd ../gen-dart/frugal_test
PUB_CACHE=${frugalDir}/local_cache/ pub upgrade --offline
PUB_CACHE=${frugalDir}/local_cache/ pub get --offline
cd ${frugalDir}

# RM and Generate Java Code
rm -rf test/integration/java/frugal-integration-test/gen-java/
frugal --gen java -r --out='test/integration/java/frugal-integration-test/gen-java' test/integration/frugalTest.frugal

# Java dependency magic
cp $SKYNET_APPLICATION_FRUGAL_ARTIFACTORY ${frugalDir}/test/integration/java/frugal-integration-test/frugal.jar
cd ${frugalDir}/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U
mvn clean compile -U

# Run tests
cd ${frugalDir}
python test/integration/test.py --retry-count=0
