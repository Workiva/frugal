package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidGoWithAsync(t *testing.T) {
	options := compiler.Options{
		File:  frugalGenFile,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/async/,async",
		Out:   outputDir + "/async",
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/go/variety_async/f_foo_service.txt", filepath.Join(outputDir, "async", "variety", "f_foo_service.go")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestValidGoFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "go:package_prefix=github.com/Workiva/frugal/test/out/",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/go/actual_base/f_types.txt", filepath.Join(outputDir, "actual_base", "golang", "f_types.go")},
		{"expected/go/actual_base/f_basefoo_service.txt", filepath.Join(outputDir, "actual_base", "golang", "f_basefoo_service.go")},

		{"expected/go/variety/f_types.txt", filepath.Join(outputDir, "variety", "f_types.go")},
		{"expected/go/variety/f_foo_service.txt", filepath.Join(outputDir, "variety", "f_foo_service.go")},
		{"expected/go/variety/f_events_scope.txt", filepath.Join(outputDir, "variety", "f_events_scope.go")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

// Ensures correct import references are used when -use-vendor is set and the
// IDL has a vendored include.
func TestValidGoVendor(t *testing.T) {
	options := compiler.Options{
		File:  includeVendor,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/go/vendor/f_myscope_scope.txt", filepath.Join(outputDir, "include_vendor", "f_myscope_scope.go")},
		{"expected/go/vendor/f_myservice_service.txt", filepath.Join(outputDir, "include_vendor", "f_myservice_service.go")},
		{"expected/go/vendor/f_types.txt", filepath.Join(outputDir, "include_vendor", "f_types.go")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

// Ensures an error is returned when -use-vendor is set and the vendored
// include does not specify a path.
func TestValidGoVendorPathNotSpecified(t *testing.T) {
	options := compiler.Options{
		File:  includeVendorNoPath,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
}

// Ensures the target IDL is generated when -use-vendor is set and it has a
// vendored namespace.
func TestValidGoVendorNamespaceTargetGenerate(t *testing.T) {
	options := compiler.Options{
		File:  vendorNamespace,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/go/vendor_namespace/f_types.txt", filepath.Join(outputDir, "vendor_namespace", "f_types.go")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
}
