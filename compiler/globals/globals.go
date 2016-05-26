package globals

import (
	"fmt"
	"time"
)

const Version = "1.4.1"

var (
	TopicDelimiter  string = "."
	Gen             string
	Out             string
	FileDir         string
	DryRun          bool
	Recurse         bool
	Verbose         bool
	Now             = time.Now()
	IntermediateIDL = []string{}

	// TODO: Remove once gen_with_frugal is the default.
	GenWithFrugalWarn bool
)

func Reset() {
	TopicDelimiter = "."
	Gen = ""
	Out = ""
	FileDir = ""
	DryRun = false
	Recurse = false
	Verbose = false
	Now = time.Now()
	IntermediateIDL = []string{}
	GenWithFrugalWarn = false
}

// PrintWarning prints the given message to stdout in yellow font.
func PrintWarning(msg string) {
	fmt.Println("\x1b[33m" + msg + "\x1b[0m")
}
