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
		File:  frugalGenFile,
		Gen:   "java:async",
		Out:   outputDir + "/async",
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/variety_async/FFoo.java", filepath.Join(outputDir, "async", "variety", "java", "FFoo.java")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
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

	files := []FileComparisonPair{
		{"expected/java/variety/AwesomeException.java", filepath.Join(outputDir, "variety", "java", "AwesomeException.java")},
		{"expected/java/variety/Event.java", filepath.Join(outputDir, "variety", "java", "Event.java")},
		{"expected/java/variety/EventWrapper.java", filepath.Join(outputDir, "variety", "java", "EventWrapper.java")},
		{"expected/java/variety/ItsAnEnum.java", filepath.Join(outputDir, "variety", "java", "ItsAnEnum.java")},
		{"expected/java/variety/TestBase.java", filepath.Join(outputDir, "variety", "java", "TestBase.java")},
		{"expected/java/variety/TestingDefaults.java", filepath.Join(outputDir, "variety", "java", "TestingDefaults.java")},
		{"expected/java/variety/TestingUnions.java", filepath.Join(outputDir, "variety", "java", "TestingUnions.java")},
		{"expected/java/variety/HealthCondition.java", filepath.Join(outputDir, "variety", "java", "HealthCondition.java")},
		{"expected/java/variety/varietyConstants.java", filepath.Join(outputDir, "variety", "java", "varietyConstants.java")},
		{"expected/java/variety/TestLowercase.java", filepath.Join(outputDir, "variety", "java", "TestLowercase.java")},
		{"expected/java/variety/FooArgs.java", filepath.Join(outputDir, "variety", "java", "FooArgs.java")},
		{"expected/java/variety/EventsPublisher.java", filepath.Join(outputDir, "variety", "java", "EventsPublisher.java")},
		{"expected/java/variety/EventsSubscriber.java", filepath.Join(outputDir, "variety", "java", "EventsSubscriber.java")},
		{"expected/java/variety/FFoo.java", filepath.Join(outputDir, "variety", "java", "FFoo.java")},

		{"expected/java/actual_base/api_exception.java", filepath.Join(outputDir, "actual_base", "java", "api_exception.java")},
		{"expected/java/actual_base/baseConstants.java", filepath.Join(outputDir, "actual_base", "java", "baseConstants.java")},
		{"expected/java/actual_base/thing.java", filepath.Join(outputDir, "actual_base", "java", "thing.java")},
		{"expected/java/actual_base/base_health_condition.java", filepath.Join(outputDir, "actual_base", "java", "base_health_condition.java")},
		{"expected/java/actual_base/FBaseFoo.java", filepath.Join(outputDir, "actual_base", "java", "FBaseFoo.java")},
		{"expected/java/actual_base/nested_thing.java", filepath.Join(outputDir, "actual_base", "java", "nested_thing.java")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/boxed_primitives/FFoo.java", filepath.Join(outputDir, "boxed_primitives", "variety", "java", "FFoo.java")},
		{"expected/java/boxed_primitives/TestingDefaults.java", filepath.Join(outputDir, "boxed_primitives", "variety", "java", "TestingDefaults.java")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/valid_vendor/FMyService.java", filepath.Join(outputDir, "valid_vendor", "include_vendor", "java", "FMyService.java")},
		{"expected/java/valid_vendor/MyScopePublisher.java", filepath.Join(outputDir, "valid_vendor", "include_vendor", "java", "MyScopePublisher.java")},
		{"expected/java/valid_vendor/MyScopeSubscriber.java", filepath.Join(outputDir, "valid_vendor", "include_vendor", "java", "MyScopeSubscriber.java")},
		{"expected/java/valid_vendor/VendoredReferences.java", filepath.Join(outputDir, "valid_vendor", "include_vendor", "java", "VendoredReferences.java")},
		{"expected/java/valid_vendor/InvalidData.java", filepath.Join(outputDir, "valid_vendor", "InvalidData.java")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)

	filesNotToGenerate := []string{
		filepath.Join(filepath.Join(outputDir, "valid_vendor", "vendor_namespace", "java", "Item.java")),
		filepath.Join(filepath.Join(outputDir, "valid_vendor", "vendor_namespace", "java", "vendor_namespaceConstants.java")),
		filepath.Join(filepath.Join(outputDir, "valid_vendor", "vendor_namespace", "java", "MyEnum.java")),
		filepath.Join(filepath.Join(outputDir, "valid_vendor", "vendor_namespace", "java", "FVendoredBase.java")),
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/vendored_but_no_use_vendor/FMyService.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "include_vendor", "java", "FMyService.java")},
		{"expected/java/vendored_but_no_use_vendor/MyScopePublisher.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "include_vendor", "java", "MyScopePublisher.java")},
		{"expected/java/vendored_but_no_use_vendor/MyScopeSubscriber.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "include_vendor", "java", "MyScopeSubscriber.java")},
		{"expected/java/vendored_but_no_use_vendor/VendoredReferences.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "include_vendor", "java", "VendoredReferences.java")},
		{"expected/java/vendored_but_no_use_vendor/InvalidData.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "InvalidData.java")},
		{"expected/java/vendored_but_no_use_vendor/Item.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "vendor_namespace", "java", "Item.java")},
		{"expected/java/vendored_but_no_use_vendor/vendor_namespaceConstants.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "vendor_namespace", "java", "vendor_namespaceConstants.java")},
		{"expected/java/vendored_but_no_use_vendor/MyEnum.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "vendor_namespace", "java", "MyEnum.java")},
		{"expected/java/vendored_but_no_use_vendor/FVendoredBase.java", filepath.Join(outputDir, "vendored_but_no_use_vendor", "vendor_namespace", "java", "FVendoredBase.java")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/valid_vendor_no_path/FMyService.java", filepath.Join(outputDir, "valid_vendor_no_path", "include_vendor_no_path", "java", "FMyService.java")},
		{"expected/java/valid_vendor_no_path/MyScopePublisher.java", filepath.Join(outputDir, "valid_vendor_no_path", "include_vendor_no_path", "java", "MyScopePublisher.java")},
		{"expected/java/valid_vendor_no_path/MyScopeSubscriber.java", filepath.Join(outputDir, "valid_vendor_no_path", "include_vendor_no_path", "java", "MyScopeSubscriber.java")},
		{"expected/java/valid_vendor_no_path/VendoredReferences.java", filepath.Join(outputDir, "valid_vendor_no_path", "include_vendor_no_path", "java", "VendoredReferences.java")},
		{"expected/java/valid_vendor_no_path/InvalidData.java", filepath.Join(outputDir, "valid_vendor_no_path", "InvalidData.java")},
	}
	copyAllFiles(t, files)
	compareAllFiles(t, files)

	filesNotToGenerate := []string{
		filepath.Join(outputDir, "valid_vendor_no_path", "vendor_namespace", "java", "Item.java"),
		filepath.Join(outputDir, "valid_vendor_no_path", "vendor_namespace", "java", "vendor_namespaceConstants.java"),
		filepath.Join(outputDir, "valid_vendor_no_path", "vendor_namespace", "java", "MyEnum.java"),
		filepath.Join(outputDir, "valid_vendor_no_path", "vendor_namespace", "java", "FVendoredBase.java"),
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
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/deprecated_logging/FFoo.java", filepath.Join(outputDir, "deprecated_logging", "variety", "java", "FFoo.java")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestValidJavaSuppressGeneratedAnnotations(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  frugalGenFile,
		Gen:   "java:generated_annotations=suppress",
		Out:   outputDir + "/suppress_generated_annotations",
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/java/suppress_generated_annotations/FFoo.java", filepath.Join(outputDir, "suppress_generated_annotations", "variety", "java", "FFoo.java")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}
