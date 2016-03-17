// Autogenerated by Frugal Compiler (1.0.6)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package event

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal/example/go/gen-go/base"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = bytes.Equal

var _ = base.GoUnusedProtection__
var GoUnusedProtection__ int

var DEFAULT_ID ID

const Thirtyfour = 34

var MAPCONSTANT map[string]string

var ConstEvent1 *Event

var ConstEvent2 *Event

var NumsList []int32

var NumsSet map[int32]bool

var MAPCONSTANT2 map[string]*Event

func init() {
	DEFAULT_ID = -1
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

	NumsSet = map[int32]bool{
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

}

type ID int64
type Int int32
type Request map[Int]string
type ItsAnEnum int64

const (
	ItsAnEnum_FIRST  ItsAnEnum = 2
	ItsAnEnum_SECOND ItsAnEnum = 3
	ItsAnEnum_THIRD  ItsAnEnum = 4
)

func (p ItsAnEnum) String() string {
	switch p {
	case ItsAnEnum_THIRD:
		return "THIRD"
	case ItsAnEnum_FIRST:
		return "FIRST"
	case ItsAnEnum_SECOND:
		return "SECOND"
	}
	return "<UNSET>"
}

func ItsAnEnumFromString(s string) (ItsAnEnum, error) {
	switch s {
	case "FIRST":
		return ItsAnEnum_FIRST, nil
	case "SECOND":
		return ItsAnEnum_SECOND, nil
	case "THIRD":
		return ItsAnEnum_THIRD, nil
	}
	return ItsAnEnum(0), fmt.Errorf("not a valid ItsAnEnum string")
}

func (p ItsAnEnum) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *ItsAnEnum) UnmarshalText(text []byte) error {
	q, err := ItsAnEnumFromString(string(text))
	if err != nil {
		return err
	}
	*p = q
	return nil
}

func (p *ItsAnEnum) Scan(value interface{}) error {
	v, ok := value.(int64)
	if !ok {
		return errors.New("Scan value is not int64")
	}
	*p = ItsAnEnum(v)
	return nil
}

func (p *ItsAnEnum) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return int64(*p), nil
}

// This docstring gets added to the generated code because it has
// the @ sign.
type Event struct {
	// ID is a unique identifier for an event.
	ID ID `thrift:"ID,1" db:"ID" json:"ID"`
	// Message contains the event payload.
	Message string `thrift:"Message,2" db:"Message" json:"Message"`
}

func NewEvent() *Event {
	return &Event{
		ID: DEFAULT_ID,
	}
}

func (p *Event) GetID() ID {
	return p.ID
}

func (p *Event) GetMessage() string {
	return p.Message
}

func (p *Event) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *Event) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.ID = ID(v)
	}
	return nil
}

func (p *Event) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.Message = v
	}
	return nil
}

func (p *Event) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("Event"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *Event) writeField1(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("ID", thrift.I64, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:ID: ", p), err)
	}
	if err := oprot.WriteI64(int64(p.ID)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.ID (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:ID: ", p), err)
	}
	return nil
}

func (p *Event) writeField2(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Message", thrift.STRING, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:Message: ", p), err)
	}
	if err := oprot.WriteString(string(p.Message)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.Message (2) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:Message: ", p), err)
	}
	return nil
}

func (p *Event) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Event(%+v)", *p)
}

type EventWrapper struct {
	ID       *ID             `thrift:"ID,1" db:"ID" json:"ID,omitempty"`
	Ev       *Event          `thrift:"Ev,2" db:"Ev" json:"Ev"`
	Events   []*Event        `thrift:"Events,3" db:"Events" json:"Events"`
	Events2  map[*Event]bool `thrift:"Events2,4" db:"Events2" json:"Events2"`
	EventMap map[ID]*Event   `thrift:"EventMap,5" db:"EventMap" json:"EventMap"`
	Nums     [][]Int         `thrift:"Nums,6" db:"Nums" json:"Nums"`
	Enums    []ItsAnEnum     `thrift:"Enums,7" db:"Enums" json:"Enums"`
}

