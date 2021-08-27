package filter

import (
	"testing"

	"github.com/Workiva/frugal/compiler/parser"
	"github.com/stretchr/testify/assert"
)

func TestStructSpecIsStructSpecified(t *testing.T) {
	s := &parser.Struct{Name: `josh`}
	var ss *structSpec
	assert.False(t, ss.isStructSpecified(s))

	ss = &structSpec{}
	assert.False(t, ss.isStructSpecified(s))

	b := true
	ss.All = &b
	assert.True(t, ss.isStructSpecified(s))

	b = false
	assert.False(t, ss.isStructSpecified(s))

	ss.Names = append(ss.Names, `notjosh`)
	assert.False(t, ss.isStructSpecified(s))

	ss.Names = append(ss.Names, `josh`)
	assert.True(t, ss.isStructSpecified(s))
}

func TestStructExistsHelpers(t *testing.T) {
	josh := &parser.Struct{Name: `josh`}

	joshType := &parser.Type{Name: `josh`}
	notJoshType := &parser.Type{Name: `notjosh`}

	testCases := []struct {
		inputStruct *parser.Struct
		inputType   *parser.Type
		expVal      bool
	}{{
		inputStruct: josh,
		inputType:   nil,
		expVal:      false,
	}, {
		inputStruct: josh,
		inputType:   joshType,
		expVal:      true,
	}, {
		inputStruct: josh,
		inputType:   &parser.Type{Name: `hasJoshInSlice`, ValueType: joshType},
		expVal:      true,
	}, {
		inputStruct: josh,
		inputType:   &parser.Type{Name: `hasASliceOfNotJosh`, ValueType: notJoshType},
		expVal:      false,
	}, {
		inputStruct: josh,
		inputType:   &parser.Type{Name: `hasJoshAsMapKey`, KeyType: joshType, ValueType: notJoshType},
		expVal:      true,
	}, {
		inputStruct: josh,
		inputType:   &parser.Type{Name: `withAMapWithoutJosh`, KeyType: notJoshType, ValueType: notJoshType},
		expVal:      false,
	}}

	for _, tc := range testCases {
		//
		assert.Equal(t, tc.expVal, structExistsInType(tc.inputStruct, tc.inputType))

		f := &parser.Field{Type: tc.inputType}
		assert.Equal(t, tc.expVal, structExistsInField(tc.inputStruct, f))

		fs := make([]*parser.Field, 10)
		fs[1] = &parser.Field{}
		fs[4] = f
		assert.Equal(t, tc.expVal, structExistsInAnyField(tc.inputStruct, fs))
	}
}

func TestGetAllSubstructs(t *testing.T) {
	in := []*parser.Struct{{
		Name: `Burger`,
		Fields: []*parser.Field{{
			Name: `Bun`,
			Type: &parser.Type{
				Name: `Bread`,
			},
		}, {
			Name: `Meat`,
			Type: &parser.Type{
				Name: `Protein`,
			},
		}},
	}, {
		Name: `Bread`,
	}}
	out := []*parser.Struct{{
		Name: `Vegetable`,
	}, {
		Name: `Protein`,
	}}

	act := getAllSubstructs(in, out)

	exp := make([]*parser.Struct, len(in))
	copy(exp, in)
	exp = append(exp, out[1])

	assert.Equal(t, exp, act)
}

func TestIsStructUsedByService(t *testing.T) {
	// TODO
}

func TestIsStructUsedByAnyService(t *testing.T) {
	// TODO
}

func TestGetNeededStructs(t *testing.T) {
	// TODO
}
