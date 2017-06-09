package compiler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/generator/dartlang"
	"github.com/Workiva/frugal/compiler/generator/golang"
	"github.com/Workiva/frugal/compiler/generator/html"
	"github.com/Workiva/frugal/compiler/generator/java"
	"github.com/Workiva/frugal/compiler/generator/python"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

// Options contains compiler options for code generation.
type Options struct {
	File    string // Frugal file to generate
	Gen     string // Language to generate
	Out     string // Output location for generated code
	Delim   string // Token delimiter for scope topics
	DryRun  bool   // Do not generate code
	Recurse bool   // Generate includes
	Verbose bool   // Verbose mode
	JSON    bool   // Json dump of the frugal model
}

// Compile parses the Frugal IDL and generates code for it, returning an error
// if something failed.
func Compile(options Options) error {
	var err error
	defer globals.Reset()
	globals.TopicDelimiter = options.Delim
	globals.Gen = options.Gen
	globals.Out = options.Out
	globals.DryRun = options.DryRun
	globals.Recurse = options.Recurse
	globals.Verbose = options.Verbose
	globals.FileDir = filepath.Dir(options.File)

	absFile, err := filepath.Abs(options.File)
	if err != nil {
		return err
	}

	frugal, err := parseFrugal(absFile)
	if err != nil {
		return err
	}
	if options.JSON {
		return generateSemVerAudit(frugal)
	}

	return generateFrugal(frugal)
}

// parseFrugal parses a frugal file.
func parseFrugal(file string) (*parser.Frugal, error) {
	if !exists(file) {
		return nil, fmt.Errorf("Frugal file not found: %s\n", file)
	}
	logv(fmt.Sprintf("Parsing %s", file))
	return parser.ParseFrugal(file)
}

// generateFrugal generates code for a frugal struct.
func generateFrugal(f *parser.Frugal) error {
	var gen = globals.Gen

	lang, options, err := cleanGenParam(gen)
	if err != nil {
		return err
	}

	// Resolve Frugal generator.
	g, err := getProgramGenerator(lang, options)
	if err != nil {
		return err
	}

	// The parsed frugal contains everything needed to generate
	if err := generateFrugalRec(f, g, true, lang); err != nil {
		return err
	}

	return nil
}

func generateJSON(f *parser.Frugal) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(globals.Out, f.Name+".json"), b, 0644)
	if err != nil {
		return err
	}
	return nil
}

type semVer struct {
	Exports    []*export `json:"exports"`
	BaseBranch string    `json:"base_branch"`
	BaseCommit string    `json:"base_commit"`
	BuildID    string    `json:"build_id"`
	HeadBranch string    `json:"head_branch"`
	HeadCommit string    `json:"head_commit"`
	Language   string    `json:"language"`
	Pull       string    `json:"pull"`
	Repo       string    `json:"repo"`
	Version    string    `json:"version"`
}

type export struct {
	Key       string      `json:"key"`
	ParentKey string      `json:"parent_key"`
	Type      string      `json:"type"`
	Grammer   interface{} `json:"grammer"`
	Meta      meta        `json:"meta"`
}

// ############## GRAMMER ##############
type varGrammer struct {
	sig
	Getter bool   `json:"getter"`
	Setter bool   `json:"setter"`
	Type   string `json:"type"`
}

type funcGrammer struct {
	sig
	Parameters params `json:"parameters"` // Ordered
	ReturnType string `json:"return_type"`
}

type enumGrammer struct {
	sig
	values []string `json:"values"`
}

type typeDefGrammer struct {
	sig
	ReturnType string `json:"return_type"`
	Parameters params `json:"parameters"`
}

type classGrammer struct {
	sig
	Extends    []string `json:"extends"`
	Mixins     []string `json:"mixins"`
	Implements []string `json:"implements"`
}

// For both named and default constructors (default just has null for sig.Name)
type constructorGrammer struct {
	sig
	Parameters params `json:"parameters"`
}

type fieldGrammer struct {
	sig
	Getter bool   `json:"getter"`
	Setter bool   `json:"setter"`
	Static bool   `json:"static"`
	Type   string `json:"type"`
}

type methodGrammer struct {
	// Signature  string `json:"signature"`
	// Name       string `json:"name"`
	// Parameters params `json:"parameters"`
	// ReturnType string `json:"return_type"`
	funcGrammer
	Static bool `json:"static"`
}

// ############## INTERNALS TO GRAMMER ##############
type sig struct {
	Signature string `json:"signature"`
	Name      string `json:"name"`
}

type params struct {
	Named      []string   `json:"named"`      // Maintain order with positional
	Positional []position `json:"positional"` // Maintain order with named
}

