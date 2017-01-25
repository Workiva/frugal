package crossrunner

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// RunConfig runs a client against a server.  Client/Server logs are created and
// failures are added to the unexpected_failures.log.  Each result is logged to
// the console.
func RunConfig(pair *Pair, port int) {
	// Get filepaths to write logs to
	err := createLogs(pair)
	if err != nil {
		panic(err)
	}
	defer pair.Client.Logs.Close()
	defer pair.Server.Logs.Close()

	// Get server and client command structs
	server, serverCmd := getCommand(pair.Server, port)
	client, clientCmd := getCommand(pair.Client, port)

	// write server log header
	log.Debug(serverCmd)
	err = writeFileHeader(pair.Server.Logs, serverCmd, pair.Server.Workdir,
		pair.Server.Timeout, pair.Client.Timeout)
	if err != nil {
		pair.ReturnCode = CrossrunnerFailure
		pair.Err = err
		return
	}

	// start the server
	sStartTime := time.Now()
	err = server.Start()
	if err != nil {
		pair.ReturnCode = CrossrunnerFailure
		pair.Err = err
		return
	}
	// Defer stopping the server to ensure the process is killed on exit
	defer func() {
		err = server.Process.Kill()
		if err != nil {
			pair.ReturnCode = CrossrunnerFailure
			pair.Err = err
			log.Info("Failed to kill " + pair.Server.Name + " server.")
		}
	}()
	stimeout := pair.Server.Timeout * time.Millisecond * 1000
	var total time.Duration
	// Poll the server healthcheck until it returns a valid status code or exceeds the timeout
	for total <= stimeout {
		// If the server hasn't started within the specified timeout, fail the test
		resp, err := (http.Get(fmt.Sprintf("http://localhost:%d", port)))
		if err != nil {
			time.Sleep(time.Millisecond * 250)
			total += (time.Millisecond * 250)
			continue
		}
		resp.Close = true
		resp.Body.Close()
		break
	}

	if total >= stimeout {
		// TODO: Add timeout error to server log
		pair.ReturnCode = TestFailure
		pair.Err = errors.New("Server has not started within the specified timeout")
		log.Debug(pair.Server.Name + " server not started within specified timeout")
		// Even though the healthcheck server hasn't started, the process has.
		// Process is killed in the deferred function above
		return
	}

	// write client log header
	err = writeFileHeader(pair.Client.Logs, clientCmd, pair.Client.Workdir,
		pair.Server.Timeout, pair.Client.Timeout)
	if err != nil {
		pair.ReturnCode = CrossrunnerFailure
		pair.Err = err
		return
	}

	// start client
	done := make(chan error, 1)
	log.Debug(clientCmd)
	cStartTime := time.Now()

	err = client.Start()
	if err != nil {
		pair.ReturnCode = TestFailure
		pair.Err = err
	}

	go func() {
		done <- client.Wait()
	}()

	select {
	case <-time.After(pair.Client.Timeout * time.Second):
		// TODO: Add timeout error to client log
		err = client.Process.Kill()
		if err != nil {
			pair.ReturnCode = CrossrunnerFailure
			pair.Err = err
			break
		}
		pair.ReturnCode = TestFailure
		pair.Err = errors.New("Client has not completed within the specified timeout")
		break
	case err := <-done:
		if err != nil {
			panic(err)
			pair.ReturnCode = CrossrunnerFailure
			pair.Err = err
		}
	}

	// write log footers
	err = writeFileFooter(pair.Client.Logs, time.Since(cStartTime))
	if err != nil {
		panic(err)
	}
	err = writeFileFooter(pair.Server.Logs, time.Since(sStartTime))
	if err != nil {
		panic(err)
	}
}
