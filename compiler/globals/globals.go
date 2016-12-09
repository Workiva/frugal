package globals

import (
	"fmt"
	"time"

	"github.com/Workiva/frugal/compiler/parser"
)

// Version of the Frugal compiler.
const Version = "1.23.0"

// Global variables.
var (
	TopicDelimiter  = "."
	Gen             string
	Out             string
	FileDir         string
	DryRun          bool
	Recurse         bool
	Verbose         bool
	UseVendor       bool
	Now             = time.Now()
	IntermediateIDL = []string{}
	CompiledFiles   = make(map[string]*parser.Frugal)

	// TODO: Remove once gen_with_frugal is the default.
	GenWithFrugalWarn bool
)

// Reset global variables to initial state.
func Reset() {
	TopicDelimiter = "."
	Gen = ""
	Out = ""
	FileDir = ""
	DryRun = false
	Recurse = false
	Verbose = false
	UseVendor = false
	Now = time.Now()
	IntermediateIDL = []string{}
	CompiledFiles = make(map[string]*parser.Frugal)
	GenWithFrugalWarn = false
}

// PrintWarning prints the given message to stdout in yellow font.
func PrintWarning(msg string) {
	fmt.Println("\x1b[33m" + msg + "\x1b[0m")
}
