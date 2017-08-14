/*
 * Copyright 2017 Workiva
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

	files := []FileComparisonPair{
		{"expected/dart/variety/f_awesome_exception.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_awesome_exception.dart")},
		{"expected/dart/variety/f_event.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_event.dart")},
		{"expected/dart/variety/f_event_wrapper.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_event_wrapper.dart")},
		{"expected/dart/variety/f_its_an_enum.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_its_an_enum.dart")},
		{"expected/dart/variety/f_test_base.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_test_base.dart")},
		{"expected/dart/variety/f_testing_defaults.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_testing_defaults.dart")},
		{"expected/dart/variety/f_testing_unions.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_testing_unions.dart")},
		{"expected/dart/variety/f_health_condition.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_health_condition.dart")},
		{"expected/dart/variety/f_test_lowercase.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_test_lowercase.dart")},
		{"expected/dart/variety/f_variety_constants.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_variety_constants.dart")},
		{"expected/dart/variety/f_events_scope.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_events_scope.dart")},
		{"expected/dart/variety/f_foo_service.dart", filepath.Join(outputDir, "variety", "lib", "src", "f_foo_service.dart")},
		{"expected/dart/variety/variety.dart", filepath.Join(outputDir, "variety", "lib", "variety.dart")},

		{"expected/dart/actual_base/f_actual_base_dart_constants.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_actual_base_dart_constants.dart")},
		{"expected/dart/actual_base/f_api_exception.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_api_exception.dart")},
		{"expected/dart/actual_base/f_thing.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_thing.dart")},
		{"expected/dart/actual_base/f_base_health_condition.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_base_health_condition.dart")},
		{"expected/dart/actual_base/f_base_foo_service.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_base_foo_service.dart")},
		{"expected/dart/actual_base/f_nested_thing.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "src", "f_nested_thing.dart")},
		{"expected/dart/actual_base/actual_base_dart.dart", filepath.Join(outputDir, "actual_base_dart", "lib", "actual_base_dart.dart")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)

}

func TestValidDartEnums(t *testing.T) {
	options := compiler.Options{
		File:    "idl/enum.frugal",
		Gen:     "dart:use_enums",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/dart/enum/f_testing_enums.dart", filepath.Join(outputDir, "enum_dart", "lib", "src", "f_testing_enums.dart")},
		{"expected/dart/enum/enum_dart.dart", filepath.Join(outputDir, "enum_dart", "lib", "enum_dart.dart")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{
			"expected/dart/include_vendor/f_my_scope_scope.dart",
			filepath.Join(outputDir, "include_vendor", "lib", "src", "f_my_scope_scope.dart"),
		},
		{
			"expected/dart/include_vendor/f_my_service_service.dart",
			filepath.Join(outputDir, "include_vendor", "lib", "src", "f_my_service_service.dart"),
		},
		{
			"expected/dart/include_vendor/include_vendor.dart",
			filepath.Join(outputDir, "include_vendor", "lib", "include_vendor.dart"),
		},
		{
			"expected/dart/include_vendor/pubspec.yaml",
			filepath.Join(outputDir, "include_vendor", "pubspec.yaml"),
		},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
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
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/dart/vendor_namespace/vendor_namespace.dart", filepath.Join(outputDir, "vendor_namespace", "lib", "vendor_namespace.dart")},
		{"expected/dart/vendor_namespace/f_item.dart", filepath.Join(outputDir, "vendor_namespace", "lib", "src", "f_item.dart")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
}
