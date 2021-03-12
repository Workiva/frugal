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

package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Supported generator annotations.
const (
	// VendorAnnotation is used on namespace definitions to indicate to any
	// consumers of the IDL where the generated code is vendored so that
	// consumers can generate code that points to it. This cannot be used with
	// "*" namespaces since it is language-dependent. Consumers then use the
	// "vendor" annotation on includes they wish to vendor. The value provided
	// on the include-side "vendor" annotation, if any, is ignored.
	//
	// When an include is annotated with "vendor", Frugal will skip generating
	// the include if -use-vendor is set since this flag indicates intention to
	// use the vendored code as advertised by the "vendor" annotation.
	//
	// If no location is specified by the "vendor" annotation, the behavior is
	// defined by the language generator.
	VendorAnnotation = "vendor"

	// DeprecatedAnnotation is the annotation to mark a service method as deprecated.
	DeprecatedAnnotation = "deprecated"
)

// ParseFrugal parses the given Frugal file into its semantic representation.
func ParseFrugal(filePath string, includeDirs []string) (*Frugal, error) {
	return parseFrugal(filePath, []string{}, includeDirs)
}

func parseFrugal(filePath string, visitedIncludes []string, includeDirs []string) (*Frugal, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	name, err := getName(file)
	if err != nil {
		return nil, err
	}

	if contains(visitedIncludes, name) {
		return nil, fmt.Errorf("Circular include: %s", append(visitedIncludes, name))
	}
	visitedIncludes = append(visitedIncludes, name)

	parsed, err := ParseReader(filePath, file)
	if err != nil {
		return nil, err
	}

	frugal := parsed.(*Frugal)
	frugal.Name = name
	frugal.File = filePath
	frugal.Dir = filepath.Dir(file.Name())
	frugal.Path = filePath
	for _, incl := range frugal.Includes {
		include := incl.Value
		if !strings.HasSuffix(include, ".thrift") && !strings.HasSuffix(include, ".frugal") {
			return nil, fmt.Errorf("Bad include name: %s", include)
		}

		inc := include
		if !filepath.IsAbs(inc) {
			inc = filepath.Join(frugal.Dir, include)
		}

		parsedIncl, err := parseFrugal(inc, visitedIncludes, includeDirs)
		if err != nil {

			if includeDirs == nil || len(includeDirs) == 0 {
				return nil, fmt.Errorf("Include %s: %s", include, err)
			}

			for _, includeDir := range includeDirs {
				inc = filepath.Join(includeDir, include)
				p, err := parseFrugal(inc, visitedIncludes, includeDirs)

				if err == nil {
					parsedIncl = p
					break
				}

			}

			if parsedIncl == nil {
				return nil, fmt.Errorf("Include %s: %s, %s", include, includeDirs, err)
			}
		}

		// Lop off extension (.frugal or .thrift)
		includeBase := include[:len(include)-7]

		// Lop off path
		includeName := filepath.Base(includeBase)

		frugal.ParsedIncludes[includeName] = parsedIncl
	}

	if err := frugal.validate(); err != nil {
		return nil, err
	}

	frugal.sort() // For determinism in generated code
	frugal.assignFrugal()

	return frugal, nil
}

func getName(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	parts := strings.Split(info.Name(), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid file: %s", f.Name())
	}
	return parts[0], nil
}

func contains(arr []string, e string) bool {
	for _, item := range arr {
		if item == e {
			return true
		}
	}
	return false
}
