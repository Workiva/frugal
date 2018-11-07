// Autogenerated by Frugal Compiler (2.23.0)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package variety

import (
	"errors"

	"github.com/Workiva/frugal/lib/gopherjs/frugal"
	"github.com/Workiva/frugal/test/expected/gopherjs/base"
)

// RedefConst is a constant.
const RedefConst = base.ConstI32FromBase

var ConstThing *base.Thing

var DEFAULT_ID ID

var OtherDefault ID

// Thirtyfour is a constant.
const Thirtyfour = 34

var MAPCONSTANT map[string]string

var ConstEvent1 *Event

var ConstEvent2 *Event

var NumsList []int32

var NumsSet map[Int]bool

var MAPCONSTANT2 map[string]*Event

var BinConst []byte

// TrueConstant is a constant.
const TrueConstant = true

// FalseConstant is a constant.
const FalseConstant = false

// ConstHc is a constant.
const ConstHc = 2

// EvilString is a constant.
const EvilString = "thin'g\" \""

// EvilString2 is a constant.
const EvilString2 = "th'ing\"ad\"f"

var ConstLower *TestLowercase

func init() {
	ConstThing = &Thing{
		AnID:    1,
		AString: "some string",
	}
	DEFAULT_ID = -1
	OtherDefault = DEFAULT_ID
	MAPCONSTANT = map[string]string{
		"hello":     "world",
		"goodnight": "moon",
	}
	ConstEvent1 = &Event{
		ID:      -2,
		Message: "first one",
	}
	ConstEvent2 = &Event{
		ID:      -7,
		Message: "second one",
	}
	NumsList = []int32{
		2,
		4,
		7,
		1,
	}
	NumsSet = map[Int]bool{
		1: true,
		3: true,
		8: true,
		0: true,
	}
	MAPCONSTANT2 = map[string]*Event{
		"hello": &Event{
			ID:      -2,
			Message: "first here",
		},
	}
	BinConst = []byte("hello")
	ConstLower = &TestLowercase{
		LowercaseInt: 2,
	}
}

// ID is a typeDef
type ID int64

// Int is a typeDef
type Int int32

// Request is a typeDef
type Request map[Int]string

// T1String is a typeDef
type T1String string

// T2String is a typeDef
type T2String T1String

// HealthCondition is an enum.
type HealthCondition int64

// HealthCondition values.
const (
	HealthConditionPASS    HealthCondition = 1
	HealthConditionWARN    HealthCondition = 2
	HealthConditionFAIL    HealthCondition = 3
	HealthConditionUNKNOWN HealthCondition = 4
)

// ItsAnEnum is an enum.
type ItsAnEnum int64

// ItsAnEnum values.
const (
	ItsAnEnumFIRST  ItsAnEnum = 2
	ItsAnEnumSECOND ItsAnEnum = 3
	ItsAnEnumTHIRD  ItsAnEnum = 4
	ItsAnEnumfourth ItsAnEnum = 5
	ItsAnEnumFifth  ItsAnEnum = 6
	ItsAnEnumsIxItH ItsAnEnum = 7
)

// TestBase is a frual serializable object.
type TestBase struct {
	BaseStruct *base.Thing
}

// NewTestBase constructs a TestBase.
func NewTestBase() *TestBase {
	return &TestBase{
		// TODO: default values

	}
}

// Unpack deserializes TestBase objects.
func (p *TestBase) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("TestBase")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.BaseStruct = base.NewThing()
			p.BaseStruct.Unpack(prot)
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes TestBase objects.
func (p *TestBase) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("TestBase")
	prot.PackFieldBegin("base_struct", frugal.STRUCT, 1)
	p.BaseStruct.Pack(prot)
	prot.PackFieldEnd(1)
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// TestLowercase is a frual serializable object.
type TestLowercase struct {
	LowercaseInt int32
}

// NewTestLowercase constructs a TestLowercase.
func NewTestLowercase() *TestLowercase {
	return &TestLowercase{
		// TODO: default values

	}
}

