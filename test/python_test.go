package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidPython(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/python/variety/constants.py", filepath.Join(outputDir, "variety", "python", "constants.py")},
		{"expected/python/variety/ttypes.py", filepath.Join(outputDir, "variety", "python", "ttypes.py")},
		{"expected/python/variety/f_Events_publisher.py", filepath.Join(outputDir, "variety", "python", "f_Events_publisher.py")},
		{"expected/python/variety/f_Events_subscriber.py", filepath.Join(outputDir, "variety", "python", "f_Events_subscriber.py")},
		{"expected/python/variety/f_Foo.py", filepath.Join(outputDir, "variety", "python", "f_Foo.py")},

		{"expected/python/actual_base/constants.py", filepath.Join(outputDir, "actual_base", "python", "constants.py")},
		{"expected/python/actual_base/ttypes.py", filepath.Join(outputDir, "actual_base", "python", "ttypes.py")},
		{"expected/python/actual_base/f_BaseFoo.py", filepath.Join(outputDir, "actual_base", "python", "f_BaseFoo.py")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestValidPythonTornado(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:tornado",
		Out:     outputDir + "/tornado",
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/python.tornado/variety/constants.py", filepath.Join(outputDir, "tornado", "variety", "python", "constants.py")},
		{"expected/python.tornado/variety/ttypes.py", filepath.Join(outputDir, "tornado", "variety", "python", "ttypes.py")},
		{"expected/python.tornado/variety/f_Events_publisher.py", filepath.Join(outputDir, "tornado", "variety", "python", "f_Events_publisher.py")},
		{"expected/python.tornado/variety/f_Events_subscriber.py", filepath.Join(outputDir, "tornado", "variety", "python", "f_Events_subscriber.py")},
		{"expected/python.tornado/variety/f_Foo.py", filepath.Join(outputDir, "tornado", "variety", "python", "f_Foo.py")},

		{"expected/python.tornado/actual_base/constants.py", filepath.Join(outputDir, "tornado", "actual_base", "python", "constants.py")},
		{"expected/python.tornado/actual_base/ttypes.py", filepath.Join(outputDir, "tornado", "actual_base", "python", "ttypes.py")},
		{"expected/python.tornado/actual_base/f_BaseFoo.py", filepath.Join(outputDir, "tornado", "actual_base", "python", "f_BaseFoo.py")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestValidPythonAsyncIO(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:asyncio",
		Out:     outputDir + "/asyncio",
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/python.asyncio/variety/constants.py", filepath.Join(outputDir, "asyncio", "variety", "python", "constants.py")},
		{"expected/python.asyncio/variety/ttypes.py", filepath.Join(outputDir, "asyncio", "variety", "python", "ttypes.py")},
		{"expected/python.asyncio/variety/f_Events_publisher.py", filepath.Join(outputDir, "asyncio", "variety", "python", "f_Events_publisher.py")},
		{"expected/python.asyncio/variety/f_Events_subscriber.py", filepath.Join(outputDir, "asyncio", "variety", "python", "f_Events_subscriber.py")},
		{"expected/python.asyncio/variety/f_Foo.py", filepath.Join(outputDir, "asyncio", "variety", "python", "f_Foo.py")},

		{"expected/python.asyncio/actual_base/constants.py", filepath.Join(outputDir, "asyncio", "actual_base", "python", "constants.py")},
		{"expected/python.asyncio/actual_base/ttypes.py", filepath.Join(outputDir, "asyncio", "actual_base", "python", "ttypes.py")},
		{"expected/python.asyncio/actual_base/f_BaseFoo.py", filepath.Join(outputDir, "asyncio", "actual_base", "python", "f_BaseFoo.py")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestPythonPackagePrefix(t *testing.T) {
	options := compiler.Options{
		File:    "idl/service_inheritance.frugal",
		Gen:     "py:package_prefix=generic_package_prefix.",
		Out:     outputDir,
		Delim:   delim,
		Recurse: false,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/python/package_prefix/f_Foo.py", filepath.Join(outputDir, "service_inheritance", "f_Foo.py")},
		{"expected/python/package_prefix/ttypes.py", filepath.Join(outputDir, "service_inheritance", "ttypes.py")},
		{"expected/python/package_prefix/constants.py", filepath.Join(outputDir, "service_inheritance", "constants.py")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestPythonExtendServiceSameFile(t *testing.T) {
	options := compiler.Options{
		File: "idl/service_extension_same_file.frugal",
		Gen: "py:asyncio",
		Out: outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/python.asyncio/service_extension_same_file/f_BasePinger.py", filepath.Join(outputDir, "service_extension_same_file", "python", "f_BasePinger.py")},
		{"expected/python.asyncio/service_extension_same_file/f_Pinger.py", filepath.Join(outputDir, "service_extension_same_file", "python", "f_Pinger.py")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}
