package crossrunner

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

func RunConfig(pair *Pair, port uint64) {
	// Get filepaths to write logs to
	err := assignLogPaths(pair)
	if err != nil {
		panic(err)
	}
	defer pair.Client.Logs.File.Close()
	defer pair.Server.Logs.File.Close()

	// Get server and client command structs
	server, serverCmd := getCommand(pair.Server, port)
	client, clientCmd := getCommand(pair.Client, port)

	// write server log header
	log.Debug(serverCmd)
	err = writeFileHeader(pair.Server.Logs.File, serverCmd, pair.Server.Workdir,
		pair.Server.Timeout, pair.Client.Timeout)
	if err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
		return
	}

	// start the server
	sStartTime := time.Now()
	err = server.Start()
	if err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
		return
	}
	stimeout := pair.Server.Timeout * time.Millisecond * 1000
	total := (time.Millisecond * 0)
	// Poll the server healthcheck until it returns a valid status code or exceeds the timeout
	for total <= stimeout {
		// If the server hasn't started within the specified timeout, fail the test
		// REVIEW: I think there is an issue with this healthcheck
		resp, err := (http.Get(fmt.Sprintf("http://localhost:%d", port)))
		if err != nil {
			time.Sleep(time.Millisecond * 250)
			total = total + (time.Millisecond * 250)
			continue
		}
		resp.Close = true
		resp.Body.Close()
		break
	}

	if total >= stimeout {
		pair.ReturnCode = TEST_FAILURE
		pair.Err = errors.New("Server has not started within the specified timeout")
		log.Debug(pair.Server.Name + " server not started within specified timeout")
		// Even though the healthcheck server hasn't started, the process has.
		// Kill it.
		err = server.Process.Kill()
		if err != nil {
			pair.ReturnCode = CROSSRUNNER_FAILURE
			log.Info("Failed to kill " + pair.Server.Name + " server.")
		}
		return
	}

	// write client log header
	err = writeFileHeader(pair.Client.Logs.File, clientCmd, pair.Client.Workdir,
		pair.Server.Timeout, pair.Client.Timeout)
	if err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
		return
	}

	// start client
	log.Debug(clientCmd)
	cStartTime := time.Now()

	err = client.Start()
	if err != nil {
		pair.ReturnCode = TEST_FAILURE
		pair.Err = err
	}
	// TODO: Need a timer to kill the config if the client doesn't complete
	// within the specified timeout
	//
	// var timer *time.Timer
	// timer = time.AfterFunc(pair.Client.Timeout*time.Second, func() {
	// 	log.Debug(pair.Client.Name + " client not completed within specified timeout")
	// 	err = client.Process.Kill()
	//  // if the client has already finished, don't worry about killing it
	// if err != nil && err.Error() != "os: process already finished" {
	// 	// If we fail to stop the client, there is a crossrunner failure
	// 	// build some sore of retry logic here?  Not sure how common this case might be.
	// 	pair.ReturnCode = CROSSRUNNER_FAILURE
	//  pair.Err = err
	// }
	// 	pair.ReturnCode = TEST_FAILURE
	// 	pair.Err = errors.New("Client has not completed within specified timeout")
	// })

	err = client.Wait()
	if err != nil {
		pair.ReturnCode = TEST_FAILURE
		pair.Err = err
	}

	// write log footers
	err = writeFileFooter(pair.Client.Logs.File, time.Since(cStartTime))
	if err != nil {
		panic(err)
	}
	err = writeFileFooter(pair.Server.Logs.File, time.Since(sStartTime))
	if err != nil {
		panic(err)
	}

	// Stop server
	err = server.Process.Kill()
	if err != nil {
		pair.ReturnCode = CROSSRUNNER_FAILURE
		pair.Err = err
	}

}
