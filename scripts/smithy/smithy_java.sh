#!/usr/bin/env bash
set -e

# JAVA
# Compile library code
cd $FRUGAL_HOME/lib/java && mvn checkstyle:check && mvn clean verify
mv target/frugal-*.jar $SMITHY_ROOT
