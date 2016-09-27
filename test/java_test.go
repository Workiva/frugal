package test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestValidJavaWithAsync(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File: frugalGenFile,
		Gen: "java:async",
		Out: outputDir + "/async",
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	fooServicePath := filepath.Join(outputDir, "async", "variety", "java", "FFoo.java")
	compareFiles(t, "expected/java/variety_async/FFoo.java", fooServicePath)
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
		Gen:     "java",
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
	itsAnEnumPath := filepath.Join(outputDir, "variety", "java", "ItsAnEnum.java")
	compareFiles(t, "expected/java/variety/ItsAnEnum.java", itsAnEnumPath)
	testBasePath := filepath.Join(outputDir, "variety", "java", "TestBase.java")
	compareFiles(t, "expected/java/variety/TestBase.java", testBasePath)
	testingDefaultsPath := filepath.Join(outputDir, "variety", "java", "TestingDefaults.java")
	compareFiles(t, "expected/java/variety/TestingDefaults.java", testingDefaultsPath)
	testingUnionsPath := filepath.Join(outputDir, "variety", "java", "TestingUnions.java")
	compareFiles(t, "expected/java/variety/TestingUnions.java", testingUnionsPath)
	healthConditionPath := filepath.Join(outputDir, "variety", "java", "HealthCondition.java")
	compareFiles(t, "expected/java/variety/HealthCondition.java", healthConditionPath)
	varietyConstantsPath := filepath.Join(outputDir, "variety", "java", "varietyConstants.java")
	compareFiles(t, "expected/java/variety/varietyConstants.java", varietyConstantsPath)
	testLowercasePath := filepath.Join(outputDir, "variety", "java", "TestLowercase.java")
	compareFiles(t, "expected/java/variety/TestLowercase.java", testLowercasePath)
	eventsPublisherPath := filepath.Join(outputDir, "variety", "java", "EventsPublisher.java")
	compareFiles(t, "expected/java/variety/EventsPublisher.java", eventsPublisherPath)
	eventsSubscriberPath := filepath.Join(outputDir, "variety", "java", "EventsSubscriber.java")
	compareFiles(t, "expected/java/variety/EventsSubscriber.java", eventsSubscriberPath)
	fooPath := filepath.Join(outputDir, "variety", "java", "FFoo.java")
	compareFiles(t, "expected/java/variety/FFoo.java", fooPath)

	apiExceptionPath := filepath.Join(outputDir, "actual_base", "java", "api_exception.java")
	compareFiles(t, "expected/java/actual_base/api_exception.java", apiExceptionPath)
	baseConstantsPath := filepath.Join(outputDir, "actual_base", "java", "baseConstants.java")
	compareFiles(t, "expected/java/actual_base/baseConstants.java", baseConstantsPath)
	thingPath := filepath.Join(outputDir, "actual_base", "java", "thing.java")
	compareFiles(t, "expected/java/actual_base/thing.java", thingPath)
	baseHealthConditionPath := filepath.Join(outputDir, "actual_base", "java", "base_health_condition.java")
	compareFiles(t, "expected/java/actual_base/base_health_condition.java", baseHealthConditionPath)
	baseFooPath := filepath.Join(outputDir, "actual_base", "java", "FBaseFoo.java")
	compareFiles(t, "expected/java/actual_base/FBaseFoo.java", baseFooPath)
}
