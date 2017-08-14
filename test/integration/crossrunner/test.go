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
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"os/exec"
)

// a testCase is a pointer to a valid test pair (client/server) and port to run
// the pair on.
type testCase struct {
	pair *Pair
	port int
}

// failures is used to store the unexpected_failures.log file
// contains a filepath, pointer to the files location, count of total failed
// configurations, and a mutex for locking.
type failures struct {
	path   string
	file   *os.File
	failed int
	mu     sync.Mutex
}



func Run(testDefinitions, outDir *string, getCommand func (config Config, port int) (cmd *exec.Cmd, formatted string)) error {
	startTime := time.Now()

	// TODO: Allow setting loglevel to debug with -V flag/-debug/similar
	// log.SetLevel(log.DebugLevel)

	// pairs is a struct of valid client/server pairs loaded from the provided
	// json file
	pairs, err := Load(*testDefinitions)
	if err != nil {
		log.Info("Error in parsing json test definitions")
		panic(err)
	}

	crossrunnerTasks := make(chan *testCase)

	// Need to create log directory for Skynet-cli. This isn't an issue on Skynet.
	if _, err = os.Stat(*outDir); os.IsNotExist(err) {
		if err = os.Mkdir(*outDir, 0755); err != nil {
			log.Infof("Unable to create '%s' directory", *outDir)
			panic(err)
		}
	}
	// Make log file for unexpected failures
	failLog := &failures{
		path: fmt.Sprintf("%s/unexpected_failures.log", *outDir),
	}
	if file, err := os.Create(failLog.path); err != nil {
		log.Info("Unable to create 'unexpected_failures.log'")
		panic(err)
	} else {
		failLog.file = file
	}
	defer failLog.file.Close()

	var (
		testsRun uint64
		wg       sync.WaitGroup
		port     int
	)

	PrintConsoleHeader()

	for workers := 1; workers <= runtime.NumCPU()*2; workers++ {
		go func(crossrunnerTasks <-chan *testCase) {
			for task := range crossrunnerTasks {
				wg.Add(1)
				// Run each configuration
				RunConfig(task.pair, task.port, getCommand)
				errorLog := "\n"
				// Check return code
				if task.pair.ReturnCode != 0 {
					if task.pair.ReturnCode == CrossrunnerFailure {
						// If there was a crossrunner failure, add logs to the client
						errorLog += "***** CROSSRUNNER FAILURE *****\n"
					} else {
						errorLog += "***** TEST FAILURE *****\n"
					}
					// Add error to client logs
					errorLog += fmt.Sprintf("%s\n", task.pair.Err.Error())
					if err := WriteCustomData(task.pair.Client.Logs.Name(), errorLog); err != nil {
						log.Infof("Failed to append crossrunner failure to %s", task.pair.Client.Logs.Name())
						panic(err)
					}
					// if failed, add to the failed count
					failLog.mu.Lock()
					failLog.failed += 1
					// copy the logs to the unexpected_failures.log file
					if err := AppendToFailures(failLog.path, task.pair); err != nil {
						log.Infof("Failed to copy %s and %s to 'unexpected_failures.log'", task.pair.Client.Logs.Name(), task.pair.Server.Logs.Name())
						panic(err)
					}
					failLog.mu.Unlock()
				}
				// Print configuration results to console
				PrintPairResult(task.pair)
				// Increment the count of tests run
				atomic.AddUint64(&testsRun, 1)
				wg.Done()
			}
		}(crossrunnerTasks)
	}

	// TODO: This could run into issues if run outside of Skynet/Skynet-cli
	port = 9000
	// Add each configuration to the crossrunnerTasks channel
	for _, pair := range pairs {
		tCase := testCase{pair, port}
		// put the test case on the crossrunnerTasks channel
		crossrunnerTasks <- &tCase
		port++
	}

	wg.Wait()
	close(crossrunnerTasks)

	// Print out console results
	runningTime := time.Since(startTime)
	testCount := atomic.LoadUint64(&testsRun)
	PrintConsoleFooter(failLog.failed, testCount, runningTime)

	// If any configurations failed, fail the suite.
	if failLog.failed > 0 {
		// If there was a failure, move the logs to correct artifact location
		err := os.Rename(failLog.path, "/testing/artifacts/unexpected_failures.log")
		if err != nil {
			log.Info("Unable to move unexpected_failures.log")
		}
		os.Exit(1)
	} else {
		// If there were no failures, remove the failures file.
		err := os.Remove("log/unexpected_failures.log")
		if err != nil {
			log.Info("Unable to remove empty unexpected_failures.log")
		}
	}
	return nil
}
