package filter

import (
	"github.com/Workiva/frugal/compiler/parser"
)

func isNameNeededByStructs(
	name string,
	ss []*parser.Struct,
) bool {
	for _, s := range ss {
		if anyFieldContainsTypeWithName(s.Fields, name) {
			return true
		}
	}
	return false
}

func isNameNeededByAnyService(
	name string,
	ss []*parser.Service,
) bool {
	for _, s := range ss {
		if isNameNeededByService(name, s) {
			return true
		}
	}
	return false
}

func isNameNeededByService(
	name string,
	s *parser.Service,
) bool {
	for _, m := range s.Methods {
		if isNameNeededByMethod(name, m) {
			return true
		}
	}
	return false
}

func isNameNeededByMethod(
	name string,
	m *parser.Method,
) bool {
	if m == nil {
		return false
	}

	if typeContainsTypeWithName(m.ReturnType, name) {
		return true
	}
	for _, arg := range m.Arguments {
		if fieldContainsTypeWithName(arg, name) {
			return true
		}
	}
	for _, exc := range m.Exceptions {
		if fieldContainsTypeWithName(exc, name) {
			return true
		}
	}

	return false
}

func anyFieldContainsTypeWithName(
	fields []*parser.Field,
	name string,
) bool {
	for _, f := range fields {
		if f != nil && fieldContainsTypeWithName(f, name) {
			return true
		}
	}
	return false
}

// fieldContainsTypeWithName returns true if the struct appears in the given field.
// The field may have sub-types that use the given struct. If so, this still
// returns true.
func fieldContainsTypeWithName(
	field *parser.Field,
	name string,
) bool {
	return typeContainsTypeWithName(field.Type, name)
}

func typeContainsTypeWithName(
	typ *parser.Type,
	name string,
) bool {
	if typ == nil {
		return false
	}

	if name == typ.Name {
		return true
	}

	// Check slices and maps by checking KeyType and ValueType
	if typeContainsTypeWithName(typ.KeyType, name) {
		return true
	}

	if typeContainsTypeWithName(typ.ValueType, name) {
		return true
	}

	return false
}
