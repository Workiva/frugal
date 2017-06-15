package main

import (
	"flag"

	"github.com/Workiva/frugal/test/integration/crossrunner"
	"os/exec"
	"fmt"
	"strings"
)

func main() {
	// These are properly configured for Frugal.
	var testDefinitions = flag.String("tests", "tests.json", "Location of json test definitions")
	var outDir = flag.String("outDir", "log", "Output directory of crossrunner logs")

	flag.Parse()

	if err := crossrunner.Run(testDefinitions, outDir, getCommand); err != nil {
		panic(err)
	}

}

func getCommand(config crossrunner.Config, port int) (cmd *exec.Cmd, formatted string) {

	command := config.Command[0]

	args := append(config.Command[1:],
		fmt.Sprintf("--protocol=%s", config.Protocol),
		fmt.Sprintf("--transport=%s", config.Transport),
		fmt.Sprintf("--port=%v", port),
	)

	cmd = exec.Command(command, args...)
	cmd.Dir = config.Workdir
	cmd.Stdout = config.Logs
	cmd.Stderr = config.Logs

	// Nicely format command here for use at the top of each log file
	formatted = fmt.Sprintf("%s %s", command, strings.Trim(fmt.Sprint(args), "[]"))

	return cmd, formatted
}