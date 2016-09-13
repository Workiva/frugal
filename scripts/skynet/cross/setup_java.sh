#!/usr/bin/env bash


export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

# Build and install java frugal library
cd ${FRUGAL_HOME}/lib/java
mvn clean verify -q
mv target/frugal-*.jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/frugal.jar
cd ${FRUGAL_HOME}/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U -q

# Compile java tests
mvn clean compile assembly:single -U -q


mv ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/target/frugal-integration-test-1.0-SNAPSHOT-jar-with-dependencies.jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/cross.jar