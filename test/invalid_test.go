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

// Ensures an error is returned when a "*" namespace has a vendor annotation.
func TestWildcardNamespaceWithVendorAnnotation(t *testing.T) {
	options := compiler.Options{
		File:  badNamespace,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
}

// Ensures an error is returned when -use-vendor is used for an unsupported
// language.
func TestVendorUnsupportedLanguage(t *testing.T) {
	options := compiler.Options{
		File:      validFile,
		Gen:       "java",
		Out:       outputDir,
		Delim:     delim,
		UseVendor: true,
	}
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
}
