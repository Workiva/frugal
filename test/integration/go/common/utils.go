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
			select {
			case called <- true:
			default:
			}
			fmt.Printf("%v(%v) = ", method.Name, args[1:])
			ret := next(service, method, args)
			fmt.Printf("%v\n", ret[:1])
			return ret
		}
	}
}

func serverLoggingMiddleware(called chan<- bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			select {
			case called <- true:
			default:
			}
			fmt.Printf("%v(%v) \n", method.Name, args[1:])
			ret := next(service, method, args)
			return ret
		}
	}
}
