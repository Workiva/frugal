package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidGo(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/,gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	fooPath := filepath.Join(outputDir, "valid", "f_foo_scope.go")
	compareFiles(t, "expected/go/f_foo_scope.txt", fooPath)
	blahServPath := filepath.Join(outputDir, "valid", "f_blah_service.go")
	compareFiles(t, "expected/go/f_blah_service.txt", blahServPath)
	blahScopePath := filepath.Join(outputDir, "valid", "f_blah_scope.go")
	compareFiles(t, "expected/go/f_blah_scope.txt", blahScopePath)
}

func TestValidGoWithAsync(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/test/out/,async,gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	blahServPath := filepath.Join(outputDir, "valid", "f_blah_service.go")
	compareFiles(t, "expected/go/f_blah_service_async.txt", blahServPath)
}

func TestValidGoFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "go:package_prefix=github.com/Workiva/frugal/test/out/,gen_with_frugal=true",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	baseFtypesPath := filepath.Join(outputDir, "actual_base", "f_types.go")
	compareFiles(t, "expected/go/actual_base/f_types.txt", baseFtypesPath)
	baseFbasefooPath := filepath.Join(outputDir, "actual_base", "f_basefoo.go")
	compareFiles(t, "expected/go/actual_base/f_basefoo.txt", baseFbasefooPath)

	varietyFtypesPath := filepath.Join(outputDir, "variety", "f_types.go")
	compareFiles(t, "expected/go/variety/f_types.txt", varietyFtypesPath)
	varietyFfooPath := filepath.Join(outputDir, "variety", "f_foo.go")
	compareFiles(t, "expected/go/variety/f_foo.txt", varietyFfooPath)
}
