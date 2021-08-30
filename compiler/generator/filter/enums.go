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
			isEnumNeededByStructs(e, f.Structs) ||
			isEnumNeededByStructs(e, f.Exceptions) ||
			isEnumNeededByStructs(e, f.Unions) ||
			isEnumNeededByAnyService(e, f.Services) {
			needed = append(needed, e)
		}
	}

	return needed
}

func isEnumNeededByStructs(
	e *parser.Enum,
	ss []*parser.Struct,
) bool {
	for _, s := range ss {
		if anyFieldContainsTypeWithName(s.Fields, e.Name) {
			return true
		}
	}
	return false
}

func isEnumNeededByAnyService(
	e *parser.Enum,
	ss []*parser.Service,
) bool {
	for _, s := range ss {
		if isEnumNeededByService(e, s) {
			return true
		}
	}
	return false
}

func isEnumNeededByService(
	e *parser.Enum,
	s *parser.Service,
) bool {
	for _, m := range s.Methods {
		if isEnumNeededByMethod(e, m) {
			return true
		}
	}
	return false
}

func isEnumNeededByMethod(
	e *parser.Enum,
	m *parser.Method,
) bool {
	if m == nil {
		return false
	}

	if typeContainsTypeWithName(m.ReturnType, e.Name) {
		return true
	}
	for _, arg := range m.Arguments {
		if fieldContainsTypeWithName(arg, e.Name) {
			return true
		}
	}
	for _, exc := range m.Exceptions {
		if fieldContainsTypeWithName(exc, e.Name) {
			return true
		}
	}

	return false
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
