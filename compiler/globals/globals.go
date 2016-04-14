package globals

import "time"

const Version = "1.2.0-dev"

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
}
