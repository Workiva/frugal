package test

import (
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestInvalid(t *testing.T) {
	options := compiler.Options{
		File:  invalidFile,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateMethodArgIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateMethodArgIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateStructFieldIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateStructFieldIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}
