package test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestValidJava(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  validFile,
		Gen:   "java:gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "foo", "FooPublisher.java")
	compareFiles(t, "expected/java/FooPublisher.java", pubPath)
	subPath := filepath.Join(outputDir, "foo", "FooSubscriber.java")
	compareFiles(t, "expected/java/FooSubscriber.java", subPath)
	pubPath = filepath.Join(outputDir, "foo", "BlahPublisher.java")
	compareFiles(t, "expected/java/BlahPublisher.java", pubPath)
	subPath = filepath.Join(outputDir, "foo", "BlahSubscriber.java")
	compareFiles(t, "expected/java/BlahSubscriber.java", subPath)
	servicePath := filepath.Join(outputDir, "foo", "FBlah.java")
	compareFiles(t, "expected/java/FBlah.java", servicePath)
}

func TestValidJavaWithAsync(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  validFile,
		Gen:   "java:async,gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	servicePath := filepath.Join(outputDir, "foo", "FBlah.java")
	compareFiles(t, "expected/java/FBlah_async.java", servicePath)
}

func TestValidJavaFrugalCompiler(t *testing.T) {
	defer globals.Reset()
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "java:gen_with_frugal=true",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	awesomeExceptionPath := filepath.Join(outputDir, "variety", "java", "AwesomeException.java")
	compareFiles(t, "expected/java/variety/AwesomeException.java", awesomeExceptionPath)
	eventPath := filepath.Join(outputDir, "variety", "java", "Event.java")
	compareFiles(t, "expected/java/variety/Event.java", eventPath)
	eventWrapperPath := filepath.Join(outputDir, "variety", "java", "EventWrapper.java")
	compareFiles(t, "expected/java/variety/EventWrapper.java", eventWrapperPath)
	fooPath := filepath.Join(outputDir, "variety", "java", "Foo.java")
	compareFiles(t, "expected/java/variety/Foo.java", fooPath)
	itsAnEnumPath := filepath.Join(outputDir, "variety", "java", "ItsAnEnum.java")
	compareFiles(t, "expected/java/variety/ItsAnEnum.java", itsAnEnumPath)
	testBasePath := filepath.Join(outputDir, "variety", "java", "TestBase.java")
	compareFiles(t, "expected/java/variety/TestBase.java", testBasePath)
	testingDefaultsPath := filepath.Join(outputDir, "variety", "java", "TestingDefaults.java")
	compareFiles(t, "expected/java/variety/TestingDefaults.java", testingDefaultsPath)
	testingUnionsPath := filepath.Join(outputDir, "variety", "java", "TestingUnions.java")
	compareFiles(t, "expected/java/variety/TestingUnions.java", testingUnionsPath)
	varietyConstantsPath := filepath.Join(outputDir, "variety", "java", "varietyConstants.java")
	compareFiles(t, "expected/java/variety/varietyConstants.java", varietyConstantsPath)

	apiExceptionPath := filepath.Join(outputDir, "actual_base", "java", "api_exception.java")
	compareFiles(t, "expected/java/actual_base/api_exception.java", apiExceptionPath)
	baseConstantsPath := filepath.Join(outputDir, "actual_base", "java", "baseConstants.java")
	compareFiles(t, "expected/java/actual_base/baseConstants.java", baseConstantsPath)
	baseFooPath := filepath.Join(outputDir, "actual_base", "java", "BaseFoo.java")
	compareFiles(t, "expected/java/actual_base/BaseFoo.java", baseFooPath)
	thingPath := filepath.Join(outputDir, "actual_base", "java", "thing.java")
	compareFiles(t, "expected/java/actual_base/thing.java", thingPath)
	nestedThingPath := filepath.Join(outputDir, "actual_base", "java", "nested_thing.java")
	compareFiles(t, "expected/java/actual_base/nested_thing.java", nestedThingPath)
}
