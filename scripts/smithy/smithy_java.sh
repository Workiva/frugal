#!/usr/bin/env bash
set -e

# JAVA
# Compile library code
cd $FRUGAL_HOME/lib/java && mvn checkstyle:check -q && mvn clean verify -q
mv $(find target -type f -name 'frugal-*.*.*.jar' | grep -v javadoc) $FRUGAL_HOME

