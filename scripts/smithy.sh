#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -eo pipefail

python scripts/smithy/verify_pr_target.py

mkdir -p test_results/

# Run each language build and tests in parallel
go get github.com/sirupsen/logrus
go run scripts/smithy/parallel_smithy.go
