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
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestInvalid(t *testing.T) {
	options := compiler.Options{
		File:  invalidFile,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateServices(t *testing.T) {
	options := compiler.Options{
		File:  duplicateServices,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateScopes(t *testing.T) {
	options := compiler.Options{
		File:  duplicateScopes,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateMethods(t *testing.T) {
	options := compiler.Options{
		File:  duplicateMethods,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateOperations(t *testing.T) {
	options := compiler.Options{
		File:  duplicateOperations,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateMethodArgIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateMethodArgIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateStructFieldIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateStructFieldIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

// Ensures an error is returned when a "*" namespace has a vendor annotation.
func TestWildcardNamespaceWithVendorAnnotation(t *testing.T) {
	options := compiler.Options{
		File:  badNamespace,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
}
