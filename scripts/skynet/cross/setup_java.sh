#!/usr/bin/env bash

set -ex

export FRUGAL_HOME=$GOPATH/src/github.com/Workiva/frugal

if [ -z "${IN_SKYNET_CLI+yes}" ]; then
    jar=$(find ${SKYNET_DIRECTORY_FRUGAL_JAVA} -name 'frugal-*.jar' ! -name 'frugal-*-*.jar')
    cp $jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/frugal.jar
else
    cd ${FRUGAL_HOME}/lib/java
    mvn clean verify -q
    rm target/frugal-*-*.jar
    mv target/frugal-*.jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/frugal.jar
fi

cd ${FRUGAL_HOME}/test/integration/java/frugal-integration-test
mvn clean install:install-file -Dfile=frugal.jar -U -q

# Compile java tests
mvn clean compile assembly:single -U -q

mv ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/target/frugal-integration-test-1.0-SNAPSHOT-jar-with-dependencies.jar ${FRUGAL_HOME}/test/integration/java/frugal-integration-test/cross.jar
