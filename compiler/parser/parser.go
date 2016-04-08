package parser

import (
	"fmt"
	"github.com/Workiva/frugal/compiler/globals"
	"os"
	"path/filepath"
	"strings"
)

// ParseFrugal parses the given Frugal file into its semantic representation.
func ParseFrugal(filePath string) (*Frugal, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	name, err := getName(file)
	if err != nil {
		return nil, err
	}

	parsed, err := ParseReader(filePath, file)
	if err != nil {
		return nil, err
	}

	frugal := parsed.(*Frugal)
	for _, incl := range frugal.Thrift.Includes {
		// parse all the includes before validating.
		// TODO when this isn't experimental this should be the only place things
		// are parsed
		include := incl.Value
		if !strings.HasSuffix(include, ".thrift") && !strings.HasSuffix(include, ".frugal") {
			return nil, fmt.Errorf("Bad include name: %s", include)
		}

		parsedIncl, err := ParseFrugal(filepath.Join(globals.FileDir, include))
		if err != nil {
			return nil, err
		}

		// Lop off extension (.frugal or .thrift)
		includeBase := include[:len(include)-7]

		// Lop off path
		includeName := filepath.Base(includeBase)

		frugal.ParsedIncludes[includeName] = parsedIncl
	}
	if err := frugal.validate(); err != nil {
		return nil, err
	}
	frugal.Name = name
	frugal.File = filePath
	frugal.Dir = filepath.Dir(file.Name())
	frugal.Path = filePath
	frugal.sort() // For determinism in generated code
	frugal.assignFrugal()

	return frugal, nil
}

func getName(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	parts := strings.Split(info.Name(), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid file: %s", f.Name())
	}
	return parts[0], nil
}