// Unpack deserializes TestLowercase objects.
func (p *TestLowercase) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("TestLowercase")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.LowercaseInt = prot.UnpackI32()
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes TestLowercase objects.
func (p *TestLowercase) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("TestLowercase")
	prot.PackI32("lowercaseInt", 1, int32(p.LowercaseInt))
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// Event is a frual serializable object.
type Event struct {
	ID      ID
	Message string
}

// NewEvent constructs a Event.
func NewEvent() *Event {
	return &Event{
		// TODO: default values

	}
}

// Unpack deserializes Event objects.
func (p *Event) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("Event")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.ID = ID(prot.UnpackI64())
		case 2:
			p.Message = prot.UnpackString()
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes Event objects.
func (p *Event) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("Event")
	prot.PackI64("ID", 1, int64(p.ID))
	prot.PackString("Message", 2, string(p.Message))
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// TestingDefaults is a frual serializable object.
type TestingDefaults struct {
	ID2        ID
	Ev1        *Event
	Ev2        *Event
	ID         ID
	Thing      string
	Thing2     string
	Listfield  []Int
	ID3        ID
	BinField   []byte
	BinField2  []byte
	BinField3  []byte
	BinField4  []byte
	List2      *[]Int
	List3      []Int
	List4      []Int
	AMap       *map[string]string
	Status     HealthCondition
	BaseStatus base.BaseHealthCondition
}

// NewTestingDefaults constructs a TestingDefaults.
func NewTestingDefaults() *TestingDefaults {
	return &TestingDefaults{
		// TODO: default values

	}
}

// Unpack deserializes TestingDefaults objects.
func (p *TestingDefaults) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("TestingDefaults")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.ID2 = ID(prot.UnpackI64())
		case 2:
			p.Ev1 = NewEvent()
			p.Ev1.Unpack(prot)
		case 3:
			p.Ev2 = NewEvent()
			p.Ev2.Unpack(prot)
		case 4:
			p.ID = ID(prot.UnpackI64())
		case 5:
			p.Thing = prot.UnpackString()
		case 6:
			p.Thing2 = prot.UnpackString()
		case 7:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.Listfield = make([]Int, size)
				for i := 0; i < size; i++ {
					(p.Listfield)[i] = Int(prot.UnpackI32())
				}
			}
			prot.UnpackListEnd()
		case 8:
			p.ID3 = ID(prot.UnpackI64())
		case 9:
			p.BinField = prot.UnpackBinary()
		case 10:
			p.BinField2 = prot.UnpackBinary()
		case 11:
			p.BinField3 = prot.UnpackBinary()
		case 12:
			p.BinField4 = prot.UnpackBinary()
		case 13:
			size := prot.UnpackListBegin()
			if size > 0 {
				temp := make([]Int, size)
				p.List2 = &temp
				for i := 0; i < size; i++ {
					(*p.List2)[i] = Int(prot.UnpackI32())
				}
			}
			prot.UnpackListEnd()
		case 14:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.List3 = make([]Int, size)
				for i := 0; i < size; i++ {
					(p.List3)[i] = Int(prot.UnpackI32())
				}
			}
			prot.UnpackListEnd()
		case 15:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.List4 = make([]Int, size)
				for i := 0; i < size; i++ {
					(p.List4)[i] = Int(prot.UnpackI32())
				}
			}
			prot.UnpackListEnd()
		case 16:
			size := prot.UnpackMapBegin()
			temp := make(map[string]string, size)
			p.AMap = &temp
			for i := 0; i < size; i++ {
				elem4 := prot.UnpackString()
				elem5 := prot.UnpackString()
				(*p.AMap)[elem4] = elem5
			}
			prot.UnpackMapEnd()
		case 17:
			p.Status = HealthCondition(prot.UnpackI32())
		case 18:
			p.BaseStatus = base.BaseHealthCondition(prot.UnpackI32())
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes TestingDefaults objects.
func (p *TestingDefaults) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("TestingDefaults")
	if p.ID2 != nil {
		prot.PackI64("ID2", 1, int64(p.ID2))
	}
	prot.PackFieldBegin("ev1", frugal.STRUCT, 2)
	p.Ev1.Pack(prot)
	prot.PackFieldEnd(2)
	prot.PackFieldBegin("ev2", frugal.STRUCT, 3)
	p.Ev2.Pack(prot)
	prot.PackFieldEnd(3)
	prot.PackI64("ID", 4, int64(p.ID))
	prot.PackString("thing", 5, string(p.Thing))
	if p.Thing2 != nil {
		prot.PackString("thing2", 6, string(p.Thing2))
	}
	prot.PackListBegin("listfield", 7, frugal.I32, len(p.Listfield))
	for _, v := range p.Listfield {
		prot.PackI32("", -1, int32(v))
	}
	prot.PackListEnd(7)
	prot.PackI64("ID3", 8, int64(p.ID3))
	prot.PackBinary("bin_field", 9, []byte(p.BinField))
	if p.BinField2 != nil {
		prot.PackBinary("bin_field2", 10, []byte(p.BinField2))
	}
	prot.PackBinary("bin_field3", 11, []byte(p.BinField3))
	if p.BinField4 != nil {
		prot.PackBinary("bin_field4", 12, []byte(p.BinField4))
	}
	if p.List2 != nil {
		prot.PackListBegin("list2", 13, frugal.I32, len(*p.List2))
		for _, v := range *p.List2 {
			prot.PackI32("", -1, int32(v))
		}
		prot.PackListEnd(13)
	}
	if p.List3 != nil {
		prot.PackListBegin("list3", 14, frugal.I32, len(p.List3))
		for _, v := range p.List3 {
			prot.PackI32("", -1, int32(v))
		}
		prot.PackListEnd(14)
	}
	prot.PackListBegin("list4", 15, frugal.I32, len(p.List4))
	for _, v := range p.List4 {
		prot.PackI32("", -1, int32(v))
	}
	prot.PackListEnd(15)
	if p.AMap != nil {
		prot.PackMapBegin("a_map", 16, frugal.STRING, frugal.STRING, len(*p.AMap))
		for k, v := range *p.AMap {
			prot.PackString("", -1, string(k))
			prot.PackString("", -1, string(v))
		}
		prot.PackMapEnd(16)
	}
	prot.PackI32("status", 17, int32(p.Status))
	prot.PackI32("base_status", 18, int32(p.BaseStatus))
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// EventWrapper is a frual serializable object.
type EventWrapper struct {
	ID               *ID
	Ev               *Event
	Events           []*Event
	Events2          map[*Event]bool
	EventMap         map[ID]*Event
	Nums             [][]Int
	Enums            []ItsAnEnum
	ABoolField       bool
	AUnion           *TestingUnions
	TypedefOfTypedef T2String
	Depr             bool
	DeprBinary       []byte
	DeprList         []bool
}

