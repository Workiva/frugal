/*
 * Copyright 2021 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package compiler_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestCircularIncludes(t *testing.T) {
	options := compiler.Options{
		File:   idl("circular_1.frugal"),
		Gen:    "go",
		Out:    "out",
		Delim:  ".",
		DryRun: true,
	}
	err := compiler.Compile(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Circular include: [circular_1 circular_2 circular_3 circular_1]")
}

func TestValidDartFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"dart/variety/f_awesome_exception.dart", "variety/lib/src/f_awesome_exception.dart"},
		{"dart/variety/f_event.dart", "variety/lib/src/f_event.dart"},
		{"dart/variety/f_event_wrapper.dart", "variety/lib/src/f_event_wrapper.dart"},
		{"dart/variety/f_its_an_enum.dart", "variety/lib/src/f_its_an_enum.dart"},
		{"dart/variety/f_test_base.dart", "variety/lib/src/f_test_base.dart"},
		{"dart/variety/f_testing_defaults.dart", "variety/lib/src/f_testing_defaults.dart"},
		{"dart/variety/f_testing_unions.dart", "variety/lib/src/f_testing_unions.dart"},
		{"dart/variety/f_health_condition.dart", "variety/lib/src/f_health_condition.dart"},
		{"dart/variety/f_test_lowercase.dart", "variety/lib/src/f_test_lowercase.dart"},
		{"dart/variety/f_foo_args.dart", "variety/lib/src/f_foo_args.dart"},
		{"dart/variety/f_variety_constants.dart", "variety/lib/src/f_variety_constants.dart"},
		{"dart/variety/f_events_scope.dart", "variety/lib/src/f_events_scope.dart"},
		{"dart/variety/f_foo_service.dart", "variety/lib/src/f_foo_service.dart"},
		{"dart/variety/variety.dart", "variety/lib/variety.dart"},

		{"dart/actual_base/f_actual_base_dart_constants.dart", "actual_base_dart/lib/src/f_actual_base_dart_constants.dart"},
		{"dart/actual_base/f_api_exception.dart", "actual_base_dart/lib/src/f_api_exception.dart"},
		{"dart/actual_base/f_thing.dart", "actual_base_dart/lib/src/f_thing.dart"},
		{"dart/actual_base/f_base_health_condition.dart", "actual_base_dart/lib/src/f_base_health_condition.dart"},
		{"dart/actual_base/f_base_foo_service.dart", "actual_base_dart/lib/src/f_base_foo_service.dart"},
		{"dart/actual_base/f_nested_thing.dart", "actual_base_dart/lib/src/f_nested_thing.dart"},
		{"dart/actual_base/actual_base_dart.dart", "actual_base_dart/lib/actual_base_dart.dart"},
	}
	suite.Run(t, options)
}

func TestValidDartNullsafe(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart:nullsafe",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"dart.nullsafe/variety/f_awesome_exception.dart", "variety/lib/src/f_awesome_exception.dart"},
		{"dart.nullsafe/variety/f_event.dart", "variety/lib/src/f_event.dart"},
		{"dart.nullsafe/variety/f_event_wrapper.dart", "variety/lib/src/f_event_wrapper.dart"},
		{"dart.nullsafe/variety/f_its_an_enum.dart", "variety/lib/src/f_its_an_enum.dart"},
		{"dart.nullsafe/variety/f_test_base.dart", "variety/lib/src/f_test_base.dart"},
		{"dart.nullsafe/variety/f_testing_defaults.dart", "variety/lib/src/f_testing_defaults.dart"},
		{"dart.nullsafe/variety/f_testing_unions.dart", "variety/lib/src/f_testing_unions.dart"},
		{"dart.nullsafe/variety/f_health_condition.dart", "variety/lib/src/f_health_condition.dart"},
		{"dart.nullsafe/variety/f_test_lowercase.dart", "variety/lib/src/f_test_lowercase.dart"},
		{"dart.nullsafe/variety/f_foo_args.dart", "variety/lib/src/f_foo_args.dart"},
		{"dart.nullsafe/variety/f_variety_constants.dart", "variety/lib/src/f_variety_constants.dart"},
		{"dart.nullsafe/variety/f_events_scope.dart", "variety/lib/src/f_events_scope.dart"},
		{"dart.nullsafe/variety/f_foo_service.dart", "variety/lib/src/f_foo_service.dart"},
		{"dart.nullsafe/variety/variety.dart", "variety/lib/variety.dart"},
		{"dart.nullsafe/variety/pubspec.yaml", "variety/pubspec.yaml"},

		{"dart.nullsafe/actual_base/f_actual_base_dart_constants.dart", "actual_base_dart/lib/src/f_actual_base_dart_constants.dart"},
		{"dart.nullsafe/actual_base/f_api_exception.dart", "actual_base_dart/lib/src/f_api_exception.dart"},
		{"dart.nullsafe/actual_base/f_thing.dart", "actual_base_dart/lib/src/f_thing.dart"},
		{"dart.nullsafe/actual_base/f_base_health_condition.dart", "actual_base_dart/lib/src/f_base_health_condition.dart"},
		{"dart.nullsafe/actual_base/f_base_foo_service.dart", "actual_base_dart/lib/src/f_base_foo_service.dart"},
		{"dart.nullsafe/actual_base/f_nested_thing.dart", "actual_base_dart/lib/src/f_nested_thing.dart"},
		{"dart.nullsafe/actual_base/actual_base_dart.dart", "actual_base_dart/lib/actual_base_dart.dart"},
		{"dart.nullsafe/actual_base/pubspec.yaml", "actual_base_dart/pubspec.yaml"},
	}
	suite.Run(t, options)
}

func TestValidDartUseInt64(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart:use_int64",
		Out:     outputDir + "/int64",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"dart.int64/variety/f_awesome_exception.dart", "int64/variety/lib/src/f_awesome_exception.dart"},
		{"dart.int64/variety/f_event.dart", "int64/variety/lib/src/f_event.dart"},
		{"dart.int64/variety/f_event_wrapper.dart", "int64/variety/lib/src/f_event_wrapper.dart"},
		{"dart.int64/variety/f_its_an_enum.dart", "int64/variety/lib/src/f_its_an_enum.dart"},
		{"dart.int64/variety/f_test_base.dart", "int64/variety/lib/src/f_test_base.dart"},
		{"dart.int64/variety/f_testing_defaults.dart", "int64/variety/lib/src/f_testing_defaults.dart"},
		{"dart.int64/variety/f_testing_unions.dart", "int64/variety/lib/src/f_testing_unions.dart"},
		{"dart.int64/variety/f_health_condition.dart", "int64/variety/lib/src/f_health_condition.dart"},
		{"dart.int64/variety/f_test_lowercase.dart", "int64/variety/lib/src/f_test_lowercase.dart"},
		{"dart.int64/variety/f_foo_args.dart", "int64/variety/lib/src/f_foo_args.dart"},
		{"dart.int64/variety/f_variety_constants.dart", "int64/variety/lib/src/f_variety_constants.dart"},
		{"dart.int64/variety/f_events_scope.dart", "int64/variety/lib/src/f_events_scope.dart"},
		{"dart.int64/variety/f_foo_service.dart", "int64/variety/lib/src/f_foo_service.dart"},
		{"dart.int64/variety/variety.dart", "int64/variety/lib/variety.dart"},

		{"dart.int64/actual_base/f_actual_base_dart_constants.dart", "int64/actual_base_dart/lib/src/f_actual_base_dart_constants.dart"},
		{"dart.int64/actual_base/f_api_exception.dart", "int64/actual_base_dart/lib/src/f_api_exception.dart"},
		{"dart.int64/actual_base/f_thing.dart", "int64/actual_base_dart/lib/src/f_thing.dart"},
		{"dart.int64/actual_base/f_base_health_condition.dart", "int64/actual_base_dart/lib/src/f_base_health_condition.dart"},
		{"dart.int64/actual_base/f_base_foo_service.dart", "int64/actual_base_dart/lib/src/f_base_foo_service.dart"},
		{"dart.int64/actual_base/f_nested_thing.dart", "int64/actual_base_dart/lib/src/f_nested_thing.dart"},
		{"dart.int64/actual_base/actual_base_dart.dart", "int64/actual_base_dart/lib/actual_base_dart.dart"},
	}
	suite.Run(t, options)
}

func TestValidDartUseNullForUnset(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "dart:use_null_for_unset",
		Out:     outputDir + "/nullunset",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"dart.nullunset/variety/f_awesome_exception.dart", "nullunset/variety/lib/src/f_awesome_exception.dart"},
		{"dart.nullunset/variety/f_event.dart", "nullunset/variety/lib/src/f_event.dart"},
		{"dart.nullunset/variety/f_event_wrapper.dart", "nullunset/variety/lib/src/f_event_wrapper.dart"},
		{"dart.nullunset/variety/f_its_an_enum.dart", "nullunset/variety/lib/src/f_its_an_enum.dart"},
		{"dart.nullunset/variety/f_test_base.dart", "nullunset/variety/lib/src/f_test_base.dart"},
		{"dart.nullunset/variety/f_testing_defaults.dart", "nullunset/variety/lib/src/f_testing_defaults.dart"},
		{"dart.nullunset/variety/f_testing_unions.dart", "nullunset/variety/lib/src/f_testing_unions.dart"},
		{"dart.nullunset/variety/f_health_condition.dart", "nullunset/variety/lib/src/f_health_condition.dart"},
		{"dart.nullunset/variety/f_test_lowercase.dart", "nullunset/variety/lib/src/f_test_lowercase.dart"},
		{"dart.nullunset/variety/f_foo_args.dart", "nullunset/variety/lib/src/f_foo_args.dart"},
		{"dart.nullunset/variety/f_variety_constants.dart", "nullunset/variety/lib/src/f_variety_constants.dart"},
		{"dart.nullunset/variety/f_events_scope.dart", "nullunset/variety/lib/src/f_events_scope.dart"},
		{"dart.nullunset/variety/f_foo_service.dart", "nullunset/variety/lib/src/f_foo_service.dart"},
		{"dart.nullunset/variety/variety.dart", "nullunset/variety/lib/variety.dart"},

		{"dart.nullunset/actual_base/f_actual_base_dart_constants.dart", "nullunset/actual_base_dart/lib/src/f_actual_base_dart_constants.dart"},
		{"dart.nullunset/actual_base/f_api_exception.dart", "nullunset/actual_base_dart/lib/src/f_api_exception.dart"},
		{"dart.nullunset/actual_base/f_thing.dart", "nullunset/actual_base_dart/lib/src/f_thing.dart"},
		{"dart.nullunset/actual_base/f_base_health_condition.dart", "nullunset/actual_base_dart/lib/src/f_base_health_condition.dart"},
		{"dart.nullunset/actual_base/f_base_foo_service.dart", "nullunset/actual_base_dart/lib/src/f_base_foo_service.dart"},
		{"dart.nullunset/actual_base/f_nested_thing.dart", "nullunset/actual_base_dart/lib/src/f_nested_thing.dart"},
		{"dart.nullunset/actual_base/actual_base_dart.dart", "nullunset/actual_base_dart/lib/actual_base_dart.dart"},
	}
	suite.Run(t, options)
}

func TestValidDartEnums(t *testing.T) {
	options := compiler.Options{
		File:    idl("enum.frugal"),
		Gen:     "dart:use_enums",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"dart/enum/f_testing_enums.dart", "enum_dart/lib/src/f_testing_enums.dart"},
		{"dart/enum/enum_dart.dart", "enum_dart/lib/enum_dart.dart"},
	}
	suite.Run(t, options)
}

// Ensures correct import references are used when -use-vendor is set and the
// IDL has a vendored include
func TestValidDartVendor(t *testing.T) {
	options := compiler.Options{
		File:  includeVendor,
		Gen:   "dart:use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	suite := ComparisonList{
		{"dart/include_vendor/f_my_scope_scope.dart", "include_vendor/lib/src/f_my_scope_scope.dart"},
		{"dart/include_vendor/f_my_service_service.dart", "include_vendor/lib/src/f_my_service_service.dart"},
		{"dart/include_vendor/f_vendored_references.dart", "include_vendor/lib/src/f_vendored_references.dart"},
		{"dart/include_vendor/include_vendor.dart", "include_vendor/lib/include_vendor.dart"},
		{"dart/include_vendor/pubspec.yaml", "include_vendor/pubspec.yaml"},
	}
	suite.Run(t, options)
}

// Ensures an error is returned when -use-vendor is set and the vendored
// include does not specify a path.
func TestValidDartVendorPathNotSpecified(t *testing.T) {
	options := compiler.Options{
		File:  includeVendorNoPath,
		Gen:   "dart:use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	err := compiler.Compile(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not specify vendor path for dart namespace")
}

// Ensures the target IDL is generated when -use-vendor is set and it has a
// vendored namespace.
func TestValidDartVendorNamespaceTargetGenerate(t *testing.T) {
	options := compiler.Options{
		File:  vendorNamespace,
		Gen:   "dart:use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	suite := ComparisonList{
		{"dart/vendor_namespace/vendor_namespace.dart", "vendor_namespace/lib/vendor_namespace.dart"},
		{"dart/vendor_namespace/f_item.dart", "vendor_namespace/lib/src/f_item.dart"},
		{"dart/vendor_namespace/f_vendored_base_service.dart", "vendor_namespace/lib/src/f_vendored_base_service.dart"},
		{"dart/vendor_namespace/f_vendor_namespace_constants.dart", "vendor_namespace/lib/src/f_vendor_namespace_constants.dart"},
		{"dart/vendor_namespace/f_my_enum.dart", "vendor_namespace/lib/src/f_my_enum.dart"},
	}
	suite.Run(t, options)
}

func TestValidGoWithAsync(t *testing.T) {
	options := compiler.Options{
		File:  frugalGenFile,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/async/,async",
		Out:   outputDir + "/async",
		Delim: delim,
	}
	suite := ComparisonList{
		{"go/variety_async/f_foo_service.txt", "async/variety/f_foo_service.go"},
	}
	suite.Run(t, options)
}

func TestValidGoFrugalCompiler(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"go/actual_base/f_types.txt", "actual_base/golang/f_types.go"},
		{"go/actual_base/f_basefoo_service.txt", "actual_base/golang/f_basefoo_service.go"},

		{"go/variety/f_types.txt", "variety/f_types.go"},
		{"go/variety/f_foo_service.txt", "variety/f_foo_service.go"},
		{"go/variety/f_events_scope.txt", "variety/f_events_scope.go"},
	}
	suite.Run(t, options)
}

// Ensures correct import references are used when -use-vendor is set and the
// IDL has a vendored include.
func TestValidGoVendor(t *testing.T) {
	options := compiler.Options{
		File:  includeVendor,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	suite := ComparisonList{
		{"go/vendor/f_myscope_scope.txt", "include_vendor/f_myscope_scope.go"},
		{"go/vendor/f_myservice_service.txt", "include_vendor/f_myservice_service.go"},
		{"go/vendor/f_types.txt", "include_vendor/f_types.go"},
	}
	suite.Run(t, options)
}

// Ensures an error is returned when -use-vendor is set and the vendored
// include does not specify a path.
func TestValidGoVendorPathNotSpecified(t *testing.T) {
	options := compiler.Options{
		File:  includeVendorNoPath,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	err := compiler.Compile(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not specify vendor path for go namespace")
}

// Ensures the target IDL is generated when -use-vendor is set and it has a
// vendored namespace.
func TestValidGoVendorNamespaceTargetGenerate(t *testing.T) {
	options := compiler.Options{
		File:  vendorNamespace,
		Gen:   "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/,use_vendor",
		Out:   outputDir,
		Delim: delim,
	}
	suite := ComparisonList{
		{"go/vendor_namespace/f_types.txt", "vendor_namespace/f_types.go"},
		{"go/vendor_namespace/f_vendoredbase_service.txt", "vendor_namespace/f_vendoredbase_service.go"},
	}
	suite.Run(t, options)
}

// Ensures includes are generated in the same order
func TestIncludeOrdering(t *testing.T) {
	options := compiler.Options{
		File:    idl("ordering/main.frugal"),
		Gen:     "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/ordering",
		Out:     gen("ordering"),
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"go/ordering/one/f_types.go", "ordering/one/f_types.go"},
		{"go/ordering/two/f_types.go", "ordering/two/f_types.go"},
		{"go/ordering/three/f_types.go", "ordering/three/f_types.go"},
		{"go/ordering/four/f_types.go", "ordering/four/f_types.go"},
		{"go/ordering/five/f_types.go", "ordering/five/f_types.go"},
	}
	suite.Run(t, options)
}

// Ensures slim generated files are correct.
func TestSlim(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/,slim",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"go/slim/f_types.go", "variety/f_types.go"},
		{"go/slim/f_foo_service.go", "variety/f_foo_service.go"},
		{"go/slim/f_events_scope.go", "variety/f_events_scope.go"},
	}
	suite.Run(t, options)
}

// Ensures deprecated logging can be suppressed
func TestSuppressedDeprecatedLogging(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "go:package_prefix=github.com/Workiva/frugal/compiler/testdata/out/,suppress_deprecated_logging",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"go/deprecated_logging/f_foo_service.go", "variety/f_foo_service.go"},
	}
	suite.Run(t, options)
}

func TestHTML(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "html",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"html/style.css", "style.css"},
		{"html/index.html", "index.html"},
		{"html/base.html", "base.html"},
		{"html/variety.html", "variety.html"},
	}
	suite.Run(t, options)
}

func TestHTMLStandalone(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "html:standalone",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"html/standalone/index.html", "index.html"},
		{"html/standalone/base.html", "base.html"},
		{"html/standalone/variety.html", "variety.html"},
	}
	suite.Run(t, options)
}

func TestInvalid(t *testing.T) {
	suite := [][2]string{
		{"invalid.frugal", "parser: syntax error"},
		{"duplicate_services.frugal", "Services foo and Foo conflict"},
		{"duplicate_scopes.frugal", "Scopes foo and Foo conflict"},
		{"duplicate_methods.frugal", "Methods Ping and ping conflict"},
		{"duplicate_operations.frugal", "Operations boo and Boo conflict"},
		{"duplicate_arg_ids.frugal", "Duplicate field"},
		{"duplicate_field_ids.frugal", "Duplicate field"},
		{"bad_namespace.frugal", "annotation not compatible with * namespace"},
		{"bad_op_type.frugal", "Invalid operation type invalid.type for test.test"},
	}
	for _, test := range suite {
		options := compiler.Options{
			File:  idl(test[0]),
			Gen:   "go",
			Out:   outputDir,
			Delim: delim,
		}
		err := compiler.Compile(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), test[1])
	}
}

func TestValidJavaWithAsync(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  frugalGenFile,
		Gen:   "java:async",
		Out:   outputDir + "/async",
		Delim: delim,
	}
	suite := ComparisonList{
		{"java/variety_async/FFoo.java", "async/variety/java/FFoo.java"},
	}
	suite.Run(t, options)
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
	suite := ComparisonList{
		{"java/variety/AwesomeException.java", "variety/java/AwesomeException.java"},
		{"java/variety/Event.java", "variety/java/Event.java"},
		{"java/variety/EventWrapper.java", "variety/java/EventWrapper.java"},
		{"java/variety/ItsAnEnum.java", "variety/java/ItsAnEnum.java"},
		{"java/variety/TestBase.java", "variety/java/TestBase.java"},
		{"java/variety/TestingDefaults.java", "variety/java/TestingDefaults.java"},
		{"java/variety/TestingUnions.java", "variety/java/TestingUnions.java"},
		{"java/variety/HealthCondition.java", "variety/java/HealthCondition.java"},
		{"java/variety/varietyConstants.java", "variety/java/varietyConstants.java"},
		{"java/variety/TestLowercase.java", "variety/java/TestLowercase.java"},
		{"java/variety/FooArgs.java", "variety/java/FooArgs.java"},
		{"java/variety/EventsPublisher.java", "variety/java/EventsPublisher.java"},
		{"java/variety/EventsSubscriber.java", "variety/java/EventsSubscriber.java"},
		{"java/variety/FFoo.java", "variety/java/FFoo.java"},

		{"java/actual_base/api_exception.java", "actual_base/java/api_exception.java"},
		{"java/actual_base/baseConstants.java", "actual_base/java/baseConstants.java"},
		{"java/actual_base/thing.java", "actual_base/java/thing.java"},
		{"java/actual_base/base_health_condition.java", "actual_base/java/base_health_condition.java"},
		{"java/actual_base/FBaseFoo.java", "actual_base/java/FBaseFoo.java"},
		{"java/actual_base/nested_thing.java", "actual_base/java/nested_thing.java"},
	}
	suite.Run(t, options)
}

func TestValidJavaBoxedPrimitives(t *testing.T) {
	defer globals.Reset()
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "java:boxed_primitives",
		Out:     outputDir + "/boxed_primitives",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"java/boxed_primitives/FFoo.java", "boxed_primitives/variety/java/FFoo.java"},
		{"java/boxed_primitives/TestingDefaults.java", "boxed_primitives/variety/java/TestingDefaults.java"},
	}
	suite.Run(t, options)
}

// Ensures correct import references are used when -use-vendor is set and the
// IDL has a vendored include.
func TestValidJavaVendor(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    includeVendor,
		Gen:     "java:use_vendor",
		Out:     outputDir + "/valid_vendor",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"java/valid_vendor/FMyService.java", "valid_vendor/include_vendor/java/FMyService.java"},
		{"java/valid_vendor/MyScopePublisher.java", "valid_vendor/include_vendor/java/MyScopePublisher.java"},
		{"java/valid_vendor/MyScopeSubscriber.java", "valid_vendor/include_vendor/java/MyScopeSubscriber.java"},
		{"java/valid_vendor/VendoredReferences.java", "valid_vendor/include_vendor/java/VendoredReferences.java"},
		{"java/valid_vendor/InvalidData.java", "valid_vendor/InvalidData.java"},
	}
	suite.Run(t, options)

	filesNotToGenerate := []string{
		"valid_vendor/vendor_namespace/java/Item.java",
		"valid_vendor/vendor_namespace/java/vendor_namespaceConstants.java",
		"valid_vendor/vendor_namespace/java/MyEnum.java",
		"valid_vendor/vendor_namespace/java/FVendoredBase.java",
	}
	assertFilesNotExist(t, filesNotToGenerate)
}

func TestValidJavaVendorButNotUseVendor(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    includeVendor,
		Gen:     "java",
		Out:     outputDir + "/vendored_but_no_use_vendor",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"java/vendored_but_no_use_vendor/FMyService.java", "vendored_but_no_use_vendor/include_vendor/java/FMyService.java"},
		{"java/vendored_but_no_use_vendor/MyScopePublisher.java", "vendored_but_no_use_vendor/include_vendor/java/MyScopePublisher.java"},
		{"java/vendored_but_no_use_vendor/MyScopeSubscriber.java", "vendored_but_no_use_vendor/include_vendor/java/MyScopeSubscriber.java"},
		{"java/vendored_but_no_use_vendor/VendoredReferences.java", "vendored_but_no_use_vendor/include_vendor/java/VendoredReferences.java"},
		{"java/vendored_but_no_use_vendor/InvalidData.java", "vendored_but_no_use_vendor/InvalidData.java"},
		{"java/vendored_but_no_use_vendor/Item.java", "vendored_but_no_use_vendor/vendor_namespace/java/Item.java"},
		{"java/vendored_but_no_use_vendor/vendor_namespaceConstants.java", "vendored_but_no_use_vendor/vendor_namespace/java/vendor_namespaceConstants.java"},
		{"java/vendored_but_no_use_vendor/MyEnum.java", "vendored_but_no_use_vendor/vendor_namespace/java/MyEnum.java"},
		{"java/vendored_but_no_use_vendor/FVendoredBase.java", "vendored_but_no_use_vendor/vendor_namespace/java/FVendoredBase.java"},
	}
	suite.Run(t, options)
}

func TestValidJavaVendorNoPathUsesDefinedNamespace(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    includeVendorNoPath,
		Gen:     "java:use_vendor",
		Out:     outputDir + "/valid_vendor_no_path",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"java/valid_vendor_no_path/FMyService.java", "valid_vendor_no_path/include_vendor_no_path/java/FMyService.java"},
		{"java/valid_vendor_no_path/MyScopePublisher.java", "valid_vendor_no_path/include_vendor_no_path/java/MyScopePublisher.java"},
		{"java/valid_vendor_no_path/MyScopeSubscriber.java", "valid_vendor_no_path/include_vendor_no_path/java/MyScopeSubscriber.java"},
		{"java/valid_vendor_no_path/VendoredReferences.java", "valid_vendor_no_path/include_vendor_no_path/java/VendoredReferences.java"},
		{"java/valid_vendor_no_path/InvalidData.java", "valid_vendor_no_path/InvalidData.java"},
	}
	suite.Run(t, options)

	filesNotToGenerate := []string{
		"valid_vendor_no_path/vendor_namespace/java/Item.java",
		"valid_vendor_no_path/vendor_namespace/java/vendor_namespaceConstants.java",
		"valid_vendor_no_path/vendor_namespace/java/MyEnum.java",
		"valid_vendor_no_path/vendor_namespace/java/FVendoredBase.java",
	}
	assertFilesNotExist(t, filesNotToGenerate)
}

func TestValidJavaSuppressDeprecatedLogging(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  frugalGenFile,
		Gen:   "java:suppress_deprecated_logging",
		Out:   outputDir + "/deprecated_logging",
		Delim: delim,
	}
	suite := ComparisonList{
		{"java/deprecated_logging/FFoo.java", "deprecated_logging/variety/java/FFoo.java"},
	}
	suite.Run(t, options)
}

func TestValidJsonFrugalCompiler(t *testing.T) {
	defer globals.Reset()
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "json:indent",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"frugal.json", "frugal.json"},
	}
	suite.Run(t, options)
}

func TestValidPython(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"python/variety/__init__.py", "variety/python/__init__.py"},
		{"python/variety/constants.py", "variety/python/constants.py"},
		{"python/variety/ttypes.py", "variety/python/ttypes.py"},
		{"python/variety/f_Events_publisher.py", "variety/python/f_Events_publisher.py"},
		{"python/variety/f_Events_subscriber.py", "variety/python/f_Events_subscriber.py"},
		{"python/variety/f_Foo.py", "variety/python/f_Foo.py"},

		{"python/actual_base/__init__.py", "actual_base/python/__init__.py"},
		{"python/actual_base/constants.py", "actual_base/python/constants.py"},
		{"python/actual_base/ttypes.py", "actual_base/python/ttypes.py"},
		{"python/actual_base/f_BaseFoo.py", "actual_base/python/f_BaseFoo.py"},
	}
	suite.Run(t, options)
}

func TestValidPythonTornado(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:tornado",
		Out:     outputDir + "/tornado",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"python.tornado/variety/__init__.py", "tornado/variety/python/__init__.py"},
		{"python.tornado/variety/constants.py", "tornado/variety/python/constants.py"},
		{"python.tornado/variety/ttypes.py", "tornado/variety/python/ttypes.py"},
		{"python.tornado/variety/f_Events_publisher.py", "tornado/variety/python/f_Events_publisher.py"},
		{"python.tornado/variety/f_Events_subscriber.py", "tornado/variety/python/f_Events_subscriber.py"},
		{"python.tornado/variety/f_Foo.py", "tornado/variety/python/f_Foo.py"},

		{"python.tornado/actual_base/__init__.py", "tornado/actual_base/python/__init__.py"},
		{"python.tornado/actual_base/constants.py", "tornado/actual_base/python/constants.py"},
		{"python.tornado/actual_base/ttypes.py", "tornado/actual_base/python/ttypes.py"},
		{"python.tornado/actual_base/f_BaseFoo.py", "tornado/actual_base/python/f_BaseFoo.py"},
	}
	suite.Run(t, options)
}

func TestValidPythonAsyncIO(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py:asyncio",
		Out:     outputDir + "/asyncio",
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"python.asyncio/variety/__init__.py", "asyncio/variety/python/__init__.py"},
		{"python.asyncio/variety/constants.py", "asyncio/variety/python/constants.py"},
		{"python.asyncio/variety/ttypes.py", "asyncio/variety/python/ttypes.py"},
		{"python.asyncio/variety/f_Events_publisher.py", "asyncio/variety/python/f_Events_publisher.py"},
		{"python.asyncio/variety/f_Events_subscriber.py", "asyncio/variety/python/f_Events_subscriber.py"},
		{"python.asyncio/variety/f_Foo.py", "asyncio/variety/python/f_Foo.py"},

		{"python.asyncio/actual_base/__init__.py", "asyncio/actual_base/python/__init__.py"},
		{"python.asyncio/actual_base/constants.py", "asyncio/actual_base/python/constants.py"},
		{"python.asyncio/actual_base/ttypes.py", "asyncio/actual_base/python/ttypes.py"},
		{"python.asyncio/actual_base/f_BaseFoo.py", "asyncio/actual_base/python/f_BaseFoo.py"},
	}
	suite.Run(t, options)
}

func TestPythonPackagePrefix(t *testing.T) {
	options := compiler.Options{
		File:    idl("service_inheritance.frugal"),
		Gen:     "py:package_prefix=generic_package_prefix.",
		Out:     outputDir,
		Delim:   delim,
		Recurse: false,
	}
	suite := ComparisonList{
		{"python/package_prefix/f_Foo.py", "service_inheritance/f_Foo.py"},
		{"python/package_prefix/ttypes.py", "service_inheritance/ttypes.py"},
		{"python/package_prefix/constants.py", "service_inheritance/constants.py"},
	}
	suite.Run(t, options)
}

func TestPythonExtendServiceSameFile(t *testing.T) {
	options := compiler.Options{
		File:  idl("service_extension_same_file.frugal"),
		Gen:   "py:asyncio",
		Out:   outputDir,
		Delim: delim,
	}
	suite := ComparisonList{
		{"python.asyncio/service_extension_same_file/f_BasePinger.py", "service_extension_same_file/python/f_BasePinger.py"},
		{"python.asyncio/service_extension_same_file/f_Pinger.py", "service_extension_same_file/python/f_Pinger.py"},
	}
	suite.Run(t, options)
}

func TestPythonAbsoluteOutputPath(t *testing.T) {
	absoluteOutputDir, err := filepath.Abs("testdata/out/absolute_path")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "py",
		Out:     absoluteOutputDir,
		Delim:   delim,
		Recurse: true,
	}
	suite := ComparisonList{
		{"python/variety/__init__.py", absoluteOutputDir + "/variety/python/__init__.py"},
		{"python/variety/constants.py", absoluteOutputDir + "/variety/python/constants.py"},
		{"python/variety/ttypes.py", absoluteOutputDir + "/variety/python/ttypes.py"},
		{"python/variety/f_Events_publisher.py", absoluteOutputDir + "/variety/python/f_Events_publisher.py"},
		{"python/variety/f_Events_subscriber.py", absoluteOutputDir + "/variety/python/f_Events_subscriber.py"},
		{"python/variety/f_Foo.py", absoluteOutputDir + "/variety/python/f_Foo.py"},

		{"python/actual_base/__init__.py", absoluteOutputDir + "/actual_base/python/__init__.py"},
		{"python/actual_base/constants.py", absoluteOutputDir + "/actual_base/python/constants.py"},
		{"python/actual_base/ttypes.py", absoluteOutputDir + "/actual_base/python/ttypes.py"},
		{"python/actual_base/f_BaseFoo.py", absoluteOutputDir + "/actual_base/python/f_BaseFoo.py"},
	}
	suite.Run(t, options)
}
