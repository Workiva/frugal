package crossrunner

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// Default timeout in seconds for client/server configurations without a defined timeout
	DefaultTimeout     = 7
	TestFailure        = 101
	CrossrunnerFailure = 102
)

// getExpandedConfigs takes a client/server at the language level and the options
// associated with that client/server and returns a list of unique configs.
func getExpandedConfigs(options options, test languages) (apps []config) {
	app := new(config)

	// Loop through each transport and protocol to construct expanded list
	for _, transport := range options.Transports {
		for _, protocol := range options.Protocols {
			app.Name = test.Name
			app.Protocol = protocol
			app.Transport = transport
			app.Command = append(test.Command, options.Command...)
			app.Workdir = test.Workdir
			app.Timeout = DefaultTimeout * time.Second
			if options.Timeout != 0 {
				app.Timeout = options.Timeout
			}
			apps = append(apps, *app)
		}
	}
	return apps
}

// GetAvailablePort returns an available port.
func GetAvailablePort() (int, error) {
	// Passing 0 allows the OS to select an available port
	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		// If unavailable, skip port
		return GetAvailablePort()
	}
	defer conn.Close()
	// conn.Addr().String returns something like "[::]:49856", trim the first 5 chars
	port, err := strconv.Atoi(conn.Addr().String()[5:])
	return port, err
}

// getCommand returns a Cmd struct used to execute a client or server and a
// nicely formatted string for verbose loggings
func getCommand(config config, port int) (cmd *exec.Cmd, formatted string) {
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
	cmd.Stdout = config.Logs
	cmd.Stderr = config.Logs

	// Nicely format command here for use at the top of each log file
	formatted = fmt.Sprintf("%s %s", command, strings.Trim(fmt.Sprint(args), "[]"))

	return cmd, formatted
}
