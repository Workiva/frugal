package main

import (
	"os/exec"
	"sync"
	"io/ioutil"
	"os"
	log "github.com/Sirupsen/logrus"
)

func main(){
	// By setting ForceColors = true, this tricks logrus into thinking it is writing
	// to a terminal output - this allows linebreaks and colors to be displayed on the
	// Smithy webview
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	testScriptDir := "scripts/smithy/"
	// For new/removed files, update smithy.yaml to no longer print
	testScripts := []string{"smithy_dart.sh", "smithy_go.sh", "smithy_java.sh", "smithy_generator.sh", "smithy_python.sh"}

	// Allow each setup script to run concurrently
	var wg sync.WaitGroup
	wg.Add(len(testScripts))

	for _, script := range testScripts {
		go runTestScript(script, testScriptDir, &wg)
	}

	wg.Wait()

}

func runTestScript(script string, scriptDir string, wg *sync.WaitGroup){
	fullScript := scriptDir + script
	log.Info("Running script:", script)
	out, err := exec.Command("/bin/bash", fullScript).CombinedOutput();

	if err != nil {
		log.Errorf("Script '%s' failed with output:\n%s", script, out)
	}


	logFile := os.ExpandEnv("${SMITHY_ROOT}/test_results/" + script + "_out.txt")
	err2 := writeFile(logFile, out)

	if err2 != nil {
		log.Errorf("Writing log file '%s' failed with error:%s", logFile, err2)
	}

	if err != nil || err2 != nil {
		os.Exit(1)
	}

	log.Info("Test script complete:", script)
	wg.Done()

}

func writeFile(logFile string, commandData []byte) (error) {

	return ioutil.WriteFile(logFile, commandData, 0644)

}