// NewEventWrapper constructs a EventWrapper.
func NewEventWrapper() *EventWrapper {
	return &EventWrapper{
		// TODO: default values

	}
}

// Unpack deserializes EventWrapper objects.
func (p *EventWrapper) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("EventWrapper")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			temp := ID(prot.UnpackI64())
			p.ID = &temp
		case 2:
			p.Ev = NewEvent()
			p.Ev.Unpack(prot)
		case 3:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.Events = make([]*Event, size)
				for i := 0; i < size; i++ {
					(p.Events)[i] = NewEvent()
					(p.Events)[i].Unpack(prot)
				}
			}
			prot.UnpackListEnd()
		case 4:
			// TODO: sets! Events2
		case 5:
			size := prot.UnpackMapBegin()
			p.EventMap = make(map[ID]*Event, size)
			for i := 0; i < size; i++ {
				elem7 := ID(prot.UnpackI64())
				elem8 := NewEvent()
				elem8.Unpack(prot)
				(p.EventMap)[elem7] = elem8
			}
			prot.UnpackMapEnd()
		case 6:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.Nums = make([][]Int, size)
				for i := 0; i < size; i++ {
					size := prot.UnpackListBegin()
					if size > 0 {
						(p.Nums)[i] = make([]Int, size)
						for i := 0; i < size; i++ {
							((p.Nums)[i])[i] = Int(prot.UnpackI32())
						}
					}
					prot.UnpackListEnd()
				}
			}
			prot.UnpackListEnd()
		case 7:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.Enums = make([]ItsAnEnum, size)
				for i := 0; i < size; i++ {
					(p.Enums)[i] = ItsAnEnum(prot.UnpackI32())
				}
			}
			prot.UnpackListEnd()
		case 8:
			p.ABoolField = prot.UnpackBool()
		case 9:
			p.AUnion = NewTestingUnions()
			p.AUnion.Unpack(prot)
		case 10:
			p.TypedefOfTypedef = T2String(prot.UnpackString())
		case 11:
			p.Depr = prot.UnpackBool()
		case 12:
			p.DeprBinary = prot.UnpackBinary()
		case 13:
			size := prot.UnpackListBegin()
			if size > 0 {
				p.DeprList = make([]bool, size)
				for i := 0; i < size; i++ {
					(p.DeprList)[i] = prot.UnpackBool()
				}
			}
			prot.UnpackListEnd()
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes EventWrapper objects.
func (p *EventWrapper) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("EventWrapper")
	if p.ID != nil {
		prot.PackI64("ID", 1, int64(*p.ID))
	}
	prot.PackFieldBegin("Ev", frugal.STRUCT, 2)
	p.Ev.Pack(prot)
	prot.PackFieldEnd(2)
	prot.PackListBegin("Events", 3, frugal.STRUCT, len(p.Events))
	for _, v := range p.Events {
		prot.PackFieldBegin("", frugal.STRUCT, -1)
		v.Pack(prot)
		prot.PackFieldEnd(-1)
	}
	prot.PackListEnd(3)
	// TODO: sets p.Events2
	prot.PackMapBegin("EventMap", 5, frugal.I64, frugal.STRUCT, len(p.EventMap))
	for k, v := range p.EventMap {
		prot.PackI64("", -1, int64(k))
		prot.PackFieldBegin("", frugal.STRUCT, -1)
		v.Pack(prot)
		prot.PackFieldEnd(-1)
	}
	prot.PackMapEnd(5)
	prot.PackListBegin("Nums", 6, frugal.LIST, len(p.Nums))
	for _, v := range p.Nums {
		prot.PackListBegin("", -1, frugal.I32, len(v))
		for _, v := range v {
			prot.PackI32("", -1, int32(v))
		}
		prot.PackListEnd(-1)
	}
	prot.PackListEnd(6)
	prot.PackListBegin("Enums", 7, frugal.I32, len(p.Enums))
	for _, v := range p.Enums {
		prot.PackI32("", -1, int32(v))
	}
	prot.PackListEnd(7)
	prot.PackBool("aBoolField", 8, bool(p.ABoolField))
	prot.PackFieldBegin("a_union", frugal.STRUCT, 9)
	p.AUnion.Pack(prot)
	prot.PackFieldEnd(9)
	prot.PackString("typedefOfTypedef", 10, string(p.TypedefOfTypedef))
	prot.PackBool("depr", 11, bool(p.Depr))
	prot.PackBinary("deprBinary", 12, []byte(p.DeprBinary))
	prot.PackListBegin("deprList", 13, frugal.BOOL, len(p.DeprList))
	for _, v := range p.DeprList {
		prot.PackBool("", -1, bool(v))
	}
	prot.PackListEnd(13)
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// FooArgs is a frual serializable object.
type FooArgs struct {
	NewMessage    string
	MessageArgs   string
	MessageResult string
}

