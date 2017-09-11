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

package html

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/parser"
)

const defaultOutputDir = "gen-html"

// Generator implements the ProgramGenerator interface for HTML.
type Generator struct {
	generatedIndex bool
	standalone     bool
}

// NewGenerator creates a new HTML ProgramGenerator.
func NewGenerator(options map[string]string) generator.ProgramGenerator {
	_, standalone := options["standalone"]
	return &Generator{standalone: standalone}
}

func (g *Generator) Generate(frugal *parser.Frugal, outputDir string) error {
	if !g.generatedIndex {
		if !g.standalone {
			stylesheet, err := os.Create(fmt.Sprintf("%s/style.css", outputDir))
			if err != nil {
				return err
			}
			if err := g.generateStylesheet(stylesheet); err != nil {
				return err
			}
		}
		index, err := os.Create(fmt.Sprintf("%s/index.html", outputDir))
		if err != nil {
			return err
		}
		if err := g.generateIndex(index, frugal); err != nil {
			return err
		}
		g.generatedIndex = true
	}

	file, err := os.Create(fmt.Sprintf("%s/%s.html", outputDir, frugal.Name))
	if err != nil {
		return err
	}
	if err := g.generateModule(file, frugal); err != nil {
		return err
	}

	return nil
}

func (g *Generator) GetOutputDir(dir string, frugal *parser.Frugal) string {
	return dir
}

func (g *Generator) DefaultOutputDir() string {
	return defaultOutputDir
}

func (g *Generator) generateStylesheet(file *os.File) error {
	_, err := file.WriteString(css)
	return err
}

type Modules []*parser.Frugal

func (m Modules) Len() int {
	return len(m)
}

func (m Modules) Less(i, j int) bool {
	return m[i].Name < m[j].Name
}

func (m Modules) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (g *Generator) generateIndex(file *os.File, frugal *parser.Frugal) error {
	modules := transitiveIncludes(frugal, Modules{})
	funcMap := template.FuncMap{"css": g.stylesheet}
	tpl, err := template.New("index").Funcs(funcMap).Parse(indexTemplate)
	if err != nil {
		return err
	}

	return tpl.Execute(file, modules)
}

func transitiveIncludes(module *parser.Frugal, modules Modules) Modules {
	modules = append(modules, module)
	for _, include := range module.ParsedIncludes {
		modules = transitiveIncludes(include, modules)
	}
	sort.Sort(modules)
	return modules
}

func (g *Generator) generateModule(file *os.File, module *parser.Frugal) error {
	funcMap := template.FuncMap{
		"css":        g.stylesheet,
		"capitalize": strings.Title,
		"lowercase":  strings.ToLower,
		"formatValue": func(value interface{}) template.HTML {
			return formatValue(value, module)
		},
		"displayType": func(typ *parser.Type) template.HTML {
			return displayType(typ, module)
		},
		"displayMethod": func(method *parser.Method) template.HTML {
			returnType := template.HTML("void")
			if method.ReturnType != nil {
				returnType = displayType(method.ReturnType, module)
			}
			display := fmt.Sprintf("%s %s(%s)", returnType, method.Name,
				displayMethodArgs(method.Arguments, module))
			throwsPrefix := "<br />    throws"
			for _, exception := range method.Exceptions {
				display += fmt.Sprintf("%s %s", throwsPrefix, displayType(exception.Type, module))
				throwsPrefix = ","
			}
			return template.HTML(display)
		},
		"displayService": func(service string) template.HTML {
			var anchor string
			if strings.Contains(service, ".") {
				includeAndName := strings.Split(service, ".")
				anchor = fmt.Sprintf("%s.html#svc_%s", includeAndName[0], includeAndName[1])
			} else {
				anchor = fmt.Sprintf("#svc_%s", service)
			}
			return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, anchor, service))
		},
	}
	tpl, err := template.New("module").Funcs(funcMap).Parse(moduleTemplate)
	if err != nil {
		return err
	}

	return tpl.Execute(file, module)
}