type position struct {
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

type meta struct {
	Line int64  `json:"line"`
	URI  string `json:"uri"`
}

func getConstants(f *parser.Frugal) (exp []*export) {
	var e *export
	// Constant maps to variable grammer in semver audit service
	for _, c := range f.Constants {
		e = &export{
			Key:     c.Name,
			Type:    f.UnderlyingType(c.Type).Name,
			Grammer: varGrammer{},
		}
		exp = append(exp, e)
	}
	return exp
}

func getEnums(f *parser.Frugal) (exp []*export) {
	var e *export
	// Eums
	for _, c := range f.Enums {
		e = &export{
			Key:  c.Name,
			Type: "enum",
		}
		var g enumGrammer
		for _, v := range c.Values {
			g.Name = v.Name
		}
		e.Grammer = g
		exp = append(exp, e)
	}
	return exp
}

func getScopes(f *parser.Frugal) (exp []*export) {
	var e *export
	// TODO Scopes map to ??? []enum maybe...
	for _, c := range f.Scopes {
		e = &export{
			Key:  c.Name,
			Type: "scope",
		}
		exp = append(exp, e)
	}
	return exp
}

func getServices(f *parser.Frugal) (exp []*export) {
	var e *export
	// TODO Services map to ?? []methods maybe...
	for _, c := range f.Services {
		e = &export{
			Key:  c.Name,
			Type: "service",
		}
		exp = append(exp, e)
	}
	return exp
}

func getStructs(structs []*parser.Struct) (exp []*export) {
	var e *export
	var g classGrammer
	// Struct maps to class grammer in semver audit service
	for _, c := range structs {
		e = &export{
			Key:  c.Name,
			Type: "class",
		}
		for _, field := range c.Fields {
			g.Name = field.Name
		}
		e.Grammer = g
		exp = append(exp, e)
	}
	return exp
}

func getTypeDefs(f *parser.Frugal) (exp []*export) {
	var e *export
	for _, c := range f.Typedefs {
		e = &export{
			Key:  c.Name,
			Type: "typedef",
		}
		var g typeDefGrammer
		g.Name = c.Name
		e.Grammer = g
		exp = append(exp, e)
	}
	return exp
}

func generateSemVerAudit(f *parser.Frugal) error {
	// Build all the exported things in semver audit service from parsed frugal model
	var exp []*export

	exp = append(exp, getScopes(f)...)
	// TODO NAMESPACES?? current frugal audit does check namespaces
	exp = append(exp, getConstants(f)...)
	exp = append(exp, getEnums(f)...)

	exp = append(exp, getServices(f)...)
	exp = append(exp, getStructs(f.Structs)...)
	exp = append(exp, getStructs(f.Exceptions)...)
	exp = append(exp, getStructs(f.Unions)...)
	// exp = append(exp, getTypeDefs(f)...) As long as we use core types can we avoid this? it's not in frugal audit right now

	// Create semver object and serialize to json
	semver := &semVer{Exports: exp, Repo: "foobar"}
	b, err := json.MarshalIndent(semver, "", "    ")
	fmt.Println(len(b))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(globals.Out, f.Name+"_SEMVER.json"), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// generateFrugalRec generates code for a frugal struct, recursively generating
// code for includes
func generateFrugalRec(f *parser.Frugal, g generator.ProgramGenerator, generate bool, lang string) error {
	if _, ok := globals.CompiledFiles[f.File]; ok {
		// Already generated this file
		return nil
	}
	globals.CompiledFiles[f.File] = f

	out := globals.Out
	if out == "" {
		out = g.DefaultOutputDir()
	}
	fullOut := g.GetOutputDir(out, f)
	if err := os.MkdirAll(out, 0777); err != nil {
		return err
	}

	logv(fmt.Sprintf("Generating \"%s\" Frugal code for %s", lang, f.File))
	if globals.DryRun || !generate {
		return nil
	}

	if err := g.Generate(f, fullOut); err != nil {
		return err
	}

	// Iterate through includes in order to ensure determinism in
	// generated code.
	for _, inclFrugal := range f.OrderedIncludes() {
		if err := generateFrugalRec(inclFrugal, g, globals.Recurse, lang); err != nil {
			return err
		}
	}

	return nil
}

// getProgramGenerator resolves the ProgramGenerator for the given language. It
// returns an error if the language is not supported.
func getProgramGenerator(lang string, options map[string]string) (generator.ProgramGenerator, error) {
	var g generator.ProgramGenerator
	switch lang {
	case "dart":
		g = generator.NewProgramGenerator(dartlang.NewGenerator(options), false)
	case "go":
		// Make sure the package prefix ends with a "/"
		if package_prefix, ok := options["package_prefix"]; ok {
			if package_prefix != "" && !strings.HasSuffix(package_prefix, "/") {
				options["package_prefix"] = package_prefix + "/"
			}
		}

		g = generator.NewProgramGenerator(golang.NewGenerator(options), false)
	case "java":
		g = generator.NewProgramGenerator(java.NewGenerator(options), true)
	case "py":
		g = generator.NewProgramGenerator(python.NewGenerator(options), true)
	case "html":
		g = html.NewGenerator(options)
	default:
		return nil, fmt.Errorf("Invalid gen value %s", lang)
	}
	return g, nil
}

// exists determines if the file at the given path exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// cleanGenParam processes a string that includes an optional trailing
// options set.  Format: <language>:<name>=<value>,<name>=<value>,...
func cleanGenParam(gen string) (lang string, options map[string]string, err error) {
	lang = gen
	options = make(map[string]string)
	if strings.Contains(gen, ":") {
		s := strings.Split(gen, ":")
		lang = s[0]
		dirty := s[1]
		var optionArray []string
		if strings.Contains(dirty, ",") {
			optionArray = strings.Split(dirty, ",")
		} else {
			optionArray = append(optionArray, dirty)
		}
		for _, option := range optionArray {
			s := strings.Split(option, "=")
			if !generator.ValidateOption(lang, s[0]) {
				err = fmt.Errorf("Unknown option '%s' for %s", s[0], lang)
			}
			if len(s) == 1 {
				options[s[0]] = ""
			} else {
				options[s[0]] = s[1]
			}
		}
	}
	return
}

// logv prints the message if in verbose mode.
func logv(msg string) {
	if globals.Verbose {
		fmt.Println(msg)
	}
}
