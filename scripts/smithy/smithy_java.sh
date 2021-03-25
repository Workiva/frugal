#!/usr/bin/env bash
set -e

# JAVA
# Compile library code
# TODO: Re-enable checkstyle
# cd $FRUGAL_HOME/lib/java && mvn checkstyle:check -q && mvn clean verify -q
cd $FRUGAL_HOME/lib/java && mvn clean test -q
mv $(find target -type f -name 'frugal-*.*.*.jar' | grep -v sources | grep -v javadoc) $FRUGAL_HOME

