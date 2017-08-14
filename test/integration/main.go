/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"

	"fmt"
	"github.com/Workiva/frugal/test/integration/crossrunner"
	"os/exec"
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
