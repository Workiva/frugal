package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidPythonTornado(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "py:tornado",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "valid", "f_Foo_publisher.py")
	compareFiles(t, "expected/python/f_Foo_publisher.py", pubPath)
	pubPath = filepath.Join(outputDir, "valid", "f_blah_publisher.py")
	compareFiles(t, "expected/python/f_blah_publisher.py", pubPath)
	subPath := filepath.Join(outputDir, "valid", "f_Foo_subscriber.py")
	compareFiles(t, "expected/python/f_Foo_subscriber.py", subPath)
	subPath = filepath.Join(outputDir, "valid", "f_blah_subscriber.py")
	compareFiles(t, "expected/python/f_blah_subscriber.py", subPath)
	servicePath := filepath.Join(outputDir, "valid", "f_Blah.py")
	compareFiles(t, "expected/python/f_Blah.py", servicePath)
	initPath := filepath.Join(outputDir, "valid", "__init__.py")
	compareFiles(t, "expected/python/__init__.py", initPath)
}

func TestValidPythonFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:tornado,gen_with_frugal",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	varietyConstantsPath := filepath.Join(outputDir, "variety", "python", "f_constants.py")
	compareFiles(t, "expected/python/variety/f_constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "variety", "python", "f_types.py")
	compareFiles(t, "expected/python/variety/f_types.py", varietyFtypesPath)
	varietyFooPath := filepath.Join(outputDir, "variety", "python", "Foo.py")
	compareFiles(t, "expected/python/variety/Foo.py", varietyFooPath)
}
