package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidDartFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	awesomeExceptionPath := filepath.Join(outputDir, "variety", "lib", "src", "f_awesome_exception.dart")
	compareFiles(t, "expected/dart/variety/f_awesome_exception.dart", awesomeExceptionPath)
	eventPath := filepath.Join(outputDir, "variety", "lib", "src", "f_event.dart")
	compareFiles(t, "expected/dart/variety/f_event.dart", eventPath)
	eventWrapperPath := filepath.Join(outputDir, "variety", "lib", "src", "f_event_wrapper.dart")
	compareFiles(t, "expected/dart/variety/f_event_wrapper.dart", eventWrapperPath)
	itsAnEnumPath := filepath.Join(outputDir, "variety", "lib", "src", "f_its_an_enum.dart")
	compareFiles(t, "expected/dart/variety/f_its_an_enum.dart", itsAnEnumPath)
	testBasePath := filepath.Join(outputDir, "variety", "lib", "src", "f_test_base.dart")
	compareFiles(t, "expected/dart/variety/f_test_base.dart", testBasePath)
	testingDefaultsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_testing_defaults.dart")
	compareFiles(t, "expected/dart/variety/f_testing_defaults.dart", testingDefaultsPath)
	testingUnionsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_testing_unions.dart")
	compareFiles(t, "expected/dart/variety/f_testing_unions.dart", testingUnionsPath)
	healthConditionPath := filepath.Join(outputDir, "variety", "lib", "src", "f_health_condition.dart")
	compareFiles(t, "expected/dart/variety/f_health_condition.dart", healthConditionPath)
	testLowercasePath := filepath.Join(outputDir, "variety", "lib", "src", "f_test_lowercase.dart")
	compareFiles(t, "expected/dart/variety/f_test_lowercase.dart", testLowercasePath)
	varietyConstantsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_variety_constants.dart")
	compareFiles(t, "expected/dart/variety/f_variety_constants.dart", varietyConstantsPath)
	eventsScopePath := filepath.Join(outputDir, "variety", "lib", "src", "f_events_scope.dart")
	compareFiles(t, "expected/dart/variety/f_events_scope.dart", eventsScopePath)
	fooServicePath := filepath.Join(outputDir, "variety", "lib", "src", "f_foo_service.dart")
	compareFiles(t, "expected/dart/variety/f_foo_service.dart", fooServicePath)
	varietyExportPath := filepath.Join(outputDir, "variety", "lib", "variety.dart")
	compareFiles(t, "expected/dart/variety/variety.dart", varietyExportPath)

	actualBaseConstantsPath := filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_actual_base_dart_constants.dart")
	compareFiles(t, "expected/dart/actual_base/f_actual_base_dart_constants.dart", actualBaseConstantsPath)
	apiExceptionPath := filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_api_exception.dart")
	compareFiles(t, "expected/dart/actual_base/f_api_exception.dart", apiExceptionPath)
	thingPath := filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_thing.dart")
	compareFiles(t, "expected/dart/actual_base/f_thing.dart", thingPath)
	baseHealthConditionPath := filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_base_health_condition.dart")
	compareFiles(t, "expected/dart/actual_base/f_base_health_condition.dart", baseHealthConditionPath)
	baseFooServicePath := filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_base_foo_service.dart")
	compareFiles(t, "expected/dart/actual_base/f_base_foo_service.dart", baseFooServicePath)
	actualBaseExportPath := filepath.Join(outputDir, "actual_base_dart", "lib", "actual_base_dart.dart")
	compareFiles(t, "expected/dart/actual_base/actual_base_dart.dart", actualBaseExportPath)
}
