package crossrunner

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Client/Server struct used in tests.json
type options struct {
	Command    []string
	Transports []string
	Protocols  []string
	Timeout    time.Duration
}

// Language struct used in tests.json
type Languages struct {
	Name       string
	Client     options
	Server     options
	Transports []string
	Protocols  []string
	Command    []string
	Workdir    string
}

//  Complete information required to shell out a client or server command
type config struct {
	Name      string
	Timeout   time.Duration
	Transport string
	Protocol  string
	Command   []string
	Workdir   string
	Logs      *LogFile
}

// Matched client and server commands
type Pair struct {
	Client     config
	Server     config
	ReturnCode int
	Err        error
}

func newPair(client, server config) *Pair {
	return &Pair{
		Client: client,
		Server: server,
	}
}

// Load takes a json file of client/server definitions and returns a list of
// valid client/server pairs
func Load(jsonFile string) (pairs []*Pair) {
	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		panic(err)
	}

	Tests := new([]Languages)

	// Unmarshal json into defined structs
	if err := json.Unmarshal(bytes, &Tests); err != nil {
		panic(err)
	}

	// Create empty lists of client and server configurations
	clients := []config{}
	servers := []config{}

	// Iterate over each language to get all client/server configurations
	for _, test := range *Tests {

		// Append "global" transports and protocols to client/server level
		test.Client.Transports = append(test.Client.Transports, test.Transports...)
		test.Server.Transports = append(test.Server.Transports, test.Transports...)
		test.Client.Protocols = append(test.Client.Protocols, test.Protocols...)
		test.Server.Protocols = append(test.Server.Protocols, test.Protocols...)

		// Get expanded list of clients/servers
		clients = append(clients, getList(test.Client, test)...)
		servers = append(servers, getList(test.Server, test)...)
	}

	// Find all valid client/server pairs
	// TODO: Accept some sort of flag(s) that would limit this list of pairs by
	// desired language(s) or other restrictions
	for _, client := range clients {
		for _, server := range servers {
			if server.Transport == client.Transport && server.Protocol == client.Protocol {
				pairs = append(pairs, newPair(client, server))
			}
		}
	}

	return pairs
}
