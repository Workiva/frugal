package compiler

import (
	"fmt"
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
	File      string // Frugal file to generate
	Gen       string // Language to generate
	Out       string // Output location for generated code
	Delim     string // Token delimiter for scope topics
	DryRun    bool   // Do not generate code
	Recurse   bool   // Generate includes
	Verbose   bool   // Verbose mode
	UseVendor bool   // Do not generate code for vendored includes
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
	globals.UseVendor = options.UseVendor
	globals.FileDir = filepath.Dir(options.File)

	absFile, err := filepath.Abs(options.File)
	if err != nil {
		return err
	}

	frugal, err := parseFrugal(absFile)
	if err != nil {
		return err
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

	// Process options for specific generators.
	lang, options, err := CleanGenParam(gen)
	if err != nil {
		return err
	}

	// TODO: Address as needed when more languages support vendoring.
	if globals.UseVendor && lang != "go" {
		return fmt.Errorf("-use-vendor not supported by %s", lang)
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

	for _, inclFrugal := range f.ParsedIncludes {
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

// CleanGenParam processes a string that includes an optional trailing
// options set.  Format: <language>:<name>=<value>,<name>=<value>,...
// TODO: unexport once plugin workaround is no longer needed in main.go.
func CleanGenParam(gen string) (lang string, options map[string]string, err error) {
	lang = gen
	options = make(map[string]string)
	if strings.Contains(gen, ":") {
		s := strings.SplitN(gen, ":", 2)
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
