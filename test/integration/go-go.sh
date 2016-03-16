#!/usr/bin/env bash

# Remove generated code
# Generate from frugal file
frugal --gen go:package_prefix=github.com/Workiva/frugal/test/integration/go/gen/ -r --out='go/gen' frugalTest.frugal
# Run go server
# Run go client

