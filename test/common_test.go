package test

import (
	"bufio"
	"os"
	"testing"
)

const (
	outputDir               = "out"
	delim                   = "."
	validFile               = "idl/valid.frugal"
	invalidFile             = "idl/invalid.frugal"
	duplicateMethodArgIds   = "idl/duplicate_arg_ids.frugal"
	duplicateStructFieldIds = "idl/duplicate_field_ids.frugal"
	frugalGenFile           = "idl/variety.frugal"
)

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
