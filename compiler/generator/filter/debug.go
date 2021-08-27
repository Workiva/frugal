package filter

import "fmt"

var (
	debug = false
)

func SetDebug(v bool) {
	debug = v
}

func debugPrintf(fmtStr string, args ...interface{}) {
	if !debug {
		return
	}

	fmt.Printf(fmtStr, args...)
}

func debugPrintln(str string) {
	if !debug {
		return
	}

	fmt.Println(str)
}
