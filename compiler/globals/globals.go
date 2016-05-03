package globals

import "time"

const Version = "1.3.0"

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