func NewEventWrapper() *EventWrapper {
	return &EventWrapper{}
}

var EventWrapper_ID_DEFAULT ID

func (p *EventWrapper) IsSetID() bool {
	return p.ID != nil
}

func (p *EventWrapper) GetID() ID {
	if !p.IsSetID() {
		return EventWrapper_ID_DEFAULT
	}
	return *p.ID
}

var EventWrapper_Ev_DEFAULT *Event

func (p *EventWrapper) IsSetEv() bool {
	return p.Ev != nil
}

func (p *EventWrapper) GetEv() *Event {
	if !p.IsSetEv() {
		return EventWrapper_Ev_DEFAULT
	}
	return p.Ev
}

func (p *EventWrapper) GetEvents() []*Event {
	return p.Events
}

func (p *EventWrapper) GetEvents2() map[*Event]bool {
	return p.Events2
}

func (p *EventWrapper) GetEventMap() map[ID]*Event {
	return p.EventMap
}

func (p *EventWrapper) GetNums() [][]Int {
	return p.Nums
}

func (p *EventWrapper) GetEnums() []ItsAnEnum {
	return p.Enums
}

func (p *EventWrapper) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		case 3:
			if err := p.ReadField3(iprot); err != nil {
				return err
			}
		case 4:
			if err := p.ReadField4(iprot); err != nil {
				return err
			}
		case 5:
			if err := p.ReadField5(iprot); err != nil {
				return err
			}
		case 6:
			if err := p.ReadField6(iprot); err != nil {
				return err
			}
		case 7:
			if err := p.ReadField7(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *EventWrapper) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.ID = &ID(v)
	}
	return nil
}

func (p *EventWrapper) ReadField2(iprot thrift.TProtocol) error {
	p.Ev = NewEvent()
	if err := p.Ev.Read(iprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Ev), err)
	}
	return nil
}

func (p *EventWrapper) ReadField3(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return thrift.PrependError("error reading list begin: ", err)
	}
	p.Events = make([]*Event, 0, size)
	for i := 0; i < size; i++ {
		elem0 := NewEvent()
		if err := elem0.Read(iprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", elem0), err)
		}
		p.Events = append(p.Events, elem0)

	}
	if err := iprot.ReadListEnd(); err != nil {
		return thrift.PrependError("error reading list end: ", err)
	}
	return nil
}

func (p *EventWrapper) ReadField4(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadSetBegin()
	if err != nil {
		return thrift.PrependError("error reading set begin: ", err)
	}
	p.Events2 = make(map[*Event]bool, 0, size)
	for i := 0; i < size; i++ {
		elem1 := NewEvent()
		if err := elem1.Read(iprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", elem1), err)
		}
		p.Events2[elem1] = true

	}
	if err := iprot.ReadSetEnd(); err != nil {
		return thrift.PrependError("error reading set end: ", err)
	}
	return nil
}

func (p *EventWrapper) ReadField5(iprot thrift.TProtocol) error {
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return thrift.PrependError("error reading map begin: ", err)
	}
	p.EventMap = make(map[ID]*Event, 0, size)
	for i := 0; i < size; i++ {
		var elem2 ID
		if v, err := iprot.ReadI64(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			elem2 = ID(v)
		}
		elem3 := NewEvent()
		if err := elem3.Read(iprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", elem3), err)
		}
		p.EventMap[elem2] = elem3

	}
	if err := iprot.ReadMapEnd(); err != nil {
		return thrift.PrependError("error reading map end: ", err)
	}
	return nil
}

