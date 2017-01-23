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


func TestValidPythonGevent(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:gevent",
		Out:     outputDir + "/gevent",
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	varietyConstantsPath := filepath.Join(outputDir, "gevent", "variety", "python", "constants.py")
	compareFiles(t, "expected/python.gevent/variety/constants.py", varietyConstantsPath)
	varietyFtypesPath := filepath.Join(outputDir, "gevent", "variety", "python", "ttypes.py")
	compareFiles(t, "expected/python.gevent/variety/ttypes.py", varietyFtypesPath)
	eventsPublisherPath := filepath.Join(outputDir, "gevent", "variety", "python", "f_Events_publisher.py")
	compareFiles(t, "expected/python.gevent/variety/f_Events_publisher.py", eventsPublisherPath)
	eventsSubscriberPath := filepath.Join(outputDir, "gevent", "variety", "python", "f_Events_subscriber.py")
	compareFiles(t, "expected/python.gevent/variety/f_Events_subscriber.py", eventsSubscriberPath)
	fooServicePath := filepath.Join(outputDir, "gevent", "variety", "python", "f_Foo.py")
	compareFiles(t, "expected/python.gevent/variety/f_Foo.py", fooServicePath)

	baseConstantsPath := filepath.Join(outputDir, "gevent", "actual_base", "python", "constants.py")
	compareFiles(t, "expected/python.gevent/actual_base/constants.py", baseConstantsPath)
	baseFtypesPath := filepath.Join(outputDir, "gevent", "actual_base", "python", "ttypes.py")
	compareFiles(t, "expected/python.gevent/actual_base/ttypes.py", baseFtypesPath)
	baseFooPath := filepath.Join(outputDir, "gevent", "actual_base", "python", "f_BaseFoo.py")
	compareFiles(t, "expected/python.gevent/actual_base/f_BaseFoo.py", baseFooPath)
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

	fFooPath := filepath.Join(outputDir, "service_inheritance", "f_Foo.py")
	compareFiles(t, "expected/python/package_prefix/f_Foo.py", fFooPath)
	ttypesPath := filepath.Join(outputDir, "service_inheritance", "ttypes.py")
	compareFiles(t, "expected/python/package_prefix/ttypes.py", ttypesPath)
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

	basePingerPath := filepath.Join(outputDir, "service_extension_same_file", "python", "f_BasePinger.py")
	compareFiles(t, "expected/python.asyncio/service_extension_same_file/f_BasePinger.py", basePingerPath)
	pingerPath := filepath.Join(outputDir, "service_extension_same_file", "python", "f_Pinger.py")
	compareFiles(t, "expected/python.asyncio/service_extension_same_file/f_Pinger.py", pingerPath)
}
