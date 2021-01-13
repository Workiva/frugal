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

package globals

import (
	"fmt"
	"time"

	"github.com/Workiva/frugal/compiler/parser"
)

// Version of the Frugal compiler.
const Version = "3.13.0"

// Global variables.
var (
	TopicDelimiter = "."
	Gen            string
	Out            string
	FileDir        string
	DryRun         bool
	Recurse        bool
	Verbose        bool
	Now            = time.Now()
	CompiledFiles  = make(map[string]*parser.Frugal)
)

// Reset global variables to initial state.
func Reset() {
	TopicDelimiter = "."
	Gen = ""
	Out = ""
	FileDir = ""
	DryRun = false
	Recurse = false
	Verbose = false
	Now = time.Now()
	CompiledFiles = make(map[string]*parser.Frugal)
}

// PrintWarning prints the given message to stdout in yellow font.
func PrintWarning(msg string) {
	fmt.Println("\x1b[33m" + msg + "\x1b[0m")
}
