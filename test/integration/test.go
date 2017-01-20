package main

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Workiva/frugal/test/integration/crossrunner"
)

// TODO: Push up a PR to v2_integration branch branch to start getting this
// torn apart. This is going to hurt!!!

// a testCase is a pointer to a valid test pair (client/server) and port to run
// the pair on
type testCase struct {
	pair *crossrunner.Pair
	port uint64
}

// failures is used to store the unepected_failures.log file
// contains a filepath, pointer to the files location, and a mutex for locking
type failures struct {
	path string
	file *os.File
	mu   sync.Mutex
}

func main() {
	startTime := time.Now()

	// path to json test definitions
	filepath := os.Args[1]

	// TODO: Allow setting loglevel to debug with -V flag/-debug/similar
	// log.SetLevel(log.DebugLevel)

	// pairs is a struct of valid client/server pairs loaded from the provided
	// json file
	pairs := crossrunner.Load(filepath)

	// REVIEW: There are a 195 pairs in the cross config, but not all are running, see comment on line 71
	log.Info(len(pairs))
	crossrunnerTasks := make(chan *testCase)

	// All tests run relative to test/integration
	os.Chdir("test/integration")

	// Make log file for unexpected failures
	failLog := &failures{
		path: "log/unexpected_failures.log",
	}
	if file, err := os.Create(failLog.path); err != nil {
		panic(err)
	} else {
		failLog.file = file
	}
	defer failLog.file.Close()

	// Start with arbitrarily high port to avoid collisions, these are still
	// checked before assiging a port to a pair
	var port uint64 = 55000
	var testsRun uint64 = 0
	var failed uint64 = 0

	crossrunner.WriteConsoleHeader()

	// REVIEW: This is exiting before all workers are finished.  Need some sort
	// of waitGroup or something to ensure all are run?
	for workers := 1; workers <= int(runtime.NumCPU()); workers++ {
		go func(crossrunnerTasks <-chan *testCase) {
			for task := range crossrunnerTasks {
				// Run each configuration
				crossrunner.RunConfig(task.pair, task.port)
				// Check return code
				if task.pair.ReturnCode == crossrunner.TEST_FAILURE {
					// if failed, add to the failed count
					atomic.AddUint64(&failed, 1)
					failLog.mu.Lock()
					// copy the logs to the unexpected_failures.log file
					if err := crossrunner.AppendToFailures(failLog.path, task.pair); err != nil {
						panic(err)
					}
					failLog.mu.Unlock()
				} else if task.pair.ReturnCode == crossrunner.CROSSRUNNER_FAILURE {
					// If there was a crossrunner failure, fail immediately
					panic(task.pair.Err)
				}
				// Write configuration results to console
				crossrunner.WritePairResult(task.pair)
				// Increment the count of tests run
				atomic.AddUint64(&testsRun, 1)
			}
		}(crossrunnerTasks)
	}

	// Add each configuration to the crossrunnerTasks channel
	for _, pair := range pairs {
		// Get next available port. Should all be clear, but this is a good sanity check.
		// REVIEW: I am still seeing inconsistent "port already allocated" errors.
		// Must be doing something wrong here, but I'm not sure what.
		available := crossrunner.CheckPort(atomic.LoadUint64(&port))
		tCase := testCase{pair, available}
		// put the test case on the crossrunnerTasks channel
		crossrunnerTasks <- &tCase
		// increment the port
		next := available + 1
		atomic.StoreUint64(&port, next)
	}

	close(crossrunnerTasks)

	// REVIEW: There are also several goroutines still active at this close,
	// presumably the test cases that haven't completed
	log.Info("Currently running goroutines: ", runtime.NumGoroutine())

	// Print out console results
	runningTime := time.Since(startTime) // Do we need to check that this number is positive?
	testCount := atomic.LoadUint64(&testsRun)
	failedCount := atomic.LoadUint64(&failed)
	crossrunner.WriteConsoleFooter(failedCount, testCount, runningTime)

	// If any configurations failed, fail the suite.
	if failedCount > 0 {
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
