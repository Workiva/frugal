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
	Key       string  `json:"key"`
	ParentKey string  `json:"parent_key"`
	Type      string  `json:"type"`
	Grammer   grammer `json:"grammer"`
	Meta      meta    `json:"meta"`
}

type grammer struct {
	Signature string      `json:"signature"`
	G         interface{} `json:"stuff"`
}

type meta struct {
	Line int64  `json:"line"`
	URI  string `json:"uri"`
}

func generateSemVerAudit(f *parser.Frugal) error {
	var exp []*export
	var e *export
	for _, c := range f.Constants {
		e = &export{
			Key:     c.Name,
			Type:    f.UnderlyingType(c.Type).Name,
			Grammer: grammer{Signature: "derp", G: c.Value},
		}
		exp = append(exp, e)
	}

	for _, c := range f.Enums {
		e = &export{
			Key:  c.Name,
			Type: "Enum",
		}
		exp = append(exp, e)
		for _, v := range c.Values {
			e = &export{
				Key:       v.Name,
				ParentKey: c.Name,
				Type:      "int",
				Grammer:   grammer{Signature: "derp", G: v.Value},
			}
			exp = append(exp, e)
		}
	}

	for _, c := range f.Exceptions {
		e = &export{
			Key:  c.Name,
			Type: c.Type.String(),
		}
		exp = append(exp, e)
		for _, field := range c.Fields {
			e = &export{
				Key:       field.Name,
				ParentKey: c.Name,
				Type:      f.UnderlyingType(field.Type).Name,
				Grammer:   grammer{Signature: field.Modifier.String(), G: field.Default},
			}
			exp = append(exp, e)
		}
	}

	for _, c := range f.Includes {
		e = &export{
			Key:     c.Name,
			Type:    "Includes",
			Grammer: grammer{Signature: "derp", G: c.Value},
		}
		exp = append(exp, e)
	}

	for _, c := range f.Namespaces {
		e = &export{
			Key:     c.Value,
			Type:    "Namespace",
			Grammer: grammer{Signature: "derp", G: c.Value},
		}
		exp = append(exp, e)
	}

	// TODO recursive for this
	for k, v := range f.ParsedIncludes {
		e = &export{
			Key:  k,
			Type: v.Name,
		}
		exp = append(exp, e)
	}

	// TODO recursive for this
	for _, c := range f.Scopes {
		e = &export{
			Key:  c.Name,
			Type: "Scope",
		}
		exp = append(exp, e)
	}

	// TODO recursive for this
	for _, c := range f.Services {
		e = &export{
			Key:  c.Name,
			Type: "Service",
		}
		exp = append(exp, e)
	}

	for _, c := range f.Structs {
		e = &export{
			Key:     c.Name,
			Type:    "Struct",
			Grammer: grammer{Signature: "derp", G: c.Type.String()},
		}
		exp = append(exp, e)
		for _, field := range c.Fields {
			e = &export{
				Key:       field.Name,
				ParentKey: c.Name,
				Type:      f.UnderlyingType(field.Type).Name,
				Grammer:   grammer{Signature: field.Modifier.String(), G: field.Default},
			}
			exp = append(exp, e)
		}
	}

	for _, c := range f.Typedefs {
		e = &export{
			Key:     c.Name,
			Type:    "Typedef",
			Grammer: grammer{Signature: "derp", G: f.UnderlyingType(c.Type).Name},
		}
		exp = append(exp, e)
	}

	for _, c := range f.Unions {
		e = &export{
			Key:     c.Name,
			Type:    "Union",
			Grammer: grammer{Signature: "derp", G: c.Type.String()},
		}
		exp = append(exp, e)
		for _, field := range c.Fields {
			e = &export{
				Key:       field.Name,
				ParentKey: c.Name,
				Type:      f.UnderlyingType(field.Type).Name,
				Grammer:   grammer{Signature: field.Modifier.String(), G: field.Default},
			}
			exp = append(exp, e)
		}
	}

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