// NewFooArgs constructs a FooArgs.
func NewFooArgs() *FooArgs {
	return &FooArgs{
		// TODO: default values

	}
}

// Unpack deserializes FooArgs objects.
func (p *FooArgs) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("FooArgs")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.NewMessage = prot.UnpackString()
		case 2:
			p.MessageArgs = prot.UnpackString()
		case 3:
			p.MessageResult = prot.UnpackString()
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes FooArgs objects.
func (p *FooArgs) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("FooArgs")
	prot.PackString("newMessage", 1, string(p.NewMessage))
	prot.PackString("messageArgs", 2, string(p.MessageArgs))
	prot.PackString("messageResult", 3, string(p.MessageResult))
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// TestingUnions is a frual serializable object.
type TestingUnions struct {
	AnID            *ID
	AString         *string
	Someotherthing  *Int
	AnInt16         *int16
	Requests        Request
	BinFieldInUnion []byte
	Depr            *bool
}

// NewTestingUnions constructs a TestingUnions.
func NewTestingUnions() *TestingUnions {
	return &TestingUnions{
		// TODO: default values

	}
}

// Unpack deserializes TestingUnions objects.
func (p *TestingUnions) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("TestingUnions")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			temp := ID(prot.UnpackI64())
			p.AnID = &temp
		case 2:
			v := prot.UnpackString()
			p.AString = &v
		case 3:
			temp := Int(prot.UnpackI32())
			p.Someotherthing = &temp
		case 4:
			v := prot.UnpackI16()
			p.AnInt16 = &v
		case 5:
			size := prot.UnpackMapBegin()
			p.Requests = make(Request, size)
			for i := 0; i < size; i++ {
				elem13 := Int(prot.UnpackI32())
				elem14 := prot.UnpackString()
				(p.Requests)[elem13] = elem14
			}
			prot.UnpackMapEnd()
		case 6:
			p.BinFieldInUnion = prot.UnpackBinary()
		case 7:
			v := prot.UnpackBool()
			p.Depr = &v
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes TestingUnions objects.
func (p *TestingUnions) Pack(prot frugal.Protocol) {
	count := 0
	if p.AnID != nil {
		count++
	}
	if p.AString != nil {
		count++
	}
	if p.Someotherthing != nil {
		count++
	}
	if p.AnInt16 != nil {
		count++
	}
	if p.Requests != nil {
		count++
	}
	if p.BinFieldInUnion != nil {
		count++
	}
	if p.Depr != nil {
		count++
	}
	if count != 1 {
		prot.Set(errors.New("TestingUnions invalid union state"))
		return
	}
	prot.PackStructBegin("TestingUnions")
	if p.AnID != nil {
		prot.PackI64("AnID", 1, int64(*p.AnID))
	}
	if p.AString != nil {
		prot.PackString("aString", 2, string(*p.AString))
	}
	if p.Someotherthing != nil {
		prot.PackI32("someotherthing", 3, int32(*p.Someotherthing))
	}
	if p.AnInt16 != nil {
		prot.PackI16("AnInt16", 4, int16(*p.AnInt16))
	}
	if p.Requests != nil {
		prot.PackMapBegin("Requests", 5, frugal.I32, frugal.STRING, len(p.Requests))
		for k, v := range p.Requests {
			prot.PackI32("", -1, int32(k))
			prot.PackString("", -1, string(v))
		}
		prot.PackMapEnd(5)
	}
	if p.BinFieldInUnion != nil {
		prot.PackBinary("bin_field_in_union", 6, []byte(p.BinFieldInUnion))
	}
	if p.Depr != nil {
		prot.PackBool("depr", 7, bool(*p.Depr))
	}
	prot.PackFieldStop()
	prot.PackStructEnd()
}

