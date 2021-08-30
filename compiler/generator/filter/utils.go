package filter

import (
	"github.com/Workiva/frugal/compiler/parser"
)

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