func (p *EventWrapper) ReadField6(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return thrift.PrependError("error reading list begin: ", err)
	}
	p.Nums = make([][]Int, 0, size)
	for i := 0; i < size; i++ {
		_, size, err := iprot.ReadListBegin()
		if err != nil {
			return thrift.PrependError("error reading list begin: ", err)
		}
		elem4 := make([]Int, 0, size)
		for i := 0; i < size; i++ {
			var elem5 Int
			if v, err := iprot.ReadI32(); err != nil {
				return thrift.PrependError("error reading field 0: ", err)
			} else {
				elem5 = Int(v)
			}
			elem4 = append(elem4, elem5)

		}
		if err := iprot.ReadListEnd(); err != nil {
			return thrift.PrependError("error reading list end: ", err)
		}
		p.Nums = append(p.Nums, elem4)

	}
	if err := iprot.ReadListEnd(); err != nil {
		return thrift.PrependError("error reading list end: ", err)
	}
	return nil
}

func (p *EventWrapper) ReadField7(iprot thrift.TProtocol) error {
	_, size, err := iprot.ReadListBegin()
	if err != nil {
		return thrift.PrependError("error reading list begin: ", err)
	}
	p.Enums = make([]ItsAnEnum, 0, size)
	for i := 0; i < size; i++ {
		var elem6 ItsAnEnum
		if v, err := iprot.ReadI64(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			elem6 = ItsAnEnum(v)
		}
		p.Enums = append(p.Enums, elem6)

	}
	if err := iprot.ReadListEnd(); err != nil {
		return thrift.PrependError("error reading list end: ", err)
	}
	return nil
}

func (p *EventWrapper) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("EventWrapper"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := p.writeField3(oprot); err != nil {
		return err
	}
	if err := p.writeField4(oprot); err != nil {
		return err
	}
	if err := p.writeField5(oprot); err != nil {
		return err
	}
	if err := p.writeField6(oprot); err != nil {
		return err
	}
	if err := p.writeField7(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *EventWrapper) writeField1(oprot thrift.TProtocol) error {
	if p.IsSetID() {
		if err := oprot.WriteFieldBegin("ID", thrift.I64, 1); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:ID: ", p), err)
		}
		if err := oprot.WriteI64(int64(*p.ID)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.ID (1) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 1:ID: ", p), err)
		}
	}
	return nil
}

func (p *EventWrapper) writeField2(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Ev", thrift.STRUCT, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:Ev: ", p), err)
	}
	if err := p.Ev.Write(oprot); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Ev), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:Ev: ", p), err)
	}
	return nil
}

func (p *EventWrapper) writeField3(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Events", thrift.LIST, 3); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 3:Events: ", p), err)
	}
	if err := oprot.WriteListBegin(thrift.STRUCT, len(p.Events)); err != nil {
		return thrift.PrependError("error writing list begin: ", err)
	}
	for _, v := range p.Events {
		if err := v.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", v), err)
		}
	}
	if err := oprot.WriteListEnd(); err != nil {
		return thrift.PrependError("error writing list end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 3:Events: ", p), err)
	}
	return nil
}

func (p *EventWrapper) writeField4(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Events2", thrift.SET, 4); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 4:Events2: ", p), err)
	}
	if err := oprot.WriteSetBegin(thrift.STRUCT, len(p.Events2)); err != nil {
		return thrift.PrependError("error writing set begin: ", err)
	}
	for v, _ := range p.Events2 {
		if err := v.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", v), err)
		}
	}
	if err := oprot.WriteSetEnd(); err != nil {
		return thrift.PrependError("error writing set end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 4:Events2: ", p), err)
	}
	return nil
}

func (p *EventWrapper) writeField5(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("EventMap", thrift.MAP, 5); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 5:EventMap: ", p), err)
	}
	if err := oprot.WriteMapBegin(thrift.I64, thrift.STRUCT, len(p.EventMap)); err != nil {
		return thrift.PrependError("error writing map begin: ", err)
	}
	for k, v := range p.EventMap {
		if err := oprot.WriteI64(int64(k)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
		}
		if err := v.Write(oprot); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", v), err)
		}
	}
	if err := oprot.WriteMapEnd(); err != nil {
		return thrift.PrependError("error writing map end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 5:EventMap: ", p), err)
	}
	return nil
}

