// vim: set ts=4 sw=4 tw=99 noet:
//
// Copyright 2014, Edmodo, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this work except in compliance with the License.
// You may obtain a copy of the License in the LICENSE file, or at:
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS"
// BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language
// governing permissions and limitations under the License.

package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

type CompileError struct {
	File    string
	Pos     Position
	Message string
}

type CompileContext struct {
	// Current file being operated on, if any.
	CurFile string

	// List of errors encountered so far.
	Errors []*CompileError

	Packages map[string]*ParseTree
}

func NewCompileContext() *CompileContext {
	return &CompileContext{}
}

// Return the folder and filename. The filename has ".thrift" stripped.
func (this *CompileContext) splitPath(file string) (string, string) {
	folder, name := filepath.Split(file)

	index := strings.LastIndex(name, ".thrift")
	if index != -1 {
		name = name[:index]
	}

	return folder, name
}

func (this *CompileContext) parse(path string) *ParseTree {
	this.Enter(path)
	defer this.Leave()

	parser, err := NewParser(this)
	if err != nil {
		this.ReportError(Position{}, "Could not open file: %s", err.Error())
		return nil
	}

	return parser.Parse()
}

func (this *CompileContext) ParseRecursive(file string) *ParseTree {
	folder, name := this.splitPath(file)

	queue := []string{name}
	parsed := map[string]*ParseTree{
		name: nil,
	}

	for len(queue) > 0 {
		// Pop a file off the queue.
		name := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		// Parse the file.
		path := filepath.Join(folder, name) + ".thrift"
		tree := this.parse(path)
		if tree == nil {
			return nil
		}

		// Mark this as parsed.
		tree.Package = name
		parsed[name] = tree

		// Enqueue everything that hasn't been parsed.
		for name, tree := range tree.Includes {
			if tree.Tree != nil {
				continue
			}

			_, inQueue := parsed[name]
			if !inQueue {
				// Add to the parsing queue, and flag the parsing results with a nil so
				// we don't add to the queue twice.
				queue = append(queue, name)
				parsed[name] = nil
			}
		}
	}

	// If we got here, everything was parsed. Update all the mappings.
	for _, tree := range parsed {
		for name, _ := range tree.Includes {
			tree.Includes[name].Tree = parsed[name]
		}
	}

	// Return the root of all parse trees (the first file).
	return parsed[name]
}

func (this *CompileContext) Enter(file string) {
	if this.CurFile != "" {
		panic("Cannot nested files")
	}
	this.CurFile = file
}

func (this *CompileContext) Leave() {
	this.CurFile = ""
}

func (this *CompileContext) HasErrors() bool {
	return len(this.Errors) > 0
}

func (this *CompileContext) ReportError(pos Position, str string, args ...interface{}) {
	this.Errors = append(this.Errors, &CompileError{
		File:    this.CurFile,
		Pos:     pos,
		Message: fmt.Sprintf(str, args...),
	})
}

func (this *CompileContext) PrintErrors() {
	for _, err := range this.Errors {
		fmt.Printf("%s (line %d, col %d): %s\n", err.File, err.Pos.Line, err.Pos.Col, err.Message)
	}
}

func (this *CompileContext) ReportRedeclaration(pos Position, name *Token) {
	this.ReportError(pos, "name '%s' was already declared on %s", name.Identifier(), name.Loc.Start)
}

// Flatten a tree of parse trees into a list, in no particular order.
func FlattenTrees(tree *ParseTree) []*ParseTree {
	trees := []*ParseTree{}
	processed := map[*ParseTree]bool{}

	queue := []*ParseTree{tree}
	for len(queue) > 0 {
		item := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		trees = append(trees, item)

		for _, other := range item.Includes {
			if other.Tree == nil {
				panic(fmt.Errorf("Cannot flatten an incomplete parse tree"))
			}
			if _, ok := processed[other.Tree]; ok {
				continue
			}
			queue = append(queue, other.Tree)
			processed[other.Tree] = true
		}
	}

	return trees
}
