#!/usr/bin/env bash

# Remove generated code
rm -rf test/integration/go/gen/frugalTest
frugal --gen go:package_prefix=github.com/Workiva/frugal/test/integration/go/gen/ -r --out='test/integration/go/gen' test/integration/frugalTest.frugal

echo "Starting Server"
go run test/integration/go/runserver.go &  # This needs to be killed at the end of the run.

sleep 2

echo "Running Client"
go run test/integration/go/runclient.go
