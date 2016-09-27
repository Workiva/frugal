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

	varietyConstantsPath := filepath.Join(outputDir, "variety", "python", "constants.py")
	compareFiles(t, "expected/python/variety/constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "variety", "python", "ttypes.py")
	compareFiles(t, "expected/python/variety/ttypes.py", varietyFtypesPath)
	eventsPublisherPath := filepath.Join(outputDir, "variety", "python", "f_Events_publisher.py")
	compareFiles(t, "expected/python/variety/f_Events_publisher.py", eventsPublisherPath)
	eventsSubscriberPath := filepath.Join(outputDir, "variety", "python", "f_Events_subscriber.py")
	compareFiles(t, "expected/python/variety/f_Events_subscriber.py", eventsSubscriberPath)
	fooServicePath := filepath.Join(outputDir, "variety", "python", "f_Foo.py")
	compareFiles(t, "expected/python/variety/f_Foo.py", fooServicePath)

	baseConstantsPath := filepath.Join(outputDir, "actual_base", "python", "constants.py")
	compareFiles(t, "expected/python/actual_base/constants.py", baseConstantsPath)
	baseFtypesPath := filepath.Join(outputDir, "actual_base", "python", "ttypes.py")
	compareFiles(t, "expected/python/actual_base/ttypes.py", baseFtypesPath)
	baseFooPath := filepath.Join(outputDir, "actual_base", "python", "f_BaseFoo.py")
	compareFiles(t, "expected/python/actual_base/f_BaseFoo.py", baseFooPath)
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

	varietyConstantsPath := filepath.Join(outputDir, "tornado", "variety", "python", "constants.py")
	compareFiles(t, "expected/python.tornado/variety/constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "tornado", "variety", "python", "ttypes.py")
	compareFiles(t, "expected/python.tornado/variety/ttypes.py", varietyFtypesPath)
	eventsPublisherPath := filepath.Join(outputDir, "tornado", "variety", "python", "f_Events_publisher.py")
	compareFiles(t, "expected/python.tornado/variety/f_Events_publisher.py", eventsPublisherPath)
	eventsSubscriberPath := filepath.Join(outputDir, "tornado", "variety", "python", "f_Events_subscriber.py")
	compareFiles(t, "expected/python.tornado/variety/f_Events_subscriber.py", eventsSubscriberPath)
	fooServicePath := filepath.Join(outputDir, "tornado", "variety", "python", "f_Foo.py")
	compareFiles(t, "expected/python.tornado/variety/f_Foo.py", fooServicePath)

	baseConstantsPath := filepath.Join(outputDir, "tornado", "actual_base", "python", "constants.py")
	compareFiles(t, "expected/python.tornado/actual_base/constants.py", baseConstantsPath)
	baseFtypesPath := filepath.Join(outputDir, "tornado", "actual_base", "python", "ttypes.py")
	compareFiles(t, "expected/python.tornado/actual_base/ttypes.py", baseFtypesPath)
	baseFooPath := filepath.Join(outputDir, "tornado", "actual_base", "python", "f_BaseFoo.py")
	compareFiles(t, "expected/python.tornado/actual_base/f_BaseFoo.py", baseFooPath)
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

	varietyConstantsPath := filepath.Join(outputDir, "asyncio", "variety", "python", "constants.py")
	compareFiles(t, "expected/python.asyncio/variety/constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "asyncio", "variety", "python", "ttypes.py")
	compareFiles(t, "expected/python.asyncio/variety/ttypes.py", varietyFtypesPath)
	eventsPublisherPath := filepath.Join(outputDir, "asyncio", "variety", "python", "f_Events_publisher.py")
	compareFiles(t, "expected/python.asyncio/variety/f_Events_publisher.py", eventsPublisherPath)
	eventsSubscriberPath := filepath.Join(outputDir, "asyncio", "variety", "python", "f_Events_subscriber.py")
	compareFiles(t, "expected/python.asyncio/variety/f_Events_subscriber.py", eventsSubscriberPath)
	fooServicePath := filepath.Join(outputDir, "asyncio", "variety", "python", "f_Foo.py")
	compareFiles(t, "expected/python.asyncio/variety/f_Foo.py", fooServicePath)

	baseConstantsPath := filepath.Join(outputDir, "asyncio", "actual_base", "python", "constants.py")
	compareFiles(t, "expected/python.asyncio/actual_base/constants.py", baseConstantsPath)
	baseFtypesPath := filepath.Join(outputDir, "asyncio", "actual_base", "python", "ttypes.py")
	compareFiles(t, "expected/python.asyncio/actual_base/ttypes.py", baseFtypesPath)
	baseFooPath := filepath.Join(outputDir, "asyncio", "actual_base", "python", "f_BaseFoo.py")
	compareFiles(t, "expected/python.asyncio/actual_base/f_BaseFoo.py", baseFooPath)
}
