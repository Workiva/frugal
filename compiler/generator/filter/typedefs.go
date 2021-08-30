package filter

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type typedefsSpec struct {
	All   *bool    `yaml:"all"`
	Names []string `yaml:"names"`
}

func (ts *typedefsSpec) isSpecified(t *parser.TypeDef) bool {
	if ts == nil {
		return false
	}

	if ts.All != nil && *ts.All {
		return true
	}

	for _, name := range ts.Names {
		if strings.EqualFold(name, t.Name) {
			return true
		}
	}

	return false
}

func getNeededTypeDefs(
	spec *filterSpec,
	f *parser.Frugal,
) []*parser.TypeDef {
	needed := make([]*parser.TypeDef, 0, len(f.Typedefs))

	for i := range f.Typedefs {
		e := f.Typedefs[i]

		if spec.Included.isTypedefSpecified(e) ||
			isNameNeededByStructs(e.Name, f.Structs) ||
			isNameNeededByStructs(e.Name, f.Exceptions) ||
			isNameNeededByStructs(e.Name, f.Unions) ||
			isNameNeededByAnyService(e.Name, f.Services) {
			needed = append(needed, e)
		}
	}

	return needed
}

func typeDefListContains(
	es []*parser.TypeDef,
	e *parser.TypeDef,
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
