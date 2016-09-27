package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidGoWithAsync(t *testing.T) {
	options := compiler.Options{
		File: frugalGenFile,
		Gen: "go:package_prefix=github.com/Workiva/frugal/test/out/async/,async",
		Out: outputDir + "/async",
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	fooServicePath := filepath.Join(outputDir, "async", "variety", "f_foo_service.go")
	compareFiles(t, "expected/go/variety_async/f_foo_service.txt", fooServicePath)
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
		t.Fatal("unexpected error", err)
	}

	baseFtypesPath := filepath.Join(outputDir, "actual_base", "f_types.go")
	compareFiles(t, "expected/go/actual_base/f_types.txt", baseFtypesPath)
	baseFbasefooPath := filepath.Join(outputDir, "actual_base", "f_basefoo_service.go")
	compareFiles(t, "expected/go/actual_base/f_basefoo_service.txt", baseFbasefooPath)

	varietyFtypesPath := filepath.Join(outputDir, "variety", "f_types.go")
	compareFiles(t, "expected/go/variety/f_types.txt", varietyFtypesPath)
	varietyFfooPath := filepath.Join(outputDir, "variety", "f_foo_service.go")
	compareFiles(t, "expected/go/variety/f_foo_service.txt", varietyFfooPath)
	varietyFeventsScopePath := filepath.Join(outputDir, "variety", "f_events_scope.go")
	compareFiles(t, "expected/go/variety/f_events_scope.txt", varietyFeventsScopePath)
}
