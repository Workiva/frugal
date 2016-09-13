#!/usr/bin/env bash
set -e

CODECOV_TOKEN='bQ4MgjJ0G2Y73v8JNX6L7yMK9679nbYB'
GORACE="halt_on_error=1"

# DART
# Wrap up package for pub
cd $FRUGAL_HOME
tar -C lib/dart -czf $SMITHY_ROOT/frugal.pub.tgz .

# Compile library code
cd $FRUGAL_HOME/lib/dart
pub get

#generate test runner
pub run dart_dev gen-test-runner

# Run the tests
pub run dart_dev test

# Run coverage
pub run dart_dev coverage --no-html

./tool/codecov.sh
pub run dart_dev format --check
pub run dart_dev analyze

