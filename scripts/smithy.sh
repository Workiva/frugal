#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -eo pipefail

python $FRUGAL_HOME/scripts/smithy/verify_pr_target.py

mkdir -p $FRUGAL_HOME/test_results/

# Run each language build and tests in parallel
go get github.com/sirupsen/logrus
cd $FRUGAL_HOME && go run scripts/smithy/parallel_smithy.go
