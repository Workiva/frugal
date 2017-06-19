#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -eo pipefail

python $SMITHY_ROOT/scripts/smithy/verify_pr_target.py

mkdir -p $SMITHY_ROOT/test_results/

# Test deps
go get github.com/Sirupsen/logrus
go get github.com/stretchr/testify/assert

# Run each language build and tests in parallel
cd $FRUGAL_HOME && go run scripts/smithy/parallel_smithy.go