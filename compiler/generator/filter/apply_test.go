package filter

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestApplyToScopes(t *testing.T) {
	ts := getTestStructs(t)

	testYaml := `excluded:
  scopes:
    all: true`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)

	assert.NotEmpty(t, ts.fileFrugal.Scopes)
	applyToScopes(fs, ts.fileFrugal)
	assert.Empty(t, ts.fileFrugal.Scopes)
}

func TestApplyToStructs(t *testing.T) {
	ts := getTestStructs(t)

	testYaml := `excluded:
  structs:
    names:
      - Burrito
      - ArtistError`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)

	// we don't want the services to keep around structs in this unit test.
	ts.fileFrugal.Services = nil

	assert.Len(t, ts.fileFrugal.Structs, 6)
	assert.NotEmpty(t, ts.fileFrugal.Exceptions)
	applyToStructs(fs, ts.fileFrugal)
	assert.Len(t, ts.fileFrugal.Structs, 5)
	assert.Empty(t, ts.fileFrugal.Exceptions)
}

func TestApplyToServices(t *testing.T) {
	ts := getTestStructs(t)

	testYaml := `excluded:
  services:
    specs:
      - name: musicService
        methods:
          - getAlbum
          - getTop5Albums`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)

	assert.Len(t, ts.musicService.Methods, 3)
	applyToServices(fs, ts.fileFrugal)
	assert.Len(t, ts.musicService.Methods, 1)
}

func TestApply(t *testing.T) {
	ts := getTestStructs(t)

	testYaml := `included:
  structs:
    names:
      - Burrito
  services:
    specs:
      - name: musicService
        methods:
          - getArtist
excluded:
  services:
    all: true
  structs:
    all: true
  scopes:
    all: true`
	tmpFile, err := ioutil.TempFile(`.`, `TestApply`)
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	filename := tmpFile.Name()
	err = ioutil.WriteFile(filename, []byte(testYaml), 0)
	require.NoError(t, err)

	assert.NotEmpty(t, ts.fileFrugal.Scopes)
	assert.Len(t, ts.musicService.Methods, 3)
	assert.Len(t, ts.fileFrugal.Structs, 6)
	assert.Len(t, ts.fileFrugal.Exceptions, 1)

	err = Apply(filename, ts.fileFrugal)
	require.NoError(t, err)

	assert.Empty(t, ts.fileFrugal.Scopes)
	assert.Len(t, ts.musicService.Methods, 1)
	assert.Len(t, ts.fileFrugal.Structs, 2)
	assert.Len(t, ts.fileFrugal.Exceptions, 1)

}
