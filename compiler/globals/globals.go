package globals

import (
	"fmt"
	"time"

	"github.com/Workiva/frugal/compiler/parser"
)

// Version of the Frugal compiler.
const Version = "2.2.3"

// Global variables.
var (
	TopicDelimiter = "."
	Gen            string
	Out            string
	FileDir        string
	DryRun         bool
	Recurse        bool
	Verbose        bool
	Now            = time.Now()
	CompiledFiles  = make(map[string]*parser.Frugal)
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
	Now = time.Now()
	CompiledFiles = make(map[string]*parser.Frugal)
}

// PrintWarning prints the given message to stdout in yellow font.
func PrintWarning(msg string) {
	fmt.Println("\x1b[33m" + msg + "\x1b[0m")
}