func (g *Generator) stylesheet() template.HTML {
	if !g.standalone {
		return template.HTML(`<link href="style.css" rel="stylesheet" type="text/css" />`)
	}
	return template.HTML(fmt.Sprintf("<style>%s</style>", css))
}

func formatValue(value interface{}, module *parser.Frugal) template.HTML {
	switch v := value.(type) {
	case string:
		return template.HTML(fmt.Sprintf(`"%s"`, v))
	case parser.Identifier:
		idCtx := module.ContextFromIdentifier(v)
		switch idCtx.Type {
		case parser.LocalConstant, parser.IncludeConstant:
			return template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, linkForConstant(idCtx.Constant, v), value))
		case parser.LocalEnum, parser.IncludeEnum:
			return template.HTML(fmt.Sprintf(`<a href="%s">%v</a>`, linkForEnum(idCtx.Enum, v), value))
		default:
			panic(fmt.Sprintf("unexpected type %d referenced by %s", idCtx.Type, value))
		}
	case []parser.KeyValue:
		display := "{ "
		prefix := ""
		for _, keyValue := range v {
			display += fmt.Sprintf("%s%s = %s", prefix,
				formatValue(keyValue.KeyToString(), module), formatValue(keyValue.Value, module))
			prefix = ", "
		}
		display += " }"
		return template.HTML(display)
	case []interface{}:
		display := "{ "
		prefix := ""
		for _, value := range v {
			display += fmt.Sprintf("%s%s", prefix, formatValue(value, module))
			prefix = ", "
		}
		display += " }"
		return template.HTML(display)
	default:
		return template.HTML(fmt.Sprintf("%v", v))
	}
}

func displayMethodArgs(args []*parser.Field, module *parser.Frugal) string {
	display := ""
	prefix := ""
	for _, arg := range args {
		display += fmt.Sprintf("%s%s %s", prefix, displayType(arg.Type, module), arg.Name)
		prefix = ", "
	}
	return display
}

func linkForConstant(constant *parser.Constant, identifier parser.Identifier) string {
	link := fmt.Sprintf("#const_%s", constant.Name)
	if strings.Contains(string(identifier), ".") {
		includeAndName := strings.Split(string(identifier), ".")
		link = fmt.Sprintf("%s.html%s", includeAndName[0], link)
	}
	return link
}

func linkForEnum(enum *parser.Enum, identifier parser.Identifier) string {
	link := fmt.Sprintf("#enum_%s", enum.Name)
	pieces := strings.Split(string(identifier), ".")
	if len(pieces) == 3 {
		link = fmt.Sprintf("%s.html%s", pieces[0], link)
	}
	return link
}

func linkForType(typ *parser.Type, module *parser.Frugal) string {
	var anchor string
	if module.IsEnum(typ) {
		anchor = fmt.Sprintf("#enum_%s", typ.ParamName())
	} else if module.IsStruct(typ) {
		anchor = fmt.Sprintf("#struct_%s", typ.ParamName())
	} else {
		anchor = fmt.Sprintf("#typedef_%s", typ.ParamName())
	}
	if strings.Contains(typ.Name, ".") {
		anchor = fmt.Sprintf("%s.html%s", typ.IncludeName(), anchor)
	}
	return anchor
}

func displayType(typ *parser.Type, module *parser.Frugal) template.HTML {
	if typ.IsPrimitive() {
		return template.HTML(typ.String())
	}
	if typ.IsCustom() {
		return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, linkForType(typ, module), typ.String()))
	}
	if typ.KeyType != nil {
		return template.HTML(fmt.Sprintf("map&lt;%s, %s&gt;",
			displayType(typ.KeyType, module), displayType(typ.ValueType, module)))
	}
	if typ.Name == "list" {
		return template.HTML(fmt.Sprintf("list&lt;%s&gt;", displayType(typ.ValueType, module)))
	}
	return template.HTML(fmt.Sprintf("set&lt;%s&gt;", displayType(typ.ValueType, module)))
}
