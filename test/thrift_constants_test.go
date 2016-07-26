package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

const (
	frugalFile   = "idl/constants.frugal"
	thriftFile   = "idl/constants.thrift"
	expectedFile = "expected/thrift/constants.thrift"
)

func TestThriftConstants(t *testing.T) {
	options := compiler.Options{
		File:               frugalFile,
		Gen:                "go:gen_with_frugal=false",
		Out:                "out",
		Delim:              ".",
		DryRun:             true,
		RetainIntermediate: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}
	outputPath := filepath.Join("idl", "constants.thrift")
	compareFiles(t, expectedFile, outputPath)

	if err := os.Remove(thriftFile); err != nil {
		t.Fatalf("Failed to remove intermediate IDL %s\n", thriftFile)
	}
}