func (p *EventWrapper) writeField6(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Nums", thrift.LIST, 6); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 6:Nums: ", p), err)
	}
	if err := oprot.WriteListBegin(thrift.LIST, len(p.Nums)); err != nil {
		return thrift.PrependError("error writing list begin: ", err)
	}
	for _, v := range p.Nums {
		if err := oprot.WriteListBegin(thrift.I32, len(v)); err != nil {
			return thrift.PrependError("error writing list begin: ", err)
		}
		for _, v := range v {
			if err := oprot.WriteI32(int32(v)); err != nil {
				return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
			}
		}
		if err := oprot.WriteListEnd(); err != nil {
			return thrift.PrependError("error writing list end: ", err)
		}
	}
	if err := oprot.WriteListEnd(); err != nil {
		return thrift.PrependError("error writing list end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 6:Nums: ", p), err)
	}
	return nil
}

func (p *EventWrapper) writeField7(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Enums", thrift.LIST, 7); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 7:Enums: ", p), err)
	}
	if err := oprot.WriteListBegin(thrift.I64, len(p.Enums)); err != nil {
		return thrift.PrependError("error writing list begin: ", err)
	}
	for _, v := range p.Enums {
		if err := oprot.WriteI64(int64(v)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
		}
	}
	if err := oprot.WriteListEnd(); err != nil {
		return thrift.PrependError("error writing list end: ", err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 7:Enums: ", p), err)
	}
	return nil
}

func (p *EventWrapper) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("EventWrapper(%+v)", *p)
}

type TestingUnions struct {
	AnID           *ID     `thrift:"AnID,1" db:"AnID" json:"AnID,omitempty"`
	AString        *string `thrift:"aString,2" db:"aString" json:"aString,omitempty"`
	Someotherthing *Int    `thrift:"someotherthing,3" db:"someotherthing" json:"someotherthing,omitempty"`
	AnInt16        *int16  `thrift:"AnInt16,4" db:"AnInt16" json:"AnInt16,omitempty"`
	Requests       Request `thrift:"Requests,5" db:"Requests" json:"Requests,omitempty"`
}

func NewTestingUnions() *TestingUnions {
	return &TestingUnions{}
}

var TestingUnions_AnID_DEFAULT ID

func (p *TestingUnions) IsSetAnID() bool {
	return p.AnID != nil
}

func (p *TestingUnions) GetAnID() ID {
	if !p.IsSetAnID() {
		return TestingUnions_AnID_DEFAULT
	}
	return *p.AnID
}

var TestingUnions_AString_DEFAULT string

func (p *TestingUnions) IsSetAString() bool {
	return p.AString != nil
}

func (p *TestingUnions) GetAString() string {
	if !p.IsSetAString() {
		return TestingUnions_AString_DEFAULT
	}
	return *p.AString
}

var TestingUnions_Someotherthing_DEFAULT Int

func (p *TestingUnions) IsSetSomeotherthing() bool {
	return p.Someotherthing != nil
}

func (p *TestingUnions) GetSomeotherthing() Int {
	if !p.IsSetSomeotherthing() {
		return TestingUnions_Someotherthing_DEFAULT
	}
	return *p.Someotherthing
}

var TestingUnions_AnInt16_DEFAULT int16

func (p *TestingUnions) IsSetAnInt16() bool {
	return p.AnInt16 != nil
}

func (p *TestingUnions) GetAnInt16() int16 {
	if !p.IsSetAnInt16() {
		return TestingUnions_AnInt16_DEFAULT
	}
	return *p.AnInt16
}

var TestingUnions_Requests_DEFAULT Request

func (p *TestingUnions) IsSetRequests() bool {
	return p.Requests != nil
}

func (p *TestingUnions) GetRequests() Request {
	return p.Requests
}

