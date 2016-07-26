package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidDart(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "dart:gen_with_frugal=false",
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

func TestValidDartFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart:gen_with_frugal=true",
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
	fooStructsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_foo_structs.dart")
	compareFiles(t, "expected/dart/variety/f_foo_structs.dart", fooStructsPath)
	itsAnEnumPath := filepath.Join(outputDir, "variety", "lib", "src", "f_its_an_enum.dart")
	compareFiles(t, "expected/dart/variety/f_its_an_enum.dart", itsAnEnumPath)
	testBasePath := filepath.Join(outputDir, "variety", "lib", "src", "f_test_base.dart")
	compareFiles(t, "expected/dart/variety/f_test_base.dart", testBasePath)
	testingDefaultsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_testing_defaults.dart")
	compareFiles(t, "expected/dart/variety/f_testing_defaults.dart", testingDefaultsPath)
	testingUnionsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_testing_unions.dart")
	compareFiles(t, "expected/dart/variety/f_testing_unions.dart", testingUnionsPath)
	varietyConstantsPath := filepath.Join(outputDir, "variety", "lib", "src", "f_variety_constants.dart")
	compareFiles(t, "expected/dart/variety/f_variety_constants.dart", varietyConstantsPath)
	varietyExportPath := filepath.Join(outputDir, "variety", "lib", "variety.dart")
	compareFiles(t, "expected/dart/variety/variety.dart", varietyExportPath)

	actualBaseConstantsPath := filepath.Join(outputDir, "actual_base", "lib", "src", "f_actual_base_constants.dart")
	compareFiles(t, "expected/dart/actual_base/f_actual_base_constants.dart", actualBaseConstantsPath)
	apiExceptionPath := filepath.Join(outputDir, "actual_base", "lib", "src", "f_api_exception.dart")
	compareFiles(t, "expected/dart/actual_base/f_api_exception.dart", apiExceptionPath)
	baseFooStructs := filepath.Join(outputDir, "actual_base", "lib", "src", "f_base_foo_structs.dart")
	compareFiles(t, "expected/dart/actual_base/f_base_foo_structs.dart", baseFooStructs)
	thingPath := filepath.Join(outputDir, "actual_base", "lib", "src", "f_thing.dart")
	compareFiles(t, "expected/dart/actual_base/f_thing.dart", thingPath)
	actualBaseExportPath := filepath.Join(outputDir, "actual_base", "lib", "actual_base.dart")
	compareFiles(t, "expected/dart/actual_base/actual_base.dart", actualBaseExportPath)
}
