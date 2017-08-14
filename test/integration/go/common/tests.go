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
	"bytes"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"github.com/Workiva/frugal/lib/go"
	"github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
)

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

func CallEverything(client *frugaltest.FFrugalTestClient) {
	ctx := frugal.NewFContext("")
	ctx.SetTimeout(5 * time.Second)
	var err error
	if err = client.TestVoid(ctx); err != nil {
		log.Fatal("Unexpected error in TestVoid() call: ", err)
	}

	ctx = frugal.NewFContext("TestString")
	strInput := "thingå∫ç"
	thing, err := client.TestString(ctx, strInput)
	if err != nil {
		log.Fatal("Unexpected error in TestString() call: ", err)
	}
	if thing != strInput {
		log.Fatalf("Unexpected TestString() result, expected 'thing' got '%s' ", thing)
	}

	ctx = frugal.NewFContext("TestBool")
	bl, err := client.TestBool(ctx, true)
	if err != nil {
		log.Fatal("Unexpected error in TestBool() call: ", err)
	}
	if !bl {
		log.Fatalf("Unexpected TestBool() result expected true, got %t ", bl)
	}

	bl, err = client.TestBool(ctx, false)
	if err != nil {
		log.Fatal("Unexpected error in TestBool() call: ", err)
	}
	if bl {
		log.Fatalf("Unexpected TestBool() result expected false, got %t ", bl)
	}

	ctx = frugal.NewFContext("TestByte")
	b, err := client.TestByte(ctx, 42)
	if err != nil {
		log.Fatal("Unexpected error in TestByte() call: ", err)
	}
	if b != 42 {
		log.Fatalf("Unexpected TestByte() result expected 42, got %d ", b)
	}

	ctx = frugal.NewFContext("TestI32")
	i32, err := client.TestI32(ctx, 4242)
	if err != nil {
		log.Fatal("Unexpected error in TestI32() call: ", err)
	}
	if i32 != 4242 {
		log.Fatalf("Unexpected TestI32() result expected 4242, got %d ", i32)
	}

	ctx = frugal.NewFContext("TestI64")
	i64, err := client.TestI64(ctx, 424242)
	if err != nil {
		log.Fatal("Unexpected error in TestI64() call: ", err)
	}
	if i64 != 424242 {
		log.Fatalf("Unexpected TestI64() result expected 424242, got %d ", i64)
	}

	ctx = frugal.NewFContext("TestDouble")
	d, err := client.TestDouble(ctx, 42.42)
	if err != nil {
		log.Fatal("Unexpected error in TestDouble() call: ", err)
	}
	if d != 42.42 {
		log.Fatalf("Unexpected TestDouble() result expected 42.42, got %f ", d)
	}

	// This currently needs to be tested with a number divisible by 4 due to a json serialization issue between go and java
	// https://github.com/Workiva/frugal/issues/412
	// Using 400 for now, will change back to 42 (101010) once the Thrift fix is implemented
	// TODO: Change back to 42
	ctx = frugal.NewFContext("TestBinary")
	binary, err := client.TestBinary(ctx, []byte(strconv.Itoa(400)))
	if err != nil {
		log.Fatal("Unexpected error in TestBinary call: ", err)
	}
	if bytes.Compare(binary, []byte(strconv.Itoa(400))) != 0 {
		log.Fatal("Unexpected TestBinary() result expected 101010, got %b ", binary)
	}

	xs := frugaltest.NewXtruct()
	xs.StringThing = "thing"
	xs.ByteThing = 42
	xs.I32Thing = 4242
	xs.I64Thing = 424242
	ctx = frugal.NewFContext("TestStruct")
	xsret, err := client.TestStruct(ctx, xs)
	if err != nil {
		log.Fatal("Unexpected error in TestStruct() call: ", err)
	}
	if *xs != *xsret {
		log.Fatalf("Unexpected TestStruct() result expected %#v, got %#v ", xs, xsret)
	}

	x2 := frugaltest.NewXtruct2()
	x2.StructThing = xs
	ctx = frugal.NewFContext("TestNest")
	x2ret, err := client.TestNest(ctx, x2)
	if err != nil {
		log.Fatal("Unexpected error in TestNest() call: ", err)
	}
	if !reflect.DeepEqual(x2, x2ret) {
		log.Fatalf("Unexpected TestNest() result expected %#v, got %#v ", x2, x2ret)
	}

	m := map[int32]int32{1: 2, 3: 4, 5: 42}
	ctx = frugal.NewFContext("TestMap")
	mret, err := client.TestMap(ctx, m)
	if err != nil {
		log.Fatal("Unexpected error in TestMap() call: ", err)
	}
	if !reflect.DeepEqual(m, mret) {
		log.Fatalf("Unexpected TestMap() result expected %#v, got %#v ", m, mret)
	}

	sm := map[string]string{"a": "2", "b": "blah", "some": "thing"}
	ctx = frugal.NewFContext("TestStringMap")
	smret, err := client.TestStringMap(ctx, sm)
	if err != nil {
		log.Fatal("Unexpected error in TestStringMap() call: ", err)
	}
	if !reflect.DeepEqual(sm, smret) {
		log.Fatalf("Unexpected TestStringMap() result expected %#v, got %#v ", sm, smret)
	}

	s := map[int32]bool{1: true, 2: true, 42: true}
	ctx = frugal.NewFContext("TestSet")
	sret, err := client.TestSet(ctx, s)
	if err != nil {
		log.Fatal("Unexpected error in TestSet() call: ", err)
	}
	if !reflect.DeepEqual(s, sret) {
		log.Fatalf("Unexpected TestSet() result expected %#v, got %#v ", s, sret)
	}

	l := []int32{1, 2, 42}
	ctx = frugal.NewFContext("TestList")
	lret, err := client.TestList(ctx, l)
	if err != nil {
		log.Fatal("Unexpected error in TestList() call: ", err)
	}
	if !reflect.DeepEqual(l, lret) {
		log.Fatalf("Unexpected TestSet() result expected %#v, got %#v ", l, lret)
	}

	// Note: No enum type check. Go is statically typed
	ctx = frugal.NewFContext("TestEnum")
	eret, err := client.TestEnum(ctx, frugaltest.Numberz_TWO)
	if err != nil {
		log.Fatal("Unexpected error in TestEnum() call: ", err)
	}
	if eret != frugaltest.Numberz_TWO {
		log.Fatalf("Unexpected TestEnum() result expected %#v, got %#v ", frugaltest.Numberz_TWO, eret)
	}

	ctx = frugal.NewFContext("TestTypedef")
	tret, err := client.TestTypedef(ctx, frugaltest.UserId(42))
	if err != nil {
		log.Fatal("Unexpected error in TestTypedef() call: ", err)
	}
	if tret != frugaltest.UserId(42) {
		log.Fatalf("Unexpected TestTypedef() result expected %#v, got %#v ", frugaltest.UserId(42), tret)
	}

	ctx = frugal.NewFContext("TestMapMap")
	mapmap, err := client.TestMapMap(ctx, 42)
	if err != nil {
		log.Fatal("Unexpected error in TestMapMap() call: ", err)
	}
	if !reflect.DeepEqual(mapmap, rmapmap) {
		log.Fatalf("Unexpected TestMapMap() result expected %#v, got %#v ", rmapmap, mapmap)
	}

	ctx = frugal.NewFContext("TestUppercaseMethod")
	upper, err := client.TestUppercaseMethod(ctx, true)
	if err != nil {
		log.Fatal("Unexpected error in TestUppercaseMethod() call: ", err)
	}
	if !upper {
		log.Fatalf("Unexpected TestUppercaseMethod() result expected true, got %t ", upper)
	}

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
	ctx = frugal.NewFContext("TestInsanity")
	insanity, err := client.TestInsanity(ctx, crazy)
	if err != nil {
		log.Fatal("Unexpected error in TestInsanity() call: ", err)
	}
	if !reflect.DeepEqual(crazy, insanity[1][2]) {
		log.Fatalf("Unexpected TestInsanity() first result expected %#v, got %#v ",
			crazy,
			insanity[1][2])
	}
	if !reflect.DeepEqual(crazy, insanity[1][3]) {
		log.Fatalf("Unexpected TestInsanity() second result expected %#v, got %#v ",
			crazy,
			insanity[1][3])
	}

	ctx = frugal.NewFContext("TestMulti")
	xxsret, err := client.TestMulti(ctx, 42, 4242, 424242, map[int16]string{1: "blah", 2: "thing"}, frugaltest.Numberz_EIGHT, frugaltest.UserId(24))
	if err != nil {
		log.Fatal("Unexpected error in TestMulti() call: ", err)
	}
	if !reflect.DeepEqual(xxs, xxsret) {
		log.Fatalf("Unexpected TestMulti() result expected %#v, got %#v ", xxs, xxsret)
	}

	ctx = frugal.NewFContext("TestException")
	err = client.TestException(ctx, "Xception")
	if err == nil {
		log.Fatal("Expecting exception in TestException() call")
	}
	if !reflect.DeepEqual(err, xcept) {
		log.Fatalf("Unexpected TestException() result expected %#v, got %#v ", xcept, err)
	}

	ctx = frugal.NewFContext("TestMultiException")
	ign, err := client.TestMultiException(ctx, "Xception", "ignoreme")
	if ign != nil || err == nil {
		log.Fatal("Expecting exception in TestMultiException() call")
	}
	if !reflect.DeepEqual(err, &frugaltest.Xception{ErrorCode: 1001, Message: "This is an Xception"}) {
		log.Fatalf("Unexpected TestMultiException() %#v ", err)
	}

	ctx = frugal.NewFContext("TestUncaughtException")
	err = client.TestUncaughtException(ctx)
	e, ok := err.(thrift.TApplicationException)
	if !ok || e.TypeId() != frugal.APPLICATION_EXCEPTION_INTERNAL_ERROR || !strings.Contains(e.Error(), "An uncaught error") {
		log.Fatalf("TestUncheckedTApplicationException expected TApplicationException with typeID=%v, got %v.\n Got error=%v", frugal.APPLICATION_EXCEPTION_INTERNAL_ERROR, e.TypeId(), e.Error())
	}

	ctx = frugal.NewFContext("TestUncheckedTApplicationException")
	err = client.TestUncheckedTApplicationException(ctx)
	e, ok = err.(thrift.TApplicationException)
	if !ok || e.TypeId() != 400 || !strings.Contains(e.Error(), "Unchecked TApplicationException") {
		log.Fatalf("TestUncheckedTApplicationException expected TApplicationException with typeID=%v, got %v.\n Got error=%v", 400, e.TypeId(), e.Error())
	}

	ctx = frugal.NewFContext("TestMultiException")
	ign, err = client.TestMultiException(ctx, "Xception2", "ignoreme")
	if ign != nil || err == nil {
		log.Fatal("Expecting exception in TestMultiException() call")
	}
	expecting := &frugaltest.Xception2{ErrorCode: 2002, StructThing: &frugaltest.Xtruct{StringThing: "This is an Xception2"}}
	if !reflect.DeepEqual(err, expecting) {
		log.Fatalf("Unexpected TestMultiException() %#v ", err)
	}

	// Request at the 1mb limit
	request := make([]byte, 1024*1024)
	ctx = frugal.NewFContext("TestRequestTooLarge")
	err = client.TestRequestTooLarge(ctx, request)
	switch e := err.(type) {
	case thrift.TTransportException:
		if e.TypeId() != frugal.TRANSPORT_EXCEPTION_REQUEST_TOO_LARGE {
			log.Fatalf("Unexpected error code %v",
				e.TypeId())
		}
		log.Println("TApplicationException")
	default:
		log.Fatalf("Unexpected TestRequestTooLarge() %v", e.Error())
	}

	request = make([]byte, 4)
	ctx = frugal.NewFContext("TestResponseTooLarge")
	response, err := client.TestResponseTooLarge(ctx, request)
	switch e := err.(type) {
	case thrift.TTransportException:
		if e.TypeId() != frugal.TRANSPORT_EXCEPTION_RESPONSE_TOO_LARGE {
			log.Fatalf("Unexpected error  %v",
				e)
		}
		log.Println("TApplicationException")
	default:
		log.Fatalf("Unexpected TestResponseTooLarge() %v",
			response)
		log.Fatalf("TestResponseTooLarge() error: %v", e.Error())
	}

	ctx = frugal.NewFContext("TestOneway")
	err = client.TestOneway(ctx, 1)
	if err != nil {
		log.Fatal("Unexpected error in TestOneway() call: ", err)
	}
}
