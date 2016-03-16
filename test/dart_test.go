package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestValidDart(t *testing.T) {
	defer globals.Reset()
	options := compiler.Options{
		File:  validFile,
		Gen:   "dart",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	blahServPath := filepath.Join(outputDir, "valid", "lib", "src", "f_blah_service.dart")
	compareFiles(t, "expected/dart/f_blah_service.dart", blahServPath)
	blahScopePath := filepath.Join(outputDir, "valid", "lib", "src", "f_blah_scope.dart")
	compareFiles(t, "expected/dart/f_blah_scope.dart", blahScopePath)
	fooPath := filepath.Join(outputDir, "valid", "lib", "src", "f_foo_scope.dart")
	compareFiles(t, "expected/dart/f_foo_scope.dart", fooPath)
}
