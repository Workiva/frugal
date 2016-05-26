package main

import (
	"flag"
	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/common"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
	"log"
	"reflect"
	"time"
	"fmt"
)

var host = flag.String("host", "localhost", "Host to connect")
var port = flag.Int64("port", 9090, "Port number to connect")
var domain_socket = flag.String("domain-socket", "", "Domain Socket (e.g. /tmp/frugaltest.frugal), instead of host and port")
var transport = flag.String("transport", "buffered", "Transport: buffered, framed, http, zlib")
var protocol = flag.String("protocol", "binary", "Protocol: binary, compact, json")
var testloops = flag.Int("testloops", 1, "Number of Tests")

func main() {
	flag.Parse()
	client, err := common.StartClient(*host, *port, *domain_socket, *transport, *protocol)
	if err != nil {
		log.Fatal("Unable to start client: ", err)
	}

	for i := 0; i < *testloops; i++ {
		callEverything(client)
	}
}

var rmapmap = map[int32]map[int32]int32{
	-4: {-4: -4, -3: -3, -2: -2, -1: -1},
	4:  {4: 4, 3: 3, 2: 2, 1: 1},
}

var xxs = &frugaltest.Xtruct{
	StringThing: "Hello2",
	ByteThing:   42,
	I32Thing:    4242,
	I64Thing:    424242,
}

var xcept = &frugaltest.Xception{ErrorCode: 1001, Message: "Xception"}

