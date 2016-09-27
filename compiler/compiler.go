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
	File               string // Frugal file to generate
	Gen                string // Language to generate
	Out                string // Output location for generated code
	Delim              string // Token delimiter for scope topics
	RetainIntermediate bool   // Do not clean up generated intermediate IDL
	DryRun             bool   // Do not generate code
	Recurse            bool   // Generate includes
	Verbose            bool   // Verbose mode
}

// Compile parses the Frugal IDL and generates code for it, returning an error
// if something failed.
func Compile(options Options) error {
	defer globals.Reset()
	globals.TopicDelimiter = options.Delim
	globals.Gen = options.Gen
	globals.Out = options.Out
	globals.DryRun = options.DryRun
	globals.Recurse = options.Recurse
	globals.Verbose = options.Verbose
	globals.FileDir = filepath.Dir(options.File)

	defer func() {
		if !options.RetainIntermediate {
			// Clean up intermediate IDL.
			for _, file := range globals.IntermediateIDL {
				// Only try to remove if file still exists.
				if _, err := os.Stat(file); err == nil {
					logv(fmt.Sprintf("Removing intermediate Thrift file %s", file))
					if err := os.Remove(file); err != nil {
						fmt.Printf("Failed to remove intermediate IDL %s\n", file)
					}
				}
			}
		}
	}()

	absFile, err := filepath.Abs(options.File)
	if err != nil {
		return err
	}

	_, err = compile(absFile, strings.HasSuffix(absFile, ".thrift"), true)
	return err
}

// compile parses the Frugal or Thrift IDL and generates code for it, returning
// an error if something failed.
func compile(file string, isThrift, generate bool) (*parser.Frugal, error) {
	var (
		gen    = globals.Gen
		out    = globals.Out
		dryRun = globals.DryRun
	)

	if frugal, ok := globals.CompiledFiles[file]; ok {
		// This file has already been compiled, skip it.
		return frugal, nil
	}

	// Ensure Frugal file exists.
	if !exists(file) {
		return nil, fmt.Errorf("Frugal file not found: %s\n", file)
	}

	// Process options for specific generators.
	lang, options, err := cleanGenParam(gen)
	if err != nil {
		return nil, err
	}

	// Resolve Frugal generator.
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
		return nil, fmt.Errorf("Invalid gen value %s", gen)
	}

	// Parse the Frugal file.
	logv(fmt.Sprintf("Parsing %s", file))
	frugal, err := parser.ParseFrugal(file)
	if err != nil {
		return nil, err
	}
	globals.CompiledFiles[file] = frugal

	if out == "" {
		out = g.DefaultOutputDir()
	}
	fullOut := g.GetOutputDir(out, frugal)
	if err := os.MkdirAll(out, 0777); err != nil {
		return nil, err
	}

	for _, include := range frugal.Thrift.Includes {
		generateInclude(frugal, include)
	}

	if dryRun || !generate {
		return frugal, nil
	}

	// Generate Frugal code.
	logv(fmt.Sprintf("Generating \"%s\" Frugal code for %s", lang, frugal.File))
	return frugal, g.Generate(frugal, fullOut)
}

// exists determines if the file at the given path exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func generateInclude(frugal *parser.Frugal, incl *parser.Include) (string, error) {
	contents := ""
	// Recurse on includes
	include := incl.Value
	if !strings.HasSuffix(include, ".thrift") && !strings.HasSuffix(include, ".frugal") {
		return "", fmt.Errorf("Bad include name: %s", include)
	}

	file := filepath.Join(frugal.Dir, include)
	parsed, err := compile(file, strings.HasSuffix(include, ".thrift"), globals.Recurse)
	if err != nil {
		return "", fmt.Errorf("Include %s: %s", file, err)
	}

	// Lop off extension (.frugal or .thrift)
	includeBase := include[:len(include)-7]

	// Lop off path
	includeName := filepath.Base(includeBase)

	frugal.ParsedIncludes[includeName] = parsed

	// Replace .frugal with .thrift
	include = includeBase + ".thrift"
	contents += fmt.Sprintf("include \"%s\"\n", include)
	contents += "\n"
	return contents, nil
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
