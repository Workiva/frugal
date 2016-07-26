package test

import (
	"bufio"
	"os"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

const (
	outputDir     = "out"
	delim         = "."
	validFile     = "idl/valid.frugal"
	invalidFile   = "idl/invalid.frugal"
	frugalGenFile = "idl/variety.frugal"
)

var (
	languages = []string{"go", "dart", "java"}
)

func TestInvalid(t *testing.T) {
	options := compiler.Options{
		File:  invalidFile,
		Gen:   languages[0] + ":gen_with_frugal=false",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func compareFiles(t *testing.T, expectedPath, generatedPath string) {
	expected, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("Failed to open file %s", expectedPath)
	}
	defer expected.Close()
	generated, err := os.Open(generatedPath)
	if err != nil {
		t.Fatalf("Failed to open file %s", generatedPath)
	}
	defer generated.Close()

	expectedScanner := bufio.NewScanner(expected)
	generatedScanner := bufio.NewScanner(generated)
	line := 1
	for expectedScanner.Scan() {
		generatedScanner.Scan()
		expectedLine := expectedScanner.Text()
		generatedLine := generatedScanner.Text()
		if expectedLine != generatedLine {
			t.Fatalf("Expected line <%s> (%s), generated line <%s> (%s) at line %d",
				expectedLine, expectedPath, generatedLine, generatedPath, line)
		}
		line++
	}

	if generatedScanner.Scan() {
		t.Fatal("Generated has more lines than expected")
	}
}
