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
	"github.com/stretchr/testify/assert"
)

const circularFile = "idl/circular_1.frugal"

func TestCircularIncludes(t *testing.T) {
	options := compiler.Options{
		File:   circularFile,
		Gen:    "go",
		Out:    "out",
		Delim:  ".",
		DryRun: true,
	}
	err := compiler.Compile(options)
	assert.Error(t, err)
	assert.Equal(
		t,
		"Include circular_2.frugal: Include circular_3.frugal: Include circular_1.frugal: Circular include: [circular_1 circular_2 circular_3 circular_1]",
		err.Error())
}
