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

package integration

import (
	"reflect"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/frugal/lib/go"
)

func newBenchmarkMiddleware() frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			return next(service, method, args)
		}
	}
}

func BenchmarkBinary(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 0, false)
}

func BenchmarkCompact(b *testing.B) {
	benchmarkBasic(b, thrift.NewTCompactProtocolFactory(), 0, false)
}

func BenchmarkJSON(b *testing.B) {
	benchmarkBasic(b, thrift.NewTJSONProtocolFactory(), 0, false)
}

func BenchmarkBinaryServerMiddleware1(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 1, false)
}

func BenchmarkBinaryServerMiddleware5(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 5, false)
}

func BenchmarkBinaryServerMiddleware10(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 10, true)
}

func BenchmarkBinaryClientServerMiddleware1(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 1, true)
}

func BenchmarkBinaryClientServerMiddleware5(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 5, true)
}

func BenchmarkBinaryClientServerMiddleware10(b *testing.B) {
	benchmarkBasic(b, thrift.NewTBinaryProtocolFactoryDefault(), 10, true)
}

func benchmarkBasic(b *testing.B, protoFactory thrift.TProtocolFactory, numMiddleware int, clientMiddleware bool) {
	b.ReportAllocs()

	middleware := make([]frugal.ServiceMiddleware, numMiddleware)
	for i := 0; i < numMiddleware; i++ {
		middleware[i] = newBenchmarkMiddleware()
	}

	// Setup server.
	processor := event.NewFFooProcessor(&BenchmarkHandler{}, middleware...)
	serverTr, err := thrift.NewTServerSocket(defaultAddr)
	if err != nil {
		b.Fatal(err)
	}
	server := frugal.NewFSimpleServerFactory4(
		frugal.NewFProcessorFactory(processor),
		serverTr,
		frugal.NewAdapterTransportFactory(),
		frugal.NewFProtocolFactory(protoFactory),
	)

	// Start server.
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
		if err := server.Serve(); err != nil {
			b.Fatal("Failed to start server:", err.Error())
		}
	}()
	<-ch

	// Setup client.
	transport, err := thrift.NewTSocket(defaultAddr)
	if err != nil {
		b.Fatal(err)
	}
	fTransportFactory := frugal.NewAdapterTransportFactory()
	fTransport := fTransportFactory.GetTransport(transport)
	defer fTransport.Close()
	if err := fTransport.Open(); err != nil {
		b.Fatal(err)
	}
	middleware = []frugal.ServiceMiddleware{}
	if clientMiddleware {
		for i := 0; i < numMiddleware; i++ {
			middleware = append(middleware, newBenchmarkMiddleware())
		}
	}
	client := event.NewFFooClient(fTransport, frugal.NewFProtocolFactory(protoFactory), middleware...)

	runBenchmarkClient(b, client)

	if err := server.Stop(); err != nil {
		b.Fatal("Failed to stop server:", err.Error())
	}
}

func runBenchmarkClient(b *testing.B, client *event.FFooClient) {
	ctx := frugal.NewFContext("")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := client.Ping(ctx); err != nil {
			b.Fatal("Unexpected Ping error:", err)
		}
	}

	b.StopTimer()
}

type BenchmarkHandler struct{}

func (b *BenchmarkHandler) Ping(ctx *frugal.FContext) error {
	return nil
}

func (b *BenchmarkHandler) Blah(ctx *frugal.FContext, num int32, str string, e *event.Event) (int64, error) {
	return 42, nil
}

func (b *BenchmarkHandler) BasePing(ctx *frugal.FContext) error {
	return nil
}

func (b *BenchmarkHandler) OneWay(ctx *frugal.FContext, id event.ID, req event.Request) error {
	return nil
}