// AwesomeException is a frual serializable object.
type AwesomeException struct {
	ID     ID
	Reason string
	Depr   bool
}

// NewAwesomeException constructs a AwesomeException.
func NewAwesomeException() *AwesomeException {
	return &AwesomeException{
		// TODO: default values

	}
}

// Unpack deserializes AwesomeException objects.
func (p *AwesomeException) Unpack(prot frugal.Protocol) {
	prot.UnpackStructBegin("AwesomeException")
	for typeID, id := prot.UnpackFieldBegin(); typeID != frugal.STOP; typeID, id = prot.UnpackFieldBegin() {
		switch id {
		case 1:
			p.ID = ID(prot.UnpackI64())
		case 2:
			p.Reason = prot.UnpackString()
		case 3:
			p.Depr = prot.UnpackBool()
		default:
			prot.Skip(typeID)
		}
		prot.UnpackFieldEnd()
	}
	prot.UnpackStructEnd()
}

// Pack serializes AwesomeException objects.
func (p *AwesomeException) Pack(prot frugal.Protocol) {
	prot.PackStructBegin("AwesomeException")
	prot.PackI64("ID", 1, int64(p.ID))
	prot.PackString("Reason", 2, string(p.Reason))
	prot.PackBool("depr", 3, bool(p.Depr))
	prot.PackFieldStop()
	prot.PackStructEnd()
}

func (p *AwesomeException) Error() string {
	return "TODO: generate errorz"
}
