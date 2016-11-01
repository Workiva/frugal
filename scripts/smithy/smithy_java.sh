#!/usr/bin/env bash
set -e

# JAVA
# Compile library code
cd $FRUGAL_HOME/lib/java && mvn checkstyle:check && mvn clean verify
mv target/frugal-*.jar $SMITHY_ROOT

$FRUGAL_HOME/scripts/smithy/codecov.sh $FRUGAL_HOME/lib/java/target/site/jacoco/jacoco.xml java_library