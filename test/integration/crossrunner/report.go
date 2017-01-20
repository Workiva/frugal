package crossrunner

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// LogFile is a client or server log with a name and a pointer to the file
type LogFile struct {
	Filepath string
	File     *os.File
}

// getLogPaths creates client and server log files with the following format:
// log/clientName-serverName_transport_protocol_role.log
func assignLogPaths(pair *Pair) (err error) {
	pair.Client.Logs = &LogFile{
		Filepath: fmt.Sprintf("log/%s-%s_%s_%s_%s.log",
			pair.Client.Name,
			pair.Server.Name,
			pair.Client.Protocol,
			pair.Client.Transport,
			"client"),
		File: nil,
	}

	pair.Server.Logs = &LogFile{
		Filepath: fmt.Sprintf("log/%s-%s_%s_%s_%s.log",
			pair.Client.Name,
			pair.Server.Name,
			pair.Server.Protocol,
			pair.Server.Transport,
			"server"),
		File: nil,
	}

	if pair.Client.Logs.File, err = createFile(pair.Client.Logs); err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
		return err
	}
	if pair.Server.Logs.File, err = createFile(pair.Server.Logs); err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
		return err
	}

	return nil
}

// createFile creates a file with the given name
func createFile(file *LogFile) (path *os.File, err error) {
	path, err = os.Create(file.Filepath)
	return path, err
}

// writeFileHeader writes the metadata associated with each run to the header
// in a LogFile
func writeFileHeader(file *os.File, cmd, dir string, delay, timeout time.Duration) (err error) {
	header := fmt.Sprintf("%v\nExecuting: %s\nDirectory: %s\nServer Timeout: %s\n Client Timeout: %s\n",
		GetTimestamp(),
		cmd,
		dir,
		delay*time.Second,
		timeout*time.Second,
	)
	header += breakLine()

	if _, err := file.WriteString(header); err != nil {
		return err
	}
	return nil
}

// breakLine returns a formatted separator line
func breakLine() string {
	return fmt.Sprintf("\n==================================================================================\n\n")
}

// starBreak returns 4 rows of stars
// Used as a break between pairs in unexpected_failures.log
func starBreak() string {
	stars := "**********************************************************************************\n"
	return fmt.Sprintf("\n\n\n%s%s%s%s\n\n", stars, stars, stars, stars)
}

// writeFileFooter writes execution time and closes the file
func writeFileFooter(file *os.File, executionTime time.Duration) (err error) {
	// defer file.Close()
	footer := breakLine()
	footer += fmt.Sprintf("Test execution took %.2f seconds\n", executionTime.Seconds())
	footer += GetTimestamp()
	_, err = file.WriteString(footer)
	return
}

// GetTimestamp returns the current time
func GetTimestamp() string {
	return time.Now().Format(time.UnixDate)
}

// writeConsoleHeader prints a header for all test configuration results to the console.
func WriteConsoleHeader() {
	fmt.Printf("%-35s%-15s%-25s%-20s\n",
		"Client-Server",
		"Protocol",
		"Transport",
		"Result")
	fmt.Printf(breakLine())
}

// writePairResult prints a formatted pair result to the console.
func WritePairResult(pair *Pair) {
	var result string
	if pair.ReturnCode == 0 {
		result = "success"
	} else {
		result = "FAILURE"
	}

	fmt.Printf("%-35s%-15s%-25s%-20s\n",
		fmt.Sprintf("%s-%s",
			pair.Client.Name,
			pair.Server.Name),
		pair.Client.Protocol,
		pair.Client.Transport,
		result)
}

// writeConsoleFooter writes the metadata associated with the test suite to the console.
func WriteConsoleFooter(failed, total uint64, runtime time.Duration) {
	fmt.Printf(breakLine())
	fmt.Println("Full logs for each test can be found at:")
	fmt.Println("  test/integration/log/client-server_protocol_transport_client.log")
	fmt.Println("  test/integration/log/client-server_protocol_transport_server.log")
	fmt.Printf("%d of %d tests failed.\n", failed, total)
	fmt.Printf("Test execution took %.1f seconds\n", runtime.Seconds())
	fmt.Printf(GetTimestamp())
}

// Append to failures adds a the client and server logs from a failed
// configuration to the unexpected_failure.log file
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
	contents += GetLogs(pair.Client.Logs.Filepath)
	contents += fmt.Sprintf(breakLine())
	// Add Server logs
	contents += "================================= SERVER LOG =====================================\n"
	contents += GetLogs(pair.Server.Logs.Filepath)
	// Write break between pairs for readability
	contents += starBreak()

	// Open unexpected_failures.log
	f, err := os.OpenFile(failLog, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Append to unexpected_failures.log
	if _, err = f.WriteString(contents); err != nil {
		return err
	}

	return nil
}

// GetLogs reads the contents of a file and returns them as a string
func GetLogs(file string) string {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
