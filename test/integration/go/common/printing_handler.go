package common

import (
	"errors"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"github.com/Workiva/frugal/lib/go"
	. "github.com/Workiva/frugal/test/integration/go/gen/frugaltest"
	"log"
)

var PrintingHandler = &printingHandler{}

type printingHandler struct{}

// TestVoid returns nothing
func (p *printingHandler) TestVoid(ctx frugal.FContext) (err error) {
	return nil
}

// TestString returns the string it was called with
func (p *printingHandler) TestString(ctx frugal.FContext, thing string) (r string, err error) {
	return thing, nil
}

// TestBool returns the bool argument it was called with
func (p *printingHandler) TestBool(ctx frugal.FContext, thing bool) (r bool, err error) {
	return thing, nil
}

// TestByte returns the int8 argument it was called with
func (p *printingHandler) TestByte(ctx frugal.FContext, thing int8) (r int8, err error) {
	return thing, nil
}

// TestI32 returns the int32 argument it was called with
func (p *printingHandler) TestI32(ctx frugal.FContext, thing int32) (r int32, err error) {
	return thing, nil
}

// TestI64 returns the int64 it was called with
func (p *printingHandler) TestI64(ctx frugal.FContext, thing int64) (r int64, err error) {
	return thing, nil
}

// TestDouble returns the double it was called with
func (p *printingHandler) TestDouble(ctx frugal.FContext, thing float64) (r float64, err error) {
	return thing, nil
}

// TestBinary returns the byte array it was called with
func (p *printingHandler) TestBinary(ctx frugal.FContext, thing []byte) (r []byte, err error) {
	return thing, nil
}

// TestStruct returns the Xtruct it was called with
func (p *printingHandler) TestStruct(ctx frugal.FContext, thing *Xtruct) (r *Xtruct, err error) {
	return thing, err
}

// TestNest returns the nested Xtruct it was called with
func (p *printingHandler) TestNest(ctx frugal.FContext, nest *Xtruct2) (r *Xtruct2, err error) {
	return nest, nil
}

// TestMap returns the map of int32s it was called with
func (p *printingHandler) TestMap(ctx frugal.FContext, thing map[int32]int32) (r map[int32]int32, err error) {
	return thing, nil
}

// TestStringMap returns the map of strings it was called with
func (p *printingHandler) TestStringMap(ctx frugal.FContext, thing map[string]string) (r map[string]string, err error) {
	return thing, nil
}

// TestSet returns the map of bools it was called with
func (p *printingHandler) TestSet(ctx frugal.FContext, thing map[int32]bool) (r map[int32]bool, err error) {
	return thing, nil
}

// TestList returns the int32 list it was called with
func (p *printingHandler) TestList(ctx frugal.FContext, thing []int32) (r []int32, err error) {
	return thing, nil
}

// TestEnum returns the enum it was called with
func (p *printingHandler) TestEnum(ctx frugal.FContext, thing Numberz) (r Numberz, err error) {
	return thing, nil
}

// TestTypedef returns the UserID it was called with
func (p *printingHandler) TestTypedef(ctx frugal.FContext, thing UserId) (r UserId, err error) {
	return thing, nil
}

// TestMapMap takes an int32 and returns a dictionary with these values:
// {-4 => {-4 => -4, -3 => -3, -2 => -2, -1 => -1, }, 4 => {1 => 1, 2 => 2, 3 => 3, 4 => 4, }, }
func (p *printingHandler) TestMapMap(ctx frugal.FContext, hello int32) (r map[int32]map[int32]int32, err error) {
	r = map[int32]map[int32]int32{
		-4: {-4: -4, -3: -3, -2: -2, -1: -1},
		4:  {4: 4, 3: 3, 2: 2, 1: 1},
	}
	return
}

// TestBool returns the bool argument it was called with
func (p *printingHandler) TestUppercaseMethod(ctx frugal.FContext, thing bool) (r bool, err error) {
	return thing, nil
}

// TestInsanity takes an insanity argument and returns it in a map:
//
//   { 1 => { 2 => argument,
//            3 => argument,
//          },
//     2 => { },
//   }
func (p *printingHandler) TestInsanity(ctx frugal.FContext, argument *Insanity) (r map[UserId]map[Numberz]*Insanity, err error) {
	r = make(map[UserId]map[Numberz]*Insanity)
	r[1] = map[Numberz]*Insanity{
		2: argument,
		3: argument,
	}
	r[2] = map[Numberz]*Insanity{}
	return
}

