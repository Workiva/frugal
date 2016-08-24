package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/generator/dartlang"
	"github.com/Workiva/frugal/compiler/generator/golang"
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

// warnGenWithoutFrugal prints a warning if generating code with thrift
// when a gen_with_frugal option exists
// TODO: Remove this once gen_with frugal is no longer experimental
// and is the default.
func warnGenWithoutFrugal(genWithFrugal bool) {
	if globals.GenWithFrugalWarn {
		return
	}
	if genWithFrugal {
		globals.PrintWarning(
			"Generating Thrift code with Frugal. If you encounter problems, file a " +
				"GitHub issue and generate your\ncode with \"gen_with_frugal=false\" to " +
				"use the Thrift compiler instead.")
	} else {
		globals.PrintWarning(
			"Consider using the \"gen_with_frugal\" language option " +
				"to have Frugal generate code in place of Thrift.\nThis is an " +
				"experimental feature. Please file a GitHub issue if you encounter " +
				"problems.")
	}
	globals.GenWithFrugalWarn = true
}

// compile parses the Frugal or Thrift IDL and generates code for it, returning
// an error if something failed.
func compile(file string, isThrift, generate bool) (*parser.Frugal, error) {
	var (
		gen    = globals.Gen
		out    = globals.Out
		dryRun = globals.DryRun
		dir    = filepath.Dir(file)
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

	// Gen with Frugal by default.
	genWithFrugal := true
	if genWithFrugalStr, ok := options["gen_with_frugal"]; ok {
		if gen, err := strconv.ParseBool(genWithFrugalStr); err != nil {
			return nil, fmt.Errorf("Invalid value '%s' for gen_with_frugal", genWithFrugalStr)
		} else {
			genWithFrugal = gen
		}
	}

	// Resolve Frugal generator.
	var g generator.ProgramGenerator
	switch lang {
	case "dart":
		// TODO: Remove this once gen_with_frugal is no longer experimental
		// and is the default
		warnGenWithoutFrugal(genWithFrugal)
		g = generator.NewProgramGenerator(dartlang.NewGenerator(options, genWithFrugal), false)
	case "go":
		// Make sure the package prefix ends with a "/"
		if package_prefix, ok := options["package_prefix"]; ok {
			if package_prefix != "" && !strings.HasSuffix(package_prefix, "/") {
				options["package_prefix"] = package_prefix + "/"
			}
		}

		// TODO: Remove this once gen_with frugal is no longer experimental
		// and is the default.
		warnGenWithoutFrugal(genWithFrugal)
		g = generator.NewProgramGenerator(golang.NewGenerator(options), false)
	case "java":
		// TODO: Remove this once gen_with frugal is no longer experimental
		// and is the default.
		warnGenWithoutFrugal(genWithFrugal)
		g = generator.NewProgramGenerator(java.NewGenerator(options), true)
	case "py":
		// TODO: Remove this once gen_with frugal is no longer experimental
		// and is the default.
		warnGenWithoutFrugal(genWithFrugal)
		g = generator.NewProgramGenerator(python.NewGenerator(options), true)
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

	if genWithFrugal {
		// If not using frugal, add parsed includes here
		// preserve what thrift does to keep ordering in the file
		for _, include := range frugal.Thrift.Includes {
			generateInclude(frugal, include)
		}
	}

	if !genWithFrugal && !isThrift {
		// Generate intermediate Thrift IDL for Frugal. If this is already a
		// .thrift file, do not generate an intermediate IDL.
		logv(fmt.Sprintf("Generating intermediate Thrift file %s",
			filepath.Join(dir, fmt.Sprintf("%s.thrift", frugal.Name))))
		idlFile, err := generateThriftIDL(dir, frugal)
		if err != nil {
			return nil, err
		}
		file = idlFile
	}

	if dryRun || !generate {
		return frugal, nil
	}

	if !genWithFrugal {
		// Generate Thrift code.
		logv(fmt.Sprintf("Generating \"%s\" Thrift code for %s", lang, file))
		if err := generateThrift(frugal, dir, file, out, removeGenWithFrugalOption(gen)); err != nil {
			return nil, err
		}
	}

	// Generate Frugal code.
	logv(fmt.Sprintf("Generating \"%s\" Frugal code for %s", lang, frugal.File))
	return frugal, g.Generate(frugal, fullOut, genWithFrugal)
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

// removeGenWithFrugalOption removes the gen_with_frugal language option from
// the gen string, if present, so that it doesn't cause issues with the Thrift
// compiler when set to false.
// TODO: Remove this once the Thrift compiler is no longer used.
func removeGenWithFrugalOption(gen string) string {
	if !strings.Contains(gen, ":") {
		return gen
	}
	s := strings.Split(gen, ":")
	lang := s[0]
	optionsStr := s[1]
	options := strings.Split(optionsStr, ",")
	cleaned := ""
	prefix := ""
	for _, option := range options {
		if !strings.HasPrefix(option, "gen_with_frugal") {
			cleaned += prefix + option
			prefix = ","
		}
	}

	if cleaned == "" {
		return lang
	}
	return fmt.Sprintf("%s:%s", lang, cleaned)
}

// logv prints the message if in verbose mode.
func logv(msg string) {
	if globals.Verbose {
		fmt.Println(msg)
	}
}
