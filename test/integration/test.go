package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Workiva/frugal/test/integration/crossrunner"
)

// a testCase is a pointer to a valid test pair (client/server) and port to run
// the pair on.
type testCase struct {
	pair *crossrunner.Pair
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

func main() {
	startTime := time.Now()

	// path to json test definitions
	var testDefinitions string
	if len(os.Args) < 2 {
		log.Fatal("Expected test definition json file. None provided.")
	} else {
		testDefinitions = os.Args[1]
	}

	// TODO: Allow setting loglevel to debug with -V flag/-debug/similar
	// log.SetLevel(log.DebugLevel)

	// pairs is a struct of valid client/server pairs loaded from the provided
	// json file
	pairs, err := crossrunner.Load(testDefinitions)
	if err != nil {
		log.Info("Error in parsing json test definitions")
		panic(err)
	}

	crossrunnerTasks := make(chan *testCase)

	// All tests run relative to test/integration
	if err := os.Chdir("test/integration"); err != nil {
		log.Info("Unable to change directory to /test/integration")
		panic(err)
	}

	// Need to create log directory for Skynet-cli. This isn't an issue on Skynet.
	if _, err = os.Stat("log"); os.IsNotExist(err) {
		if err = os.Mkdir("log", 0755); err != nil {
			log.Info("Unable to create 'log' directory")
			panic(err)
		}
	}
	// Make log file for unexpected failures
	failLog := &failures{
		path: "log/unexpected_failures.log",
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

	crossrunner.PrintConsoleHeader()

	for workers := 1; workers <= runtime.NumCPU()*2; workers++ {
		go func(crossrunnerTasks <-chan *testCase) {
			for task := range crossrunnerTasks {
				wg.Add(1)
				// Run each configuration
				crossrunner.RunConfig(task.pair, task.port)
				errorLog := "\n"
				// Check return code
				if task.pair.ReturnCode != 0 {
					if task.pair.ReturnCode == crossrunner.CrossrunnerFailure {
						// If there was a crossrunner failure, add logs to the client
						errorLog += "***** CROSSRUNNER FAILURE *****\n"
					} else {
						errorLog += "***** TEST FAILURE *****\n"
					}
					// Add error to client logs
					errorLog += fmt.Sprintf("%s\n", task.pair.Err.Error())
					if err := crossrunner.WriteCustomData(task.pair.Client.Logs.Name(), errorLog); err != nil {
						log.Infof("Failed to append crossrunner failure to %s", task.pair.Client.Logs.Name())
						panic(err)
					}
					// if failed, add to the failed count
					failLog.mu.Lock()
					failLog.failed += 1
					// copy the logs to the unexpected_failures.log file
					if err := crossrunner.AppendToFailures(failLog.path, task.pair); err != nil {
						log.Infof("Failed to copy %s and %s to 'unexpected_failures.log'", task.pair.Client.Logs.Name(), task.pair.Server.Logs.Name())
						panic(err)
					}
					failLog.mu.Unlock()
				}
				// Print configuration results to console
				crossrunner.PrintPairResult(task.pair)
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
	crossrunner.PrintConsoleFooter(failLog.failed, testCount, runningTime)

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
}
