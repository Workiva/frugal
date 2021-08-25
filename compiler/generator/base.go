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

package generator

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Workiva/frugal/compiler/parser"
	"gopkg.in/yaml.v2"
)

// BaseGenerator contains base generator logic which language generators can
// extend.
type BaseGenerator struct {
	Options map[string]string
	Frugal  *parser.Frugal
	elemNum int
}

func (b *BaseGenerator) FilterInput(f *parser.Frugal) {
	filename, ok := b.Options[`filter_yaml`]
	if !ok {
		return
	}
	if len(filename) == 0 {
		filename = `frugal_filter.yaml`
	}

	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("could not read file %q: %v\n", filename, err)
		return
	}

	gf := &generatorFilter{}

	err = yaml.Unmarshal(input, gf)
	if err != nil {
		fmt.Printf("could not unmarshl file %q: %v\n", filename, err)
		return
	}

	for i := 0; i < len(f.Services); i++ {
		service := f.Services[i]
		if shouldEntirelyRemoveService(gf, service) {
			f.Services = append(f.Services[:i], f.Services[i+1:]...)
			i--
			continue
		}

		applyFilterToService(gf, service)
	}

	// TODO do the same for structs...
}

// CreateFile creates a new file using the given configuration.
func (b *BaseGenerator) CreateFile(name, outputDir, suffix string, usePrefix bool) (*os.File, error) {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return nil, err
	}
	prefix := ""
	if usePrefix {
		prefix = FilePrefix
	}
	return os.Create(fmt.Sprintf("%s/%s%s.%s", outputDir, prefix, name, suffix))
}

// GenerateNewline adds the specific number of newlines to the given file.
func (b *BaseGenerator) GenerateNewline(file *os.File, count int) error {
	str := ""
	for i := 0; i < count; i++ {
		str += "\n"
	}
	_, err := file.WriteString(str)
	return err
}

// GenerateInlineComment generates an inline comment.
func (b *BaseGenerator) GenerateInlineComment(comment []string, indent string) string {
	inline := ""
	for _, line := range comment {
		inline += indent + "// " + line + "\n"
	}
	return inline
}

// GenerateBlockComment generates a C-style comment.
func (b *BaseGenerator) GenerateBlockComment(comment []string, indent string) string {
	block := indent + "/**\n"
	for _, line := range comment {
		block += indent + " * " + line + "\n"
	}
	block += indent + " */\n"
	return block
}

// SetFrugal sets the Frugal parse tree for this generator.
func (b *BaseGenerator) SetFrugal(f *parser.Frugal) {
	b.Frugal = f
}

func (b *BaseGenerator) GetElem() string {
	s := fmt.Sprintf("elem%d", b.elemNum)
	b.elemNum++
	return s
}

func (b *BaseGenerator) GetServiceMethodTypes(service *parser.Service) []*parser.Struct {
	structs := []*parser.Struct{}
	for _, method := range service.Methods {
		arg := &parser.Struct{
			Name:   fmt.Sprintf("%s_args", method.Name),
			Fields: method.Arguments,
			Type:   parser.StructTypeStruct,
		}

		for _, field := range arg.Fields {
			if field.Modifier == parser.Optional {
				field.Modifier = parser.Default
			}
		}
		structs = append(structs, arg)

		if !method.Oneway {
			numReturns := 0
			if method.ReturnType != nil {
				numReturns = 1
			}

			fields := make([]*parser.Field, len(method.Exceptions)+numReturns, len(method.Exceptions)+numReturns)
			if numReturns == 1 {
				fields[0] = parser.FieldFromType(method.ReturnType, "success")
			}
			copy(fields[numReturns:], method.Exceptions)
			for _, field := range fields {
				field.Modifier = parser.Optional
			}

			result := &parser.Struct{
				Name:   fmt.Sprintf("%s_result", method.Name),
				Fields: fields,
				Type:   parser.StructTypeStruct,
			}
			structs = append(structs, result)
		}
	}
	return structs
}
