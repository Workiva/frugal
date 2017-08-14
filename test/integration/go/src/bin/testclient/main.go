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

package main

import (
	"flag"
	"log"

	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "stateless", "Transport: stateless, http")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()
	pubSub := make(chan bool)
	sent := make(chan bool)
	clientMiddlewareCalled := make(chan bool, 1)
	client, err := common.StartClient(*host, *port, *transport, *protocol, pubSub, sent, clientMiddlewareCalled)
	if err != nil {
		log.Fatal("Unable to start client: ", err)
	}

	common.CallEverything(client)

	select {
	case <-clientMiddlewareCalled:
	default:
		log.Fatal("Client middleware not invoked")
	}

	close(pubSub)
	<-sent
}

