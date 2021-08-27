package filter

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Workiva/frugal/compiler/parser"
)

var (
	fBool = false
	tBool = true

	validYamlSpec = &filterSpec{
		Included: &definitionsSpec{
			Services: &servicesSpec{
				All: &fBool,
				Specs: []serviceSpec{{
					Name:   `myServiceA`,
					Entire: &fBool,
					Methods: []string{
						`myMethod1`,
						`myMethod2`,
					},
				}},
			},
			Structs: &structSpec{
				All: &fBool,
				Names: []string{
					`structA`,
					`structB`,
				},
			},
			Scopes: &scopesSpec{
				All: &tBool,
			},
		},
		Excluded: &definitionsSpec{
			Services: &servicesSpec{
				All: &tBool,
				Specs: []serviceSpec{{
					Name:   `myServiceB`,
					Entire: &tBool,
					Methods: []string{
						`myMethod6`,
						`myMethod7`,
					},
				}},
			},
			Structs: &structSpec{
				All: &tBool,
				Names: []string{
					`structZ`,
					`structY`,
				},
			},
			Scopes: &scopesSpec{
				All: &fBool,
			},
		},
	}
)

const (
	validYaml = `included:
  services:
    all: false
    specs:
      - name: myServiceA
        all: false
        methods:
          - myMethod1
          - myMethod2
  structs:
    all: false
    names:
      - structA
      - structB
  scopes:
    all: true
excluded:
  services:
    all: true
    specs:
    - name: myServiceB
      all: true
      methods:
        - myMethod6
        - myMethod7
  structs:
    all: true
    names:
      - structZ
      - structY
  scopes:
    all: false`
)

func TestNewYamlSpec(t *testing.T) {
	fs, err := newYamlSpec(`DNE`)
	require.Error(t, err, `file should not exist`)
	assert.Nil(t, fs, `should not have received a spec`)

	tmpFile, err := ioutil.TempFile(`.`, `TestNewYamlSpecTestFile`)
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	filename := tmpFile.Name()
	err = ioutil.WriteFile(filename, []byte(`invalid yaml!!`), 0)
	require.NoError(t, err)

	fs, err = newYamlSpec(filename)
	assert.Error(t, err, `file should exist, but have invalid yaml`)
	assert.Nil(t, fs, `yaml should NOT look good`)

	err = ioutil.WriteFile(filename, []byte(validYaml), 0)
	require.NoError(t, err)

	fs, err = newYamlSpec(filename)
	assert.NoError(t, err, `file should exist with valid yaml`)
	assert.NotNil(t, fs, `yaml should look good`)
	assert.Equal(t, validYamlSpec, fs)
}

func TestDefinitionsSpecIsServiceSpecified(t *testing.T) {
	s := &parser.Service{Name: `josh`}
	var ds *definitionsSpec
	assert.False(t, ds.isServiceSpecified(s))

	ds = &definitionsSpec{}
	assert.False(t, ds.isServiceSpecified(s))

	ds.Services = &servicesSpec{}
	assert.False(t, ds.isServiceSpecified(s))

	b := false
	ds.Services.All = &b
	b = true
	assert.True(t, ds.isServiceSpecified(s))
	b = false
	assert.False(t, ds.isServiceSpecified(s))

	ds.Services.Specs = append(ds.Services.Specs, serviceSpec{
		Name: `notjosh`,
	})
	assert.False(t, ds.isServiceSpecified(s))

	ds.Services.Specs = append(ds.Services.Specs, serviceSpec{
		Name: `josh`,
	})
	assert.True(t, ds.isServiceSpecified(s))
}

func TestDefinitionsSpecIsEntireServiceSpecified(t *testing.T) {
	s := &parser.Service{Name: `josh`}
	var ds *definitionsSpec
	assert.False(t, ds.isEntireServiceSpecified(s))

	ds = &definitionsSpec{}
	assert.False(t, ds.isEntireServiceSpecified(s))

	ds.Services = &servicesSpec{}
	assert.False(t, ds.isEntireServiceSpecified(s))

	b := false
	ds.Services.All = &b
	b = true
	assert.True(t, ds.isEntireServiceSpecified(s))
	b = false
	assert.False(t, ds.isEntireServiceSpecified(s))

	ds.Services.Specs = append(ds.Services.Specs, serviceSpec{
		Name: `notjosh`,
	})
	assert.False(t, ds.isEntireServiceSpecified(s))

	ds.Services.Specs = append(ds.Services.Specs, serviceSpec{
		Name: `josh`,
	})
	assert.False(t, ds.isEntireServiceSpecified(s))

	b2 := true
	ds.Services.Specs[1].Entire = &b2
	assert.True(t, ds.isEntireServiceSpecified(s))

	b2 = false
	assert.False(t, ds.isEntireServiceSpecified(s))
}

func TestDefinitionsSpecIsEntireScopeSpecified(t *testing.T) {
	s := &parser.Scope{Name: `josh`}
	var ds *definitionsSpec
	assert.False(t, ds.isEntireScopeSpecified(s))

	ds = &definitionsSpec{}
	assert.False(t, ds.isEntireScopeSpecified(s))

	ds.Scopes = &scopesSpec{}
	assert.False(t, ds.isEntireScopeSpecified(s))

	b := false
	ds.Scopes.All = &b
	assert.False(t, ds.isEntireScopeSpecified(s))

	b = true
	assert.True(t, ds.isEntireScopeSpecified(s))
}

func TestDefinitionsSpecIsStructSpecified(t *testing.T) {
	s := &parser.Struct{Name: `josh`}
	var ds *definitionsSpec
	assert.False(t, ds.isStructSpecified(s))

	ds = &definitionsSpec{}
	assert.False(t, ds.isStructSpecified(s))

	ds.Structs = &structSpec{}
	assert.False(t, ds.isStructSpecified(s))

	b := true
	ds.Structs.All = &b
	assert.True(t, ds.isStructSpecified(s))

	b = false
	assert.False(t, ds.isStructSpecified(s))

	ds.Structs.Names = append(ds.Structs.Names, `notjosh`)
	assert.False(t, ds.isStructSpecified(s))

	ds.Structs.Names = append(ds.Structs.Names, `josh`)
	assert.True(t, ds.isStructSpecified(s))
}