func callEverything(client *frugaltest.FFrugalTestClient) {
	ctx := frugal.NewFContext("")
	ctx.SetTimeout(5 * time.Second)
	var err error
	fmt.Printf("TestVoid()")
	if err = client.TestVoid(ctx); err != nil {
		log.Fatal("Unexpected error in TestVoid() call: ", err)
	}
	fmt.Printf(" = void\n")

	fmt.Printf("TestString(\"thing\")")
	thing, err := client.TestString(ctx, "thing")
	if err != nil {
		log.Fatal("Unexpected error in TestString() call: ", err)
	}
	if thing != "thing" {
		log.Fatal("Unexpected TestString() result, expected 'thing' got '%s' ", thing)
	}
	fmt.Printf(" = %v \n", thing)

	fmt.Printf("TestBool(true)")
	bl, err := client.TestBool(ctx, true)
	if err != nil {
		log.Fatal("Unexpected error in TestBool() call: ", err)
	}
	if !bl {
		log.Fatal("Unexpected TestBool() result expected true, got %f ", bl)
	}
	fmt.Printf(" = %v \n", bl)
	fmt.Printf("TestBool(false)")
	bl, err = client.TestBool(ctx, false)
	if err != nil {
		log.Fatal("Unexpected error in TestBool() call: ", err)
	}
	if bl {
		log.Fatal("Unexpected TestBool() result expected false, got %f ", bl)
	}
	fmt.Printf(" = %v \n", bl)

	fmt.Print("TestByte(42)")
	b, err := client.TestByte(ctx, 42)
	if err != nil {
		log.Fatal("Unexpected error in TestByte() call: ", err)
	}
	if b != 42 {
		log.Fatal("Unexpected TestByte() result expected 42, got %d ", b)
	}
	fmt.Printf(" = %v \n", b)

	fmt.Printf("TestI32(4242)")
	i32, err := client.TestI32(ctx, 4242)
	if err != nil {
		log.Fatal("Unexpected error in TestI32() call: ", err)
	}
	if i32 != 4242 {
		log.Fatal("Unexpected TestI32() result expected 4242, got %d ", i32)
	}
	fmt.Printf(" = %v \n", i32)

	fmt.Print("TestI64(424242)")
	i64, err := client.TestI64(ctx, 424242)
	if err != nil {
		log.Fatal("Unexpected error in TestI64() call: ", err)
	}
	if i64 != 424242 {
		log.Fatal("Unexpected TestI64() result expected 424242, got %d ", i64)
	}
	fmt.Printf(" = %v \n", i64)

	fmt.Printf("TestDouble(42.42)")
	d, err := client.TestDouble(ctx, 42.42)
	if err != nil {
		log.Fatal("Unexpected error in TestDouble() call: ", err)
	}
	if d != 42.42 {
		log.Fatal("Unexpected TestDouble() result expected 42.42, got %f ", d)
	}
	fmt.Printf(" = %v \n", d)

	// TODO: add TestBinary() call

	xs := frugaltest.NewXtruct()
	xs.StringThing = "thing"
	xs.ByteThing = 42
	xs.I32Thing = 4242
	xs.I64Thing = 424242
	fmt.Printf("TestStruct(%v)", xs)
	xsret, err := client.TestStruct(ctx, xs)
	if err != nil {
		log.Fatal("Unexpected error in TestStruct() call: ", err)
	}
	if *xs != *xsret {
		log.Fatal("Unexpected TestStruct() result expected %#v, got %#v ", xs, xsret)
	}
	fmt.Printf(" = %v \n", xsret)

	x2 := frugaltest.NewXtruct2()
	x2.StructThing = xs
	fmt.Printf("TestNest(%v)", x2)
	x2ret, err := client.TestNest(ctx, x2)
	if err != nil {
		log.Fatal("Unexpected error in TestNest() call: ", err)
	}
	if !reflect.DeepEqual(x2, x2ret) {
		log.Fatal("Unexpected TestNest() result expected %#v, got %#v ", x2, x2ret)
	}
	fmt.Printf(" = %v \n", x2ret)

	m := map[int32]int32{1: 2, 3: 4, 5: 42}
	fmt.Printf("TestMap(%v)", m)
	mret, err := client.TestMap(ctx, m)
	if err != nil {
		log.Fatal("Unexpected error in TestMap() call: ", err)
	}
	if !reflect.DeepEqual(m, mret) {
		log.Fatal("Unexpected TestMap() result expected %#v, got %#v ", m, mret)
	}
	fmt.Printf(" = %v \n", bl)

	sm := map[string]string{"a": "2", "b": "blah", "some": "thing"}
	fmt.Printf("TestStringMap(%v)", sm)
	smret, err := client.TestStringMap(ctx, sm)
	if err != nil {
		log.Fatal("Unexpected error in TestStringMap() call: ", err)
	}
	if !reflect.DeepEqual(sm, smret) {
		log.Fatal("Unexpected TestStringMap() result expected %#v, got %#v ", sm, smret)
	}
	fmt.Printf(" = %v \n", smret)

	s := map[int32]bool{1: true, 2: true, 42: true}
	fmt.Printf("TestSet(%v)", s)
	sret, err := client.TestSet(ctx, s)
	if err != nil {
		log.Fatal("Unexpected error in TestSet() call: ", err)
	}
	if !reflect.DeepEqual(s, sret) {
		log.Fatal("Unexpected TestSet() result expected %#v, got %#v ", s, sret)
	}
	fmt.Printf(" = %v \n", sret)

	l := []int32{1, 2, 42}
	fmt.Printf("TestList(%v)", l)
	lret, err := client.TestList(ctx, l)
	if err != nil {
		log.Fatal("Unexpected error in TestList() call: ", err)
	}
	if !reflect.DeepEqual(l, lret) {
		log.Fatal("Unexpected TestSet() result expected %#v, got %#v ", l, lret)
	}
	fmt.Printf(" = %v \n", lret)

	fmt.Printf("TestEnum(%v)", frugaltest.Numberz_TWO)
	eret, err := client.TestEnum(ctx, frugaltest.Numberz_TWO)
	if err != nil {
		log.Fatal("Unexpected error in TestEnum() call: ", err)
	}
	if eret != frugaltest.Numberz_TWO {
		log.Fatal("Unexpected TestEnum() result expected %#v, got %#v ", frugaltest.Numberz_TWO, eret)
	}
	fmt.Printf(" = %v \n", eret)

	fmt.Printf("TestTypedef(%v)", frugaltest.UserId(42))
	tret, err := client.TestTypedef(ctx, frugaltest.UserId(42))
	if err != nil {
		log.Fatal("Unexpected error in TestTypedef() call: ", err)
	}
	if tret != frugaltest.UserId(42) {
		log.Fatal("Unexpected TestTypedef() result expected %#v, got %#v ", frugaltest.UserId(42), tret)
	}
	fmt.Printf(" = %v \n", tret)

	fmt.Printf("TestMapMap(42)")
	mapmap, err := client.TestMapMap(ctx, 42)
	if err != nil {
		log.Fatal("Unexpected error in TestMapMap() call: ", err)
	}
	if !reflect.DeepEqual(mapmap, rmapmap) {
		log.Fatal("Unexpected TestMapMap() result expected %#v, got %#v ", rmapmap, mapmap)
	}
	fmt.Printf(" = %v \n", mapmap)

	crazy := frugaltest.NewInsanity()
	crazy.UserMap = map[frugaltest.Numberz]frugaltest.UserId{
		frugaltest.Numberz_FIVE:  5,
		frugaltest.Numberz_EIGHT: 8,
	}
	truck1 := frugaltest.NewXtruct()
	truck1.StringThing = "Goodbye4"
	truck1.ByteThing = 4
	truck1.I32Thing = 4
	truck1.I64Thing = 4
	truck2 := frugaltest.NewXtruct()
	truck2.StringThing = "Hello2"
	truck2.ByteThing = 2
	truck2.I32Thing = 2
	truck2.I64Thing = 2
	crazy.Xtructs = []*frugaltest.Xtruct{
		truck1,
		truck2,
	}
	fmt.Printf("TestInsanity(%v)", crazy)
	insanity, err := client.TestInsanity(ctx, crazy)
	if err != nil {
		log.Fatal("Unexpected error in TestInsanity() call: ", err)
	}
	if !reflect.DeepEqual(crazy, insanity[1][2]) {
		log.Fatal("Unexpected TestInsanity() first result expected %#v, got %#v ",
			crazy,
			insanity[1][2])
	}
	if !reflect.DeepEqual(crazy, insanity[1][3]) {
		log.Fatal("Unexpected TestInsanity() second result expected %#v, got %#v ",
			crazy,
			insanity[1][3])
	}
	if len(insanity[2][6].UserMap) > 0 || len(insanity[2][6].Xtructs) > 0 {
		log.Fatal("Unexpected TestInsanity() non-empty result got %#v ",
			insanity[2][6])
	}
	fmt.Printf(" = %v \n", insanity)

	fmt.Printf("TestMulti()")
	xxsret, err := client.TestMulti(ctx, 42, 4242, 424242, map[int16]string{1: "blah", 2: "thing"}, frugaltest.Numberz_EIGHT, frugaltest.UserId(24))
	if err != nil {
		log.Fatal("Unexpected error in TestMulti() call: ", err)
	}
	if !reflect.DeepEqual(xxs, xxsret) {
		log.Fatal("Unexpected TestMulti() result expected %#v, got %#v ", xxs, xxsret)
	}
	fmt.Printf(" = %v \n", xxsret)

	fmt.Printf("TestException(\"Xception\")")
	err = client.TestException(ctx, "Xception")
	if err == nil {
		log.Fatal("Expecting exception in TestException() call")
	}
	if !reflect.DeepEqual(err, xcept) {
		log.Fatal("Unexpected TestException() result expected %#v, got %#v ", xcept, err)
	}
	fmt.Printf(" = %v \n", err)

	// TODO: Need to handle the test case where an untyped exception is thrown. Handle reopening the transport after frugal freaks out.
	// err = client.TestException(ctx, "TException") // This is closing the transport
	// _, ok := err.(thrift.TApplicationException)
	// if err == nil || !ok {
	// 	log.Fatal("Unexpected TestException() result expected ApplicationError, got %#v ", err)
	// }

	fmt.Printf("TestMultiException(\"Xception\", \"ignoreme\")")
	ign, err := client.TestMultiException(ctx, "Xception", "ignoreme")
	if ign != nil || err == nil {
		log.Fatal("Expecting exception in TestMultiException() call")
	}
	if !reflect.DeepEqual(err, &frugaltest.Xception{ErrorCode: 1001, Message: "This is an Xception"}) {
		log.Fatal("Unexpected TestMultiException() %#v ", err)
	}
	fmt.Printf(" = %v \n", &frugaltest.Xception{ErrorCode: 1001, Message: "This is an Xception"})

	fmt.Printf("TestMultiException(\"Xception2\", \"ignoreme\")")
	ign, err = client.TestMultiException(ctx, "Xception2", "ignoreme")
	if ign != nil || err == nil {
		log.Fatal("Expecting exception in TestMultiException() call")
	}
	expecting := &frugaltest.Xception2{ErrorCode: 2002, StructThing: &frugaltest.Xtruct{StringThing: "This is an Xception2"}}
	if !reflect.DeepEqual(err, expecting) {
		log.Fatal("Unexpected TestMultiException() %#v ", err)
	}
	fmt.Printf(" = %v \n", expecting)

	fmt.Printf("TestOneWay()")
	err = client.TestOneway(ctx, 2)
	if err != nil {
		log.Fatal("Unexpected error in TestOneway() call: ", err)
	}
	fmt.Printf(" - no error returned \n")

	//Make sure the connection still alive
	fmt.Printf("TestVoid()")
	if err = client.TestVoid(ctx); err != nil {
		log.Fatal("Unexpected error in TestVoid() call: ", err)
	}
	fmt.Printf(" = void\n")
}
