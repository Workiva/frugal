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

package common

import (
	"fmt"
	"reflect"

	"github.com/Workiva/frugal/lib/go"
	"github.com/nats-io/go-nats"
)

const (
	preambleHeader = "preamble"
	rambleHeader   = "ramble"
)

func getNatsConn() *nats.Conn {
	addr := nats.DefaultURL
	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{addr}
	natsOptions.Secure = false
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	return conn
}

func clientLoggingMiddleware(called chan<- bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			var (
				isByteArray bool
				length      int
				ret         frugal.Results
			)

			select {
			case called <- true:
			default:
			}

			isByteArray, length = checkForByteArray(args)

			// If first argument after the fContext is a byte array
			// print the length (size) instead of the contents
			if isByteArray {
				fmt.Printf("%v(%v) = ", method.Name, length)
				ret = next(service, method, args)
				byteArray, ok := ret[0].([]byte)
				if ok {
					length = len(byteArray)
				}
				fmt.Printf("%v\n", length)
			} else {
				fmt.Printf("%v(%v) = ", method.Name, args[1:])
				ret = next(service, method, args)
				fmt.Printf("%v\n", ret[:1])
			}
			return ret
		}
	}
}

func serverLoggingMiddleware(called chan<- bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			var (
				isByteArray bool
				length      int
				ret         frugal.Results
			)

			select {
			case called <- true:
			default:
			}

			ret = next(service, method, args)

			isByteArray, length = checkForByteArray(args)

			if isByteArray {
				byteArray, ok := ret[0].([]byte)
				if ok {
					isByteArray = true
					retLength := len(byteArray)
					fmt.Printf("%v(%v) = ", method.Name, length)
					fmt.Printf("%v\n", retLength)
				}
			} else {
				fmt.Printf("%v(%v) = ", method.Name, args[1:])
				fmt.Printf("%v\n", ret[:1])
			}
			return ret
		}
	}
}

func checkForByteArray(args frugal.Arguments) (isByteArray bool, length int) {
	if len(args) > 1 {
		byteArray, ok := args[1].([]byte)
		if ok {
			isByteArray = true
			length = len(byteArray)
			return isByteArray, length
		}
	}
	return false, 0
}
