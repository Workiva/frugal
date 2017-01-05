#!/usr/bin/env bash

# This is so `tee` doesn't absorb a non-zero exit code
set -eo pipefail

python $SMITHY_ROOT/scripts/smithy/verify_pr_target.py

mkdir -p $SMITHY_ROOT/test_results/

# Copy the vendored deps into the $GOPATH so we dont need
# another Godeps workspace.
cp -r $FRUGAL_HOME/vendor/* $GOPATH/src/
cp -r $FRUGAL_HOME/lib/go/vendor/* $GOPATH/src

# Run each language build and tests in parallel
cd $FRUGAL_HOME
go run scripts/smithy/parallel_smithy.go
