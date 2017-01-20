package crossrunner

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

const (
	// Default timeout in seconds for client/server configutions without a defined timeout
	DEFAULT_TIMEOUT     = 7
	TEST_FAILURE        = 101
	CROSSRUNNER_FAILURE = 102
)

func getList(options options, test Languages) (apps []config) {
	app := new(config)

	// Loop through each transport and protocol to construct expanded list
	for _, transport := range options.Transports {
		for _, protocol := range options.Protocols {
			app.Name = test.Name
			app.Protocol = protocol
			app.Transport = transport
			app.Command = append(test.Command, options.Command...)
			app.Workdir = test.Workdir
			app.Timeout = DEFAULT_TIMEOUT
			if options.Timeout != 0 {
				app.Timeout = options.Timeout
			}
			apps = append(apps, *app)
		}
	}
	return apps
}

// Return next available port
func CheckPort(port uint64) uint64 {
	// Check if port is available
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// If unavailable, skip port
		return CheckPort(port + 1)
	}
	conn.Close()
	return port
}

// getCommand returns a Cmd struct used to execute a client or server and a
// nicely formatted string for verbose logging
func getCommand(config config, port uint64) (cmd *exec.Cmd, formatted string) {
	var args []string

	command := config.Command[0]
	// Not sure if we need to check that the slice is longer than 1
	args = config.Command[1:]

	args = append(args, []string{
		fmt.Sprintf("--protocol=%s", config.Protocol),
		fmt.Sprintf("--transport=%s", config.Transport),
		fmt.Sprintf("--port=%v", port),
	}...)

	cmd = exec.Command(command, args...)
	cmd.Dir = config.Workdir
	cmd.Stdout = config.Logs.File
	cmd.Stderr = config.Logs.File

	// Nicely format command here for use at the top of each log file
	formatted = fmt.Sprintf("%s %s", command, strings.Trim(fmt.Sprint(args), "[]"))

	return cmd, formatted
}