// TestMulti takes several different types of arguments:
// @param byte arg0 -
// @param i32 arg1 -
// @param i64 arg2 -
// @param map<i16, string> arg3 -
// @param Numberz arg4 -
// @param UserId arg5 -
// @return Xtruct - returns an Xtruct with StringThing = "Hello2, ByteThing = arg0, I32Thing = arg1
//  and I64Thing = arg2
func (p *printingHandler) TestMulti(ctx frugal.FContext, arg0 int8, arg1 int32, arg2 int64, arg3 map[int16]string, arg4 Numberz, arg5 UserId) (r *Xtruct, err error) {
	r = NewXtruct()

	r.StringThing = "Hello2"
	r.ByteThing = arg0
	r.I32Thing = arg1
	r.I64Thing = arg2
	return
}

// TestException
// @param string arg - a string indication what type of exception to throw
// if arg == "Xception" throw Xception with errorCode = 1001 and message = arg
// else if arg == "TException" throw TException
// else do not throw anything
//
// Parameters:
//  - Arg
func (p *printingHandler) TestException(ctx frugal.FContext, arg string) (err error) {
	switch arg {
	case "Xception":
		e := NewXception()
		e.ErrorCode = 1001
		e.Message = arg
		return e
	case "TException":
		return errors.New("Just TException")
	}
	return
}

// TestUncaughtException
// Raises an unexpected non-defined, non-TApplication exception in the processor handler.
func (p *printingHandler) TestUncaughtException(ctx frugal.FContext) (err error) {
	return errors.New("An uncaught error")
}


// TestUncheckedTApplicationException
// Raises an unexpected non-defined, non-TApplication exception in the processor handler.
func (p *printingHandler) TestUncheckedTApplicationException(ctx frugal.FContext) (err error) {
	return thrift.NewTApplicationException(400, "Unchecked TApplicationException")

}

// TestRequestTooLarge
// Only used for testing with NATS.
// This case should never be hit because the client should encounter a
// message size error.
func(p *printingHandler) TestRequestTooLarge(ctx frugal.FContext, request []byte) (err error) {
	log.Fatal("TestRequestTooLarge should never be successfully called.")
	return
}

// TestRequestAlmostTooLarge
// Only used for testing with NATS.
// This case should never be hit because the client should encounter a
// message size error.
func(p *printingHandler) TestRequestAlmostTooLarge(ctx frugal.FContext, request []byte) (err error) {
	log.Fatal("TestRequestAlmostTooLarge should never be succeessfully called.")
	return
}

// TestResponseTooLarge
// Only used for testing with NATS.
// Takes a []btye that is under the 1mb limit and returns with a message that is
// over the 1mb limit.
func(p *printingHandler) TestResponseTooLarge (ctx frugal.FContext, request []byte) (response []byte, err error) {
	response = make([]byte, 1024*1024)
	return response, nil
}

// TestMultiException
// @param string arg - a string indication what type of exception to throw
// if arg0 == "Xception" throw Xception with errorCode = 1001 and message = "This is an Xception"
// else if arg0 == "Xception2" throw Xception2 with errorCode = 2002 and message = "This is an Xception2"
// else do not throw anything
// @return Xtruct - an Xtruct with StringThing = arg1
//
// Parameters:
//  - Arg0
//  - Arg1
func (p *printingHandler) TestMultiException(ctx frugal.FContext, arg0 string, arg1 string) (r *Xtruct, err error) {
	switch arg0 {

	case "Xception":
		e := NewXception()
		e.ErrorCode = 1001
		e.Message = "This is an Xception"
		return nil, e
	case "Xception2":
		e := NewXception2()
		e.ErrorCode = 2002
		e.StructThing = NewXtruct()
		e.StructThing.StringThing = "This is an Xception2"
		return nil, e
	default:
		r = NewXtruct()
		r.StringThing = arg1
		return
	}
}

// TestOneway takes an int32 and returns nothing
func (p *printingHandler) TestOneway(ctx frugal.FContext, msToSleep int32) (err error) {
	time.Sleep(time.Millisecond * time.Duration(msToSleep))
	return
}
