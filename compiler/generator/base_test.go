package generator

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Workiva/frugal/compiler/generator/filter"
)

func TestBaseGeneratorFilterInput(t *testing.T) {
	ts := filter.GetTestStructs(t)
	bg := &BaseGenerator{}

	// no issues when the options aren't specified
	bg.FilterInput(nil)

	// doesn't do anything when the debug filter is on
	bg.Options = map[string]string{
		`debugFilter`: ``,
	}
	bg.FilterInput(nil)
	delete(bg.Options, `debugFilter`)

	// swallows errors when given a non-existent filename
	bg.Options[`filter_yaml`] = `dne`
	bg.FilterInput(ts.FileFrugal)

	tmpFile, err := ioutil.TempFile(`.`, `TestApply`)
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	filename := tmpFile.Name()

	// swallows errors when the file has bad yaml in it
	err = ioutil.WriteFile(filename, []byte(`bad yaml`), 0)
	require.NoError(t, err)
	bg.Options[`filter_yaml`] = filename
	bg.FilterInput(ts.FileFrugal)

	testYaml := `included:
  structs:
    names:
      - Burrito
  services:
    specs:
      - name: MusicService
        methods:
          - getArtist
excluded:
  services:
    all: true
  structs:
    all: true
  scopes:
    all: true`
	err = ioutil.WriteFile(filename, []byte(testYaml), 0)
	require.NoError(t, err)

	// Ok, now let's apply a filter
	assert.NotEmpty(t, ts.FileFrugal.Scopes)
	assert.Len(t, ts.MusicService.Methods, 3)
	assert.Len(t, ts.FileFrugal.Structs, 6)
	assert.Len(t, ts.FileFrugal.Exceptions, 1)

	bg.FilterInput(ts.FileFrugal)

	assert.Empty(t, ts.FileFrugal.Scopes)
	assert.Len(t, ts.MusicService.Methods, 1)
	assert.Len(t, ts.FileFrugal.Structs, 2)
	assert.Len(t, ts.FileFrugal.Exceptions, 1)
}
