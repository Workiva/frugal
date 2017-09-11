#!/usr/bin/env bash
set -e

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

pub run dart_dev format --check
pub run dart_dev analyze --fatal-lints

$FRUGAL_HOME/scripts/smithy/codecov.sh $FRUGAL_HOME/lib/dart/coverage/coverage.lcov dartlibrarytest
