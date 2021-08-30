// +build !prod

package filter

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Workiva/frugal/compiler/parser"
)

const (
	albumIDL = `struct Artist {
	0: string Name
	1: int64 BirthYear
}
struct Song {
	0: string Title
	1: int64 DurationMS
}
struct Year {
	0: int64 AD
}
struct Place {
	0: int64 Position
	1: int32 DurationWeeks
}
struct Album {
	0: Artist Artist
	1: list<Song> Songs
	2: map<Year, Place> BillboardRanks
	3: string title
}
exception ArtistError {
	0: string message
}
service MusicService {
    Artist getArtist(1: string name) throws (1: ArtistError aErr)

    Album getAlbum(1: string title)

    list<Album> getTop5Albums(1: Year year)
}
struct Burrito {
	0: string meat
	1: int64 weightGrams
}
scope NewReleases prefix global.releases {
	Album: albums
	Song: songs
}`
)

func GetTestStructs(t *testing.T) testingIDL {
	return getTestStructs(t)
}

type testingIDL struct {
	FileFrugal *parser.Frugal

	artist, song, year, place, album *parser.Struct

	artistError *parser.Struct

	MusicService *parser.Service

	getArtist, getAlbum, getTop5Albums *parser.Method

	burrito *parser.Struct
}

func getTestStructs(t *testing.T) testingIDL {
	tmpFile, err := ioutil.TempFile(`.`, `getTestStructs`)
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	filename := tmpFile.Name()
	err = ioutil.WriteFile(filename, []byte(albumIDL), 0)
	require.NoError(t, err)

	i, err := parser.ParseFile(filename)
	require.NoError(t, err)
	require.NotNil(t, i)
	require.IsType(t, &parser.Frugal{}, i)

	res := testingIDL{
		FileFrugal: i.(*parser.Frugal),
	}

	require.NotEmpty(t, res.FileFrugal.Exceptions)
	for i := range res.FileFrugal.Exceptions {
		e := res.FileFrugal.Exceptions[i]
		switch e.Name {
		case `ArtistError`:
			res.artistError = e
		}
	}

	require.NotEmpty(t, res.FileFrugal.Structs)
	for i := range res.FileFrugal.Structs {
		s := res.FileFrugal.Structs[i]
		switch s.Name {
		case `Artist`:
			res.artist = s
		case `Song`:
			res.song = s
		case `Year`:
			res.year = s
		case `Place`:
			res.place = s
		case `Album`:
			res.album = s
		case `Burrito`:
			res.burrito = s
		}
	}
	for i := range res.FileFrugal.Services {
		s := res.FileFrugal.Services[i]
		switch s.Name {
		case `MusicService`:
			res.MusicService = s
		}
	}

	require.NotNil(t, res.artist)
	require.NotNil(t, res.song)
	require.NotNil(t, res.year)
	require.NotNil(t, res.place)
	require.NotNil(t, res.album)
	require.NotNil(t, res.artistError)
	require.NotNil(t, res.MusicService)

	for i := range res.MusicService.Methods {
		m := res.MusicService.Methods[i]
		switch m.Name {
		case `getArtist`:
			res.getArtist = m
		case `getAlbum`:
			res.getAlbum = m
		case `getTop5Albums`:
			res.getTop5Albums = m
		}
	}
	require.NotNil(t, res.getArtist)
	require.NotNil(t, res.getAlbum)
	require.NotNil(t, res.getTop5Albums)
	require.NotNil(t, res.burrito)

	return res
}