func (p *TestingUnions) CountSetFieldsTestingUnions() int {
	count := 0
	if p.IsSetAnID() {
		count++
	}
	if p.IsSetAString() {
		count++
	}
	if p.IsSetSomeotherthing() {
		count++
	}
	if p.IsSetAnInt16() {
		count++
	}
	if p.IsSetRequests() {
		count++
	}
	return count
}

func (p *TestingUnions) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		case 3:
			if err := p.ReadField3(iprot); err != nil {
				return err
			}
		case 4:
			if err := p.ReadField4(iprot); err != nil {
				return err
			}
		case 5:
			if err := p.ReadField5(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *TestingUnions) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.AnID = &ID(v)
	}
	return nil
}

func (p *TestingUnions) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.AString = &v
	}
	return nil
}

func (p *TestingUnions) ReadField3(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI32(); err != nil {
		return thrift.PrependError("error reading field 3: ", err)
	} else {
		p.Someotherthing = &Int(v)
	}
	return nil
}

func (p *TestingUnions) ReadField4(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI16(); err != nil {
		return thrift.PrependError("error reading field 4: ", err)
	} else {
		p.AnInt16 = &v
	}
	return nil
}

func (p *TestingUnions) ReadField5(iprot thrift.TProtocol) error {
	_, _, size, err := iprot.ReadMapBegin()
	if err != nil {
		return thrift.PrependError("error reading map begin: ", err)
	}
	p.Requests = make(Request, 0, size)
	for i := 0; i < size; i++ {
		var elem7 Int
		if v, err := iprot.ReadI32(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			elem7 = Int(v)
		}
		var elem8 string
		if v, err := iprot.ReadString(); err != nil {
			return thrift.PrependError("error reading field 0: ", err)
		} else {
			elem8 = v
		}
		p.Requests[elem7] = elem8

	}
	if err := iprot.ReadMapEnd(); err != nil {
		return thrift.PrependError("error reading map end: ", err)
	}
	return nil
}

func (p *TestingUnions) Write(oprot thrift.TProtocol) error {
	if c := p.CountSetFieldsTestingUnions(); c != 1 {
		fmt.Errorf("%T write union: exactly one field must be set (%d set).", p, c)
	}
	if err := oprot.WriteStructBegin("TestingUnions"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := p.writeField3(oprot); err != nil {
		return err
	}
	if err := p.writeField4(oprot); err != nil {
		return err
	}
	if err := p.writeField5(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *TestingUnions) writeField1(oprot thrift.TProtocol) error {
	if p.IsSetAnID() {
		if err := oprot.WriteFieldBegin("AnID", thrift.I64, 1); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:AnID: ", p), err)
		}
		if err := oprot.WriteI64(int64(*p.AnID)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.AnID (1) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 1:AnID: ", p), err)
		}
	}
	return nil
}

func (p *TestingUnions) writeField2(oprot thrift.TProtocol) error {
	if p.IsSetAString() {
		if err := oprot.WriteFieldBegin("aString", thrift.STRING, 2); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:aString: ", p), err)
		}
		if err := oprot.WriteString(string(*p.AString)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.aString (2) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 2:aString: ", p), err)
		}
	}
	return nil
}

func (p *TestingUnions) writeField3(oprot thrift.TProtocol) error {
	if p.IsSetSomeotherthing() {
		if err := oprot.WriteFieldBegin("someotherthing", thrift.I32, 3); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 3:someotherthing: ", p), err)
		}
		if err := oprot.WriteI32(int32(*p.Someotherthing)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.someotherthing (3) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 3:someotherthing: ", p), err)
		}
	}
	return nil
}

func (p *TestingUnions) writeField4(oprot thrift.TProtocol) error {
	if p.IsSetAnInt16() {
		if err := oprot.WriteFieldBegin("AnInt16", thrift.I16, 4); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 4:AnInt16: ", p), err)
		}
		if err := oprot.WriteI16(int16(*p.AnInt16)); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T.AnInt16 (4) field write error: ", p), err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 4:AnInt16: ", p), err)
		}
	}
	return nil
}

