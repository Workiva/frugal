package filter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/Workiva/frugal/compiler/parser"
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
	ts := getTestStructs(t)
	artistType := ts.album.Fields[0].Type
	require.NotNil(t, artistType)
	assert.Equal(t, `Artist`, artistType.Name)
	songType := ts.album.Fields[1].Type.ValueType
	require.NotNil(t, songType)
	assert.Equal(t, `Song`, songType.Name)
	yearType := ts.album.Fields[2].Type.KeyType
	require.NotNil(t, yearType)
	assert.Equal(t, `Year`, yearType.Name)
	placeType := ts.album.Fields[2].Type.ValueType
	require.NotNil(t, placeType)
	assert.Equal(t, `Place`, placeType.Name)

	testCases := []struct {
		inputType   *parser.Type
		inputStruct *parser.Struct
		expVal      bool
	}{{
		inputType:   nil,
		inputStruct: ts.album,
		expVal:      false,
	}, {
		inputType:   artistType,
		inputStruct: ts.artist,
		expVal:      true,
	}, {
		inputType:   songType,
		inputStruct: ts.song,
		expVal:      true,
	}, {
		inputType:   yearType,
		inputStruct: ts.year,
		expVal:      true,
	}, {
		inputType:   placeType,
		inputStruct: ts.place,
		expVal:      true,
	}, {
		inputType:   songType,
		inputStruct: ts.artist,
		expVal:      false,
	}, {
		inputType:   yearType,
		inputStruct: ts.artist,
		expVal:      false,
	}, {
		inputType:   placeType,
		inputStruct: ts.artist,
		expVal:      false,
	}}

	for _, tc := range testCases {
		msg := fmt.Sprintf("testing %+v with type %+v", tc.inputStruct, tc.inputType)
		assert.Equal(t, tc.expVal, typeContainsStruct(tc.inputType, tc.inputStruct), msg)

		f := &parser.Field{Type: tc.inputType}
		assert.Equal(t, tc.expVal, fieldContainsStruct(f, tc.inputStruct), msg)

		fs := make([]*parser.Field, 10)
		fs[1] = &parser.Field{}
		fs[4] = f
		assert.Equal(t, tc.expVal, anyFieldContainsStruct(fs, tc.inputStruct), msg)
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
	ts := getTestStructs(t)

	assert.True(t, isStructUsedByService(ts.artist, ts.musicService))
	assert.False(t, isStructUsedByService(ts.burrito, ts.musicService))
}

func TestIsStructUsedByAnyService(t *testing.T) {
	ts := getTestStructs(t)

	assert.True(t, isStructUsedByAnyService(ts.artist, ts.fileFrugal.Services))
	assert.False(t, isStructUsedByAnyService(ts.burrito, ts.fileFrugal.Services))
}

func TestGetNeededStructs(t *testing.T) {
	// the yaml specifies that we want Burrito
	// but none of the other structs.
	testYaml := `included:
  structs:
    names:
      - Burrito`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)
	ts := getTestStructs(t)

	actStructs := getNeededStructs(fs, ts.fileFrugal)
	// The MusicService still needs all of the other structs, so it
	// should have all of the structs in the IDL.
	assert.Len(t, actStructs, 7)

	// now let's remove getAlbum and getTop5Albums from the *parser.Frugal
	// so we don't need _all_ of the structs anymore.
	for i := range ts.musicService.Methods {
		switch ts.musicService.Methods[i] {
		case ts.getAlbum, ts.getTop5Albums:
			ts.musicService.Methods[i] = nil
		}
	}
	actStructs = getNeededStructs(fs, ts.fileFrugal)
	// Now it should only have:
	// - Burrito (as requested)
	// - Artist (used by getArtist)
	// - ArtistError (used by getArtist)
	assert.Len(t, actStructs, 3)
}

func TestStructSliceIncludes(t *testing.T) {
	assert.False(t, structListContains(nil, nil))
	assert.False(t, structListContains([]*parser.Struct{{}}, nil))

	artist := &parser.Struct{
		Name: `Artist`,
		Type: parser.StructTypeStruct,
	}
	artistException := &parser.Struct{
		Name: `Artist`,
		// Surprise! It's an exception with the same name!
		Type: parser.StructTypeException,
	}
	assert.False(t, structListContains([]*parser.Struct{artist}, artistException))
	assert.True(t, structListContains([]*parser.Struct{artist}, artist))
	assert.True(t, structListContains([]*parser.Struct{artist, artistException}, artistException))
}
