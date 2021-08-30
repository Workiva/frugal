package filter

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type enumsSpec struct {
	All   *bool    `yaml:"all"`
	Names []string `yaml:"names"`
}

func (ss *enumsSpec) isSpecified(
	e *parser.Enum,
) bool {
	if ss == nil {
		return false
	}

	if ss.All != nil && *ss.All {
		return true
	}

	for _, name := range ss.Names {
		if strings.EqualFold(name, e.Name) {
			return true
		}
	}

	return false
}

func getNeededEnums(
	spec *filterSpec,
	f *parser.Frugal,
) []*parser.Enum {
	needed := make([]*parser.Enum, 0, len(f.Enums))

	for i := range f.Enums {
		e := f.Enums[i]

		if spec.Included.isEnumSpecified(e) ||
			isNameNeededByStructs(e.Name, f.Structs) ||
			isNameNeededByStructs(e.Name, f.Exceptions) ||
			isNameNeededByStructs(e.Name, f.Unions) ||
			isNameNeededByAnyService(e.Name, f.Services) {
			needed = append(needed, e)
		}
	}

	return needed
}

func enumListContains(
	es []*parser.Enum,
	e *parser.Enum,
) bool {
	if e == nil {
		return false
	}
	for _, other := range es {
		if e.Name == other.Name {
			return true
		}
	}
	return false
}