func (p *TestingUnions) writeField5(oprot thrift.TProtocol) error {
	if p.IsSetRequests() {
		if err := oprot.WriteFieldBegin("Requests", thrift.MAP, 5); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field begin error 5:Requests: ", p), err)
		}
		if err := oprot.WriteMapBegin(thrift.I32, thrift.STRING, len(p.Requests)); err != nil {
			return thrift.PrependError("error writing map begin: ", err)
		}
		for k, v := range p.Requests {
			if err := oprot.WriteI32(int32(k)); err != nil {
				return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
			}
			if err := oprot.WriteString(string(v)); err != nil {
				return thrift.PrependError(fmt.Sprintf("%T. (0) field write error: ", p), err)
			}
		}
		if err := oprot.WriteMapEnd(); err != nil {
			return thrift.PrependError("error writing map end: ", err)
		}
		if err := oprot.WriteFieldEnd(); err != nil {
			return thrift.PrependError(fmt.Sprintf("%T write field end error 5:Requests: ", p), err)
		}
	}
	return nil
}

func (p *TestingUnions) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("TestingUnions(%+v)", *p)
}

type AwesomeException struct {
	// ID is a unique identifier for an awesome exception.
	ID ID `thrift:"ID,1" db:"ID" json:"ID"`
	// Reason contains the error message.
	Reason string `thrift:"Reason,2" db:"Reason" json:"Reason"`
}

func NewAwesomeException() *AwesomeException {
	return &AwesomeException{}
}

func (p *AwesomeException) GetID() ID {
	return p.ID
}

func (p *AwesomeException) GetReason() string {
	return p.Reason
}

func (p *AwesomeException) Read(iprot thrift.TProtocol) error {
	if _, err := iprot.ReadStructBegin(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
	}

	for {
		_, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
		if err != nil {
			return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if err := p.ReadField1(iprot); err != nil {
				return err
			}
		case 2:
			if err := p.ReadField2(iprot); err != nil {
				return err
			}
		default:
			if err := iprot.Skip(fieldTypeId); err != nil {
				return err
			}
		}
		if err := iprot.ReadFieldEnd(); err != nil {
			return err
		}
	}
	if err := iprot.ReadStructEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
	}
	return nil
}

func (p *AwesomeException) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadI64(); err != nil {
		return thrift.PrependError("error reading field 1: ", err)
	} else {
		p.ID = ID(v)
	}
	return nil
}

func (p *AwesomeException) ReadField2(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return thrift.PrependError("error reading field 2: ", err)
	} else {
		p.Reason = v
	}
	return nil
}

func (p *AwesomeException) Write(oprot thrift.TProtocol) error {
	if err := oprot.WriteStructBegin("AwesomeException"); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
	}
	if err := p.writeField1(oprot); err != nil {
		return err
	}
	if err := p.writeField2(oprot); err != nil {
		return err
	}
	if err := oprot.WriteFieldStop(); err != nil {
		return thrift.PrependError("write field stop error: ", err)
	}
	if err := oprot.WriteStructEnd(); err != nil {
		return thrift.PrependError("write struct stop error: ", err)
	}
	return nil
}

func (p *AwesomeException) writeField1(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("ID", thrift.I64, 1); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:ID: ", p), err)
	}
	if err := oprot.WriteI64(int64(p.ID)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.ID (1) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 1:ID: ", p), err)
	}
	return nil
}

func (p *AwesomeException) writeField2(oprot thrift.TProtocol) error {
	if err := oprot.WriteFieldBegin("Reason", thrift.STRING, 2); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:Reason: ", p), err)
	}
	if err := oprot.WriteString(string(p.Reason)); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T.Reason (2) field write error: ", p), err)
	}
	if err := oprot.WriteFieldEnd(); err != nil {
		return thrift.PrependError(fmt.Sprintf("%T write field end error 2:Reason: ", p), err)
	}
	return nil
}

func (p *AwesomeException) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("AwesomeException(%+v)", *p)
}

func (p *AwesomeException) Error() string {
	return p.String()
}
