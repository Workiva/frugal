package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidPython(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "py:gen_with_frugal=false",
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

func TestValidPythonAsyncio(t *testing.T) {
	outDir := outputDir + "/asyncio"
	options := compiler.Options{
		File:  validFile,
		Gen:   "py:asyncio",
		Out:   outDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outDir, "valid", "f_Foo_publisher.py")
	compareFiles(t, "expected/python.asyncio/f_Foo_publisher.py", pubPath)
	pubPath = filepath.Join(outDir, "valid", "f_blah_publisher.py")
	compareFiles(t, "expected/python.asyncio/f_blah_publisher.py", pubPath)
	subPath := filepath.Join(outDir, "valid", "f_Foo_subscriber.py")
	compareFiles(t, "expected/python.asyncio/f_Foo_subscriber.py", subPath)
	subPath = filepath.Join(outDir, "valid", "f_blah_subscriber.py")
	compareFiles(t, "expected/python.asyncio/f_blah_subscriber.py", subPath)
	servicePath := filepath.Join(outDir, "valid", "f_Blah.py")
	compareFiles(t, "expected/python.asyncio/f_Blah.py", servicePath)
	initPath := filepath.Join(outDir, "valid", "__init__.py")
	compareFiles(t, "expected/python.asyncio/__init__.py", initPath)
}

func TestValidPythonTornado(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "py:tornado,gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "valid", "f_Foo_publisher.py")
	compareFiles(t, "expected/python.tornado/f_Foo_publisher.py", pubPath)
	pubPath = filepath.Join(outputDir, "valid", "f_blah_publisher.py")
	compareFiles(t, "expected/python.tornado/f_blah_publisher.py", pubPath)
	subPath := filepath.Join(outputDir, "valid", "f_Foo_subscriber.py")
	compareFiles(t, "expected/python.tornado/f_Foo_subscriber.py", subPath)
	subPath = filepath.Join(outputDir, "valid", "f_blah_subscriber.py")
	compareFiles(t, "expected/python.tornado/f_blah_subscriber.py", subPath)
	servicePath := filepath.Join(outputDir, "valid", "f_Blah.py")
	compareFiles(t, "expected/python.tornado/f_Blah.py", servicePath)
	initPath := filepath.Join(outputDir, "valid", "__init__.py")
	compareFiles(t, "expected/python.tornado/__init__.py", initPath)
}

func TestValidPythonFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:tornado",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	baseConstantsPath := filepath.Join(outputDir, "actual_base", "python", "constants.py")
	compareFiles(t, "expected/python/actual_base/constants.py", baseConstantsPath)
	baseFtypesPath := filepath.Join(outputDir, "actual_base", "python", "ttypes.py")
	compareFiles(t, "expected/python/actual_base/ttypes.py", baseFtypesPath)
	baseFooPath := filepath.Join(outputDir, "actual_base", "python", "BaseFoo.py")
	compareFiles(t, "expected/python/actual_base/BaseFoo.py", baseFooPath)

	varietyConstantsPath := filepath.Join(outputDir, "variety", "python", "constants.py")
	compareFiles(t, "expected/python/variety/constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "variety", "python", "ttypes.py")
	compareFiles(t, "expected/python/variety/ttypes.py", varietyFtypesPath)
	varietyFooPath := filepath.Join(outputDir, "variety", "python", "Foo.py")
	compareFiles(t, "expected/python/variety/Foo.py", varietyFooPath)
}
