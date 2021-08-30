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
			debugPrintf("Including %q\n", s.Name)
			debugPrintln("\tSpecified in Include")
			// this struct is needed because the caller specified it
			// specifically.
			subset = append(subset, s)
			continue
		}
		if isStructUsedByAnyService(s, f.Services) {
			debugPrintf("Including %q\n", s.Name)
			debugPrintln("\tUsed by a Service")
			// it's needed because a Service needs it.
			subset = append(subset, s)
			continue
		}

		// It would be nice to add a check for any of the `*Frugal` structs that
		// include this `f`, but we currently do not have access to that while
		// running the frugal binary. We would need a reverse look-up of `ParsedIncludes`.
		// In the meantime, we require consumers to manually define all of the structs
		// their methods will need from other packages.

		debugPrintf("Skipping %q\n", s.Name)
		debugPrintln("\tIt is not directly specified in struct input or service method input")
		notAdded = append(notAdded, s)
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
		if m == nil {
			continue
		}

		if typeContainsStruct(m.ReturnType, s) {
			return true
		}
		for _, arg := range m.Arguments {
			if fieldContainsStruct(arg, s) {
				return true
			}
		}
		for _, exc := range m.Exceptions {
			if fieldContainsStruct(exc, s) {
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
			if anyFieldContainsStruct(s.Fields, other) {
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

func anyFieldContainsStruct(
	fields []*parser.Field,
	s *parser.Struct,
) bool {
	for _, f := range fields {
		if f != nil && fieldContainsStruct(f, s) {
			return true
		}
	}
	return false
}

// fieldContainsStruct returns true if the struct appears in the given field.
// The field may have sub-types that use the given struct. If so, this still
// returns true.
func fieldContainsStruct(
	field *parser.Field,
	s *parser.Struct,
) bool {
	return typeContainsStruct(field.Type, s)
}

func typeContainsStruct(
	typ *parser.Type,
	s *parser.Struct,
) bool {
	if typ == nil {
		return false
	}

	if s.Name == typ.Name {
		return true
	}

	// Check slices and maps by checking KeyType and ValueType
	if typeContainsStruct(typ.KeyType, s) {
		return true
	}

	if typeContainsStruct(typ.ValueType, s) {
		return true
	}

	return false
}

func structListContains(
	ss []*parser.Struct,
	s *parser.Struct,
) bool {
	if s == nil {
		return false
	}
	for _, other := range ss {
		if s.Name == other.Name && s.Type == other.Type {
			return true
		}
	}
	return false
}
