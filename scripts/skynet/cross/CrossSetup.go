package main

import (
	"os/exec"
	"sync"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

func main(){

	setupScriptDir := "scripts/skynet/cross/"
	setupScripts := []string{"setup_dart.sh", "setup_go.sh", "setup_tornado.sh", "setup_java.sh"}

	// Allow each setup script to run concurrently
	var wg sync.WaitGroup
	wg.Add(len(setupScripts))

	for _, script := range setupScripts {
		go runSetupScript(script, setupScriptDir, &wg)
	}

	wg.Wait()

}

func runSetupScript(script string, scriptDir string, wg *sync.WaitGroup){
	fullScript := scriptDir + script
	log.Info("Running setup script:", script)
	out, err := exec.Command("sh", fullScript).CombinedOutput();

	if err != nil {
		log.Errorf("Script '%s' failed with output:%s", script, out)
	}

	logFile := "/testing/artifacts/" + script + "_out.txt"
	err2 := writeFile(logFile, out)
	if err2 != nil {
		log.Errorf("Writing log file '%s' failed with error:%s", logFile, err2)
	}

	if err != nil || err2 != nil {
		os.Exit(1)
	}

	log.Info("Setup script complete:", script)
	wg.Done()

}

func writeFile(logFile string, commandData []byte) (error) {

	return ioutil.WriteFile(logFile, commandData, 0644)

}