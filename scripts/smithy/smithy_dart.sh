#!/usr/bin/env bash
set -e

# Wrap up package for pub
tar -C lib/dart -czf frugal.pub.tgz .

# Compile library code
cd lib/dart
timeout 5m dart pub get

# Run the tests
dart test

dart format --set-exit-if-changed -o none lib test
