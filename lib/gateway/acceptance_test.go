package gateway

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	"github.com/Workiva/frugal/compiler"
)

const (
	defaultGatewayHost = "http://localhost:8000"
	defaultIDLFile     = "test_idl.frugal"
	outputDir          = "gen-go"
)

type harness struct {
	idlFile   string
	outputDir string
}

func (h *harness) check(e error) {
	if e != nil {
		h.cleanUp()
		panic(e)
	}
}

func matchJSON(expected string, actual string) bool {
	return true
}

func (h *harness) cleanUp() {
	// os.RemoveAll(h.outputDir)
	// os.Remove(h.idlFile)

	// TODO: close goroutines?

	// if h.frugalProcess != nil {
	// 	h.frugalProcess.Signal(os.Interrupt)
	// }

	// if h.gatewayProcess != nil {
	// 	fmt.Println("closing gateway")
	// 	h.frugalProcess.Signal(os.Interrupt)
	// }

}

func (h *harness) addToIDL(idl *gherkin.DocString) error {
	f, err := os.OpenFile(h.idlFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	h.check(err)

	content := fmt.Sprintf("\n\n%s\n\n", idl.Content) // Add whitespace between chunks
	_, err = f.WriteString(content)
	h.check(err)

	err = f.Close()
	h.check(err)

	return nil
}

func (h *harness) compile(lang string) error {
	options := compiler.Options{
		File:    h.idlFile,
		Gen:     lang,
		Out:     h.outputDir,
		Delim:   ".",
		Recurse: false,
		Verbose: false,
	}

	err := compiler.Compile(options)
	h.check(err)

	return nil
}

func (h *harness) compileGo() error {
	return h.compile("go")
}

func (h *harness) compileGateway() error {
	return h.compile("gateway")
}

func (h *harness) runServer(command string) error {
	cmd := exec.Command("go", "run", command)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}

func (h *harness) aFrugalServerIsRunning() error {
	return h.runServer("acceptanceHttp/main.go")
}

func (h *harness) aGatewayIsRunning() error {
	return h.runServer("acceptanceGateway/main.go")
}

func (h *harness) shouldReturn(method string, endpoint string, expected *gherkin.DocString) error {
	endpoint = defaultGatewayHost + endpoint
	switch strings.ToLower(method) {
	case "get":
		resp, err := http.Get(endpoint)
		fmt.Println(resp)
		if err != nil {
			return err
		}
		// if !matchJSON(expected.Content, resp) {
		// 	return errors.New("no match")
		// }
		// case "post":
		// 	_, err := http.Post(endpoint)
		// 	return err
	}
	// case "
	// resp, err := http.Get()
	// fmt.Println(resp)
	return nil
}

func FeatureContext(s *godog.Suite) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	h := &harness{
		idlFile:   filepath.Join(filepath.Dir(ex), defaultIDLFile),
		outputDir: outputDir,
	}

	// Start from a clean slate
	s.BeforeSuite(func() {
		h.cleanUp()
	})

	// Clean the IDL and any generated code after running each scenario
	s.AfterScenario(func(interface{}, error) {
		h.cleanUp()
	})

	// Scenario: Exposing an endpoint
	// Given
	s.Step(`^valid IDL with body$`, h.addToIDL)
	s.Step(`^a service method with annotations like$`, h.addToIDL)
	// When
	s.Step(`^the compiler generates a Frugal processor$`, h.compileGo)
	s.Step(`^a Frugal server is running$`, h.aFrugalServerIsRunning)
	s.Step(`^the compiler generates HTTP proxy handlers$`, h.compileGateway)
	s.Step(`^a proxy server is running$`, h.aGatewayIsRunning)

	s.Step(`^the method "([^"]*)" on endpoint "([^"]*)" should return with$`, h.shouldReturn)

	// Close the file
	// if err := f.Close(); err != nil {
	// 	panic(err)
	// }
}
