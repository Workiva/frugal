#!/usr/bin/env bash
set -e

# Wrap up package for pub
cd $FRUGAL_HOME
tar -C lib/dart -czf $FRUGAL_HOME/frugal.pub.tgz .

# Compile library code
cd $FRUGAL_HOME/lib/dart
timeout 5m dart pub get

# Run the tests
dart test

dart format --set-exit-if-changed -o none lib test
