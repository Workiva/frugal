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

package crossrunner

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// createLogs creates client and server log files with the following format:
// log/clientName-serverName_transport_protocol_role.log.
func createLogs(pair *Pair) (err error) {
	pair.Client.Logs, err = os.Create(fmt.Sprintf("log/%s-%s_%s_%s_%s.log",
		pair.Client.Name,
		pair.Server.Name,
		pair.Client.Protocol,
		pair.Client.Transport,
		"client"))
	if err != nil {
		return err
	}

	pair.Server.Logs, err = os.Create(fmt.Sprintf("log/%s-%s_%s_%s_%s.log",
		pair.Client.Name,
		pair.Server.Name,
		pair.Server.Protocol,
		pair.Server.Transport,
		"server"))
	if err != nil {
		return err
	}

	return nil
}

// writeFileHeader writes the metadata associated with each run to the header
// in a LogFile.
func writeFileHeader(file *os.File, cmd, dir string, delay, timeout time.Duration) error {
	header := fmt.Sprintf("%s\nExecuting: %s\nDirectory: %s\nServer Timeout: %s\nClient Timeout: %s\n",
		GetTimestamp(),
		cmd,
		dir,
		delay*time.Second,
		timeout*time.Second,
	)
	header += breakLine()

	_, err := file.WriteString(header)
	return err
}

// breakLine returns a formatted separator line.
func breakLine() string {
	return fmt.Sprint("\n==================================================================================\n\n")
}

// starBreak returns 4 rows of stars.
// Used as a break between pairs in unexpected_failures.log.
func starBreak() string {
	stars := "**********************************************************************************\n"
	return fmt.Sprintf("\n\n\n%s%s%s%s\n\n", stars, stars, stars, stars)
}

func writeClientTimeout(pair *Pair, role string) error {
	timeout := breakLine()
	timeout += fmt.Sprintf("%s client exceeded specified timeout", role)
	_, err := pair.Client.Logs.WriteString(timeout)
	if err != nil {
		return err
	}
	_, err = pair.Server.Logs.WriteString(timeout)
	return err
}

func writeServerTimeout(file *os.File, role string) error {
	timeout := breakLine()
	timeout += fmt.Sprintf("%s server has not started within specified timeout", role)
	_, err := file.WriteString(timeout)
	return err
}

// writeFileFooter writes execution time and closes the file.
func writeFileFooter(file *os.File, executionTime time.Duration) error {
	footer := breakLine()
	footer += fmt.Sprintf("Test execution took %.2f seconds\n", executionTime.Seconds())
	footer += GetTimestamp()
	_, err := file.WriteString(footer)
	return err
}

// WriteCustomData writes any string to a file.
func WriteCustomData(file string, info string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(info); err != nil {
		return err
	}
	return nil
}

// GetTimestamp returns the current time.
func GetTimestamp() string {
	return time.Now().Format(time.UnixDate)
}

// PrintConsoleHeader prints a header for all test configuration results to the console.
func PrintConsoleHeader() {
	fmt.Printf("%-35s%-15s%-25s%-20s\n",
		"Client-Server",
		"Protocol",
		"Transport",
		"Result")
	fmt.Print(breakLine())
}

// PrintPairResult prints a formatted pair result to the console.
func PrintPairResult(pair *Pair) {
	var result string
	if pair.ReturnCode == 0 {
		result = "success"
	} else if pair.ReturnCode == CrossrunnerFailure {
		// Colorize failures red - this probably only works in Linux
		result = "\x1b[31;1mCROSSRUNNER FAILURE\x1b[37;1m"
	} else {
		result = "\x1b[31;1mFAILURE\x1b[37;1m"
	}

	fmt.Printf("%-35s%-15s%-25s%-20s\n",
		fmt.Sprintf("%s-%s",
			pair.Client.Name,
			pair.Server.Name),
		pair.Client.Protocol,
		pair.Client.Transport,
		result)
}

// PrintConsoleFooter writes the metadata associated with the test suite to the console.
func PrintConsoleFooter(failed int, total uint64, runtime time.Duration) {
	fmt.Print(breakLine())
	fmt.Println("Full logs for each test can be found at:")
	// TODO: allow configurability of log location
	fmt.Println("  test/integration/log/client-server_protocol_transport_client.log")
	fmt.Println("  test/integration/log/client-server_protocol_transport_server.log")
	fmt.Printf("%d of %d tests failed.\n", failed, total)
	fmt.Printf("Test execution took %.1f seconds\n", runtime.Seconds())
	fmt.Print(GetTimestamp())
}

// Append to failures adds a the client and server logs from a failed
// configuration to the unexpected_failure.log file.
func AppendToFailures(failLog string, pair *Pair) (err error) {
	// Add Header
	contents := fmt.Sprintf("Client - %s\nServer - %s\nProtocol - %s\nTransport - %s\n",
		pair.Client.Name,
		pair.Server.Name,
		pair.Server.Protocol,
		pair.Server.Transport,
	)
	// Add Client logs
	contents += "================================= CLIENT LOG =====================================\n"
	contents += getFileContents(pair.Client.Logs.Name())
	contents += fmt.Sprint(breakLine())
	// Add Server logs
	contents += "================================= SERVER LOG =====================================\n"
	contents += getFileContents(pair.Server.Logs.Name())
	// Write break between pairs for readability
	contents += starBreak()

	// Open unexpected_failures.log
	f, err := os.OpenFile(failLog, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Append to unexpected_failures.log
	_, err = f.WriteString(contents)
	return err
}

// GetLogs reads the contents of a file and returns them as a string.
func getFileContents(file string) string {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
