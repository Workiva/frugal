package filter

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type structSpec struct {
	All   *bool    `yaml:"all"`
	Names []string `yaml:"names"`
}

func (ss *structSpec) isStructSpecified(
	s *parser.Struct,
) bool {
	if ss == nil {
		return false
	}

	if ss.All != nil && *ss.All {
		return true
	}

	for _, name := range ss.Names {
		if strings.EqualFold(name, s.Name) {
			return true
		}
	}

	return false
}

// getNeededStructs will return the known struct-ish types that will be needed.
//
// It includes Structs, Exceptions, and Unions (struct-ish).
// This will not be able to pick up types that are used across
// frugal generated repos appropriately.
func getNeededStructs(
	spec *filterSpec,
	f *parser.Frugal,
) []*parser.Struct {

	// Structs, Exceptions, and Unions can all inherit from each other.
	// i.e. If a union is a union of structs, then we'll need to
	// parse through all three types to notice that.
	allParserStructs := make([]*parser.Struct, 0, len(f.Structs))
	allParserStructs = append(allParserStructs, f.Structs...)
	allParserStructs = append(allParserStructs, f.Exceptions...)
	allParserStructs = append(allParserStructs, f.Unions...)

	subset := make([]*parser.Struct, 0, len(allParserStructs)/2)
	notAdded := make([]*parser.Struct, 0, len(allParserStructs)/2)

	for _, s := range allParserStructs {
		if spec.Included.isStructSpecified(s) {
			debugPrintf("Including struct %q\n\tSpecified in Include\n", s.Name)
			// this struct is needed because the caller specified it
			// specifically.
			subset = append(subset, s)
		} else if isStructUsedByAnyService(s, f.Services) {
			debugPrintf("Including struct %q\n\tUsed by a Service\n", s.Name)
			// it's needed because a Service needs it.
			subset = append(subset, s)
		} else {
			debugPrintf("Initially skipping struct %q\n\tNot directly specified in struct input or service method input\n", s.Name)
			notAdded = append(notAdded, s)
		}
	}

	return getAllSubstructs(subset, notAdded)
}

func isStructUsedByAnyService(
	s *parser.Struct,
	services []*parser.Service,
) bool {
	for _, service := range services {
		if isStructUsedByService(s, service) {
			// this struct is needed because a service wants it.
			return true
		}
	}
	return false
}

func isStructUsedByService(
	s *parser.Struct,
	service *parser.Service,
) bool {
	if s == nil || service == nil {
		return false
	}

	for _, m := range service.Methods {
		if structExistsInType(s, m.ReturnType) {
			return true
		}
		for _, arg := range m.Arguments {
			if structExistsInField(s, arg) {
				return true
			}
		}
		for _, exc := range m.Exceptions {
			if structExistsInField(s, exc) {
				return true
			}
		}
	}

	return false
}

// getAllSubstructs iterates through the types of the subset
// to determine which of the `notInSubset` Structs also need
// to be pulled in. It returns the `subset` + any needed from
// the `notInSubset`.
//
// Note: the input slices may be mutated during the call of this func.
func getAllSubstructs(
	subset, notInSubset []*parser.Struct,
) []*parser.Struct {
	toCheck := make([]*parser.Struct, 0, len(subset)+len(notInSubset))
	toCheck = append(toCheck, subset...)

	for len(toCheck) > 0 {
		s := toCheck[0]
		toCheck = toCheck[1:]

		// look at all of the structs that aren't in the "known to include" subset.
		// If we find any that are used by `toCheck`, then add them to the subset,
		// and make sure we iterate down into them by also adding them to `toCheck`.
		for i := 0; len(notInSubset) > 0 && i < len(notInSubset); i++ {
			other := notInSubset[i]
			if structExistsInAnyField(other, s.Fields) {
				debugPrintf("Now including struct %q\n\tUsed in field (or sub-field) of %q\n", other.Name, s.Name)
				subset = append(subset, other)
				toCheck = append(toCheck, other)
				notInSubset = append(notInSubset[:i], notInSubset[i+1:]...)
				i--
			}
		}
	}
	return subset
}

func structExistsInAnyField(
	s *parser.Struct,
	fields []*parser.Field,
) bool {
	for _, f := range fields {
		if structExistsInField(s, f) {
			return true
		}
	}
	return false
}

// structExistsInField returns true if the struct appears in the given field.
// The field may have sub-types that use the given struct. If so, this still
// returns true.
func structExistsInField(
	s *parser.Struct,
	field *parser.Field,
) bool {
	return structExistsInType(s, field.Type)
}

func structExistsInType(
	s *parser.Struct,
	typ *parser.Type,
) bool {
	if typ == nil {
		return false
	}

	if s.Name == typ.Name {
		return true
	}

	// Check slices and maps by checking KeyType and ValueType
	if structExistsInType(s, typ.KeyType) {
		return true
	}

	if structExistsInType(s, typ.ValueType) {
		return true
	}

	return false
}
