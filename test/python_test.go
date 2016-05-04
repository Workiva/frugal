package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidPython(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "py",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "valid", "foo_publisher.py")
	compareFiles(t, "expected/python/foo_publisher.py", pubPath)
	pubPath = filepath.Join(outputDir, "valid", "blah_publisher.py")
	compareFiles(t, "expected/python/blah_publisher.py", pubPath)
}
