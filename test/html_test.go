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

func TestHTML(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "html",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/html/style.css", filepath.Join(outputDir, "style.css")},
		{"expected/html/index.html", filepath.Join(outputDir, "index.html")},
		{"expected/html/base.html", filepath.Join(outputDir, "base.html")},
		{"expected/html/variety.html", filepath.Join(outputDir, "variety.html")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}

func TestHTMLStandalone(t *testing.T) {
	options := compiler.Options{
		File:    frugalGenFile,
		Gen:     "html:standalone",
		Out:     outputDir,
		Delim:   delim,
		Recurse: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	files := []FileComparisonPair{
		{"expected/html/standalone/index.html", filepath.Join(outputDir, "index.html")},
		{"expected/html/standalone/base.html", filepath.Join(outputDir, "base.html")},
		{"expected/html/standalone/variety.html", filepath.Join(outputDir, "variety.html")},
	}

	copyAllFiles(t, files)
	compareAllFiles(t, files)
}
