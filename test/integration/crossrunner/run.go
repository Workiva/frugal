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
	var err error
	// Create client/server log files
	if err = createLogs(pair); err != nil {
		log.Debugf("Failed to create logs for % client and %s server", pair.Client.Name, pair.Server.Name)
		reportCrossrunnerFailure(pair, err)
		return
	}
	defer pair.Client.Logs.Close()
	defer pair.Server.Logs.Close()

	// Get server and client command structs
	server, serverCmd := getCommand(pair.Server, port)
	client, clientCmd := getCommand(pair.Client, port)

	// write server log header
	log.Debug(serverCmd)

	if err = writeFileHeader(pair.Server.Logs, serverCmd, pair.Server.Workdir,
		pair.Server.Timeout, pair.Client.Timeout); err != nil {
		log.Debugf("Failed to write header to %", pair.Server.Logs.Name())
		reportCrossrunnerFailure(pair, err)
		return
	}

	// start the server
	sStartTime := time.Now()
	if err = server.Start(); err != nil {
		log.Debugf("Failed to start %s server", pair.Server.Name)
		reportCrossrunnerFailure(pair, err)
		return
	}
	// Defer stopping the server to ensure the process is killed on exit
	defer func() {
		if err = server.Process.Kill(); err != nil {
			reportCrossrunnerFailure(pair, err)
			log.Info("Failed to kill " + pair.Server.Name + " server.")
			return
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
		if err = writeServerTimeout(pair.Server.Logs, pair.Server.Name); err != nil {
			log.Debugf("Failed to write server timeout to %s", pair.Server.Logs.Name)
			reportCrossrunnerFailure(pair, err)
			return
		}
		pair.ReturnCode = TestFailure
		pair.Err = errors.New("Server has not started within the specified timeout")
		log.Debug(pair.Server.Name + " server not started within specified timeout")
		// Even though the healthcheck server hasn't started, the process has.
		// Process is killed in the deferred function above
		return
	}

	// write client log header
	if err = writeFileHeader(pair.Client.Logs, clientCmd, pair.Client.Workdir,
		pair.Server.Timeout, pair.Client.Timeout); err != nil {
		log.Debugf("Failed to write header to %s", pair.Client.Logs.Name())
		reportCrossrunnerFailure(pair, err)
		return
	}

	// start client
	done := make(chan error, 1)
	log.Debug(clientCmd)
	cStartTime := time.Now()

	if err = client.Start(); err != nil {
		log.Debugf("Failed to start %s client", pair.Client.Name)
		pair.ReturnCode = TestFailure
		pair.Err = err
	}

	go func() {
		done <- client.Wait()
	}()

	select {
	case <-time.After(pair.Client.Timeout * time.Second):
		// TODO: It's a bit annoying to have this message duplicated in the
		// unexpected_failures.log. Is there a better way to report this?
		if err = writeClientTimeout(pair, pair.Client.Name); err != nil {
			log.Debugf("Failed to write timeout error to %s", pair.Client.Logs.Name())

			reportCrossrunnerFailure(pair, err)
			return
		}

		if err = client.Process.Kill(); err != nil {
			log.Infof("Error killing %s", pair.Client.Name)
			reportCrossrunnerFailure(pair, err)
			return
		}
		pair.ReturnCode = TestFailure
		pair.Err = errors.New("Client has not completed within the specified timeout")
		break
	case err := <-done:
		if err != nil {
			log.Debugf("Error in %s client", pair.Client.Name)
			pair.ReturnCode = TestFailure
			pair.Err = err
		}
	}

	// write log footers
	if err = writeFileFooter(pair.Client.Logs, time.Since(cStartTime)); err != nil {
		log.Debugf("Failed to write footer to %s", pair.Client.Logs.Name())
		reportCrossrunnerFailure(pair, err)
		return
	}
	if err = writeFileFooter(pair.Server.Logs, time.Since(sStartTime)); err != nil {
		log.Debugf("Failed to write footer to %s", pair.Client.Logs.Name())
		reportCrossrunnerFailure(pair, err)
		return
	}
}

// reportCrossrunnerFailure is used in the error case when something goes wrong
// in the crossrunner.
func reportCrossrunnerFailure(pair *Pair, err error) {
	log.Info("Unexpected error: " + err.Error())
	pair.ReturnCode = CrossrunnerFailure
	pair.Err = err
	return
}
