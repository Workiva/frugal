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

	log "github.com/Sirupsen/logrus"

	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/common"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var transport = flag.String("transport", "http", "Transport: stateless, http")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")

func main() {
	flag.Parse()
	clientMiddlewareCalled := make(chan bool, 1)
	client, err := common.StartHTTPClient(*host, *port, *transport, *protocol, clientMiddlewareCalled)
	if err != nil {
		log.Fatal("Unable to start client: ", err)
	}
	log.Infoln("GOOOO SIDEKICK GO!!")
	ctx := frugal.NewFContext("Go Sidekick TestString")
	strInput := "Go run toward the light!"
	thing, err := client.TestString(ctx, strInput)
	if err != nil {
		log.Fatal("Basic Unexpected error in TestString() call: ", err)
	}
	if thing != strInput {
		log.Fatalf("Basic Unexpected TestString() result, expected 'thing' got '%s' ", thing)
	}

	ctx = frugal.NewFContext("Go Sidekick TestString 2")
	strInput = "Run away! Run Away!"
	thing, err = client.TestString(ctx, strInput)
	if err != nil {
		log.Fatal("Basic Unexpected error in TestString() 2 call: ", err)
	}
	if thing != strInput {
		log.Fatalf("Basic Unexpected TestString() 2 result, expected 'thing' got '%s' ", thing)
	}
}
