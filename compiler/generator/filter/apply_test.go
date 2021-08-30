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

	assert.NotEmpty(t, ts.FileFrugal.Scopes)
	applyToScopes(fs, ts.FileFrugal)
	assert.Empty(t, ts.FileFrugal.Scopes)
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
	ts.FileFrugal.Services = nil

	assert.Len(t, ts.FileFrugal.Structs, 6)
	assert.NotEmpty(t, ts.FileFrugal.Exceptions)
	applyToStructs(fs, ts.FileFrugal)
	assert.Len(t, ts.FileFrugal.Structs, 5)
	assert.Empty(t, ts.FileFrugal.Exceptions)
}

func TestApplyToServices(t *testing.T) {
	ts := getTestStructs(t)

	testYaml := `excluded:
  services:
    specs:
      - name: MusicService
        methods:
          - getAlbum
          - getTop5Albums`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)

	assert.Len(t, ts.MusicService.Methods, 3)
	applyToServices(fs, ts.FileFrugal)
	assert.Len(t, ts.MusicService.Methods, 1)
}

func TestApply(t *testing.T) {
	ts := getTestStructs(t)

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
	tmpFile, err := ioutil.TempFile(`.`, `TestApply`)
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	filename := tmpFile.Name()
	err = ioutil.WriteFile(filename, []byte(testYaml), 0)
	require.NoError(t, err)

	assert.NotEmpty(t, ts.FileFrugal.Scopes)
	assert.Len(t, ts.MusicService.Methods, 3)
	assert.Len(t, ts.FileFrugal.Structs, 6)
	assert.Len(t, ts.FileFrugal.Exceptions, 1)

	err = Apply(filename, ts.FileFrugal)
	require.NoError(t, err)

	assert.Empty(t, ts.FileFrugal.Scopes)
	assert.Len(t, ts.MusicService.Methods, 1)
	assert.Len(t, ts.FileFrugal.Structs, 2)
	assert.Len(t, ts.FileFrugal.Exceptions, 1)

}
