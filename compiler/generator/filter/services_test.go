package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/Workiva/frugal/compiler/parser"
)

func TestServicesSpecIsEntireServiceSpecified(t *testing.T) {
	s := &parser.Service{Name: `albumService`}

	var ss *servicesSpec

	assert.False(t, ss.isEntireServiceSpecified(nil))

	ss = &servicesSpec{}
	assert.False(t, ss.isEntireServiceSpecified(s))

	b := false
	ss.All = &b
	b = true
	assert.True(t, ss.isEntireServiceSpecified(s))
	b = false
	assert.False(t, ss.isEntireServiceSpecified(s))

	ss.Specs = append(ss.Specs, serviceSpec{
		Name: `notalbumService`,
	})
	assert.False(t, ss.isEntireServiceSpecified(s))

	ss.Specs = append(ss.Specs, serviceSpec{
		Name: `albumService`,
	})
	assert.False(t, ss.isEntireServiceSpecified(s))

	b2 := true
	ss.Specs[1].Entire = &b2
	assert.True(t, ss.isEntireServiceSpecified(s))

	b2 = false
	assert.False(t, ss.isEntireServiceSpecified(s))
}

func TestServicesSpecIsServiceSpecified(t *testing.T) {
	s := &parser.Service{Name: `albumService`}
	var ss *servicesSpec

	assert.False(t, ss.isServiceSpecified(nil))

	ss = &servicesSpec{}
	assert.False(t, ss.isServiceSpecified(s))

	b := false
	ss.All = &b
	b = true
	assert.True(t, ss.isServiceSpecified(s))
	b = false
	assert.False(t, ss.isServiceSpecified(s))

	ss.Specs = append(ss.Specs, serviceSpec{
		Name: `notalbumService`,
	})
	assert.False(t, ss.isServiceSpecified(s))

	ss.Specs = append(ss.Specs, serviceSpec{
		Name: `albumService`,
	})
	assert.True(t, ss.isServiceSpecified(s))
}

func TestServiceSpecIsEntireServiceSpecified(t *testing.T) {
	s := &parser.Service{Name: `albumService`}
	ss := &serviceSpec{
		Name: `notalbumService`,
	}

	assert.False(t, ss.isEntireServiceSpecified(s))

	ss.Name = `albumService`
	assert.False(t, ss.isEntireServiceSpecified(s))

	b := false
	ss.Entire = &b
	assert.False(t, ss.isEntireServiceSpecified(s))

	b = true
	assert.True(t, ss.isEntireServiceSpecified(s))
}

func TestServicesSpecMatches(t *testing.T) {
	s := &parser.Service{Name: `albumService`}
	ss := &serviceSpec{
		Name: `notalbumService`,
	}

	assert.False(t, ss.matches(s))

	ss.Name = `albumService`
	assert.True(t, ss.matches(s))

	ss.Name = `aLbUmSeRvIcE`
	assert.True(t, ss.matches(s))
}

func TestServicesSpecIsMethodSpecified(t *testing.T) {
	ss := &serviceSpec{
		Name: `albumService`,
	}

	assert.False(t, ss.isMethodSpecified(nil))

	m := &parser.Method{
		Name: `getArtist`,
	}
	assert.False(t, ss.isMethodSpecified(m))

	ss.Methods = append(ss.Methods, `getAlbum`)
	assert.False(t, ss.isMethodSpecified(m))

	ss.Methods = append(ss.Methods, `getArtist`)
	assert.True(t, ss.isMethodSpecified(m))
}

func TestMethodSliceIncludes(t *testing.T) {
	assert.False(t, methodSliceIncludes(nil, nil))
	assert.False(t, methodSliceIncludes([]*parser.Method{{}}, nil))

	m1 := &parser.Method{
		Name: `getArtist`,
	}
	m2 := &parser.Method{
		Name: `getAlbum`,
	}
	assert.False(t, methodSliceIncludes([]*parser.Method{m1}, m2))
	assert.True(t, methodSliceIncludes([]*parser.Method{m1, m2}, m2))
}

func TestApplyFilterToService(t *testing.T) {
	// the yaml specifies that we want Burrito
	// but none of the other structs.
	testYaml := `included:
  services:
    specs:
      - name: MusicService
        methods:
          - getAlbum
excluded:
  services:
    all: true`
	fs := &filterSpec{}
	err := yaml.Unmarshal([]byte(testYaml), fs)
	require.NoError(t, err)
	ts := getTestStructs(t)

	assert.Len(t, ts.MusicService.Methods, 3)
	applyFilterToService(fs, ts.MusicService)
	// based on the yaml, we should only have the single method `getAlbum` now
	require.Len(t, ts.MusicService.Methods, 1)
	assert.Equal(t, `getAlbum`, ts.MusicService.Methods[0].Name)
}
