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

// AFAICT this will not be able to pick up types that are used across
// frugal generated repos appropriately...
func getNeededStructs(
	spec *filterSpec,
	f *parser.Frugal,
) []*parser.Struct {

	allStructIshs := make([]*parser.Struct, 0, len(f.Structs))
	allStructIshs = append(allStructIshs, f.Structs...)
	allStructIshs = append(allStructIshs, f.Exceptions...)
	allStructIshs = append(allStructIshs, f.Unions...)

	subset := make([]*parser.Struct, 0, len(allStructIshs)/2)
	notAdded := make([]*parser.Struct, 0, len(allStructIshs)/2)

	for _, s := range allStructIshs {
		if spec.Included.isStructSpecified(s) {
			// this struct is needed because the caller specified it
			// specifically.
			subset = append(subset, s)
		} else if isStructUsedByAnyService(s, f.Services) {
			// it's needed because a Service needs it.
			subset = append(subset, s)
		} else {
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
		if isUsedByService(s, service) {
			// this struct is needed because a service wants it.
			return true
		}
	}
	return false
}

func isUsedByService(
	s *parser.Struct,
	service *parser.Service,
) bool {
	if s == nil || service == nil {
		return false
	}

	for _, m := range service.Methods {
		if m.ReturnType != nil {
			if s.Name == m.ReturnType.Name {
				return true
			}
		}
		for _, arg := range m.Arguments {
			if s.Name == arg.Type.Name {
				return true
			}
		}
		for _, exc := range m.Exceptions {
			if s.Name == exc.Type.Name {
				return true
			}
		}
	}

	return false
}

// the subset and notInSubset slices may be mutated during the call of this func.
func getAllSubstructs(
	subset, notInSubset []*parser.Struct,
) []*parser.Struct {
	toCheck := make([]*parser.Struct, 0, len(subset)+len(notInSubset))
	toCheck = append(toCheck, subset...)

	for len(toCheck) > 0 {
		s := toCheck[0]
		toCheck = toCheck[1:]

		for i := 0; len(notInSubset) > 0 && i < len(notInSubset); i++ {
			other := notInSubset[i]
			for _, f := range s.Fields {
				if structContainsType(other, f) {
					subset = append(subset, other)
					toCheck = append(toCheck, other)
					notInSubset = append(notInSubset[:i], notInSubset[i+1:]...)
					i--
					break
				}
			}
		}
	}
	return subset
}

func structContainsType(
	s *parser.Struct,
	field *parser.Field,
) bool {
	return typeContainsType(s, field.Type)
}

func typeContainsType(
	s *parser.Struct,
	typ *parser.Type,
) bool {
	if typ == nil {
		return false
	}

	if s.Name == typ.Name {
		return true
	}

	// Check slices and maps:
	if typeContainsType(s, typ.KeyType) {
		return true
	}

	if typeContainsType(s, typ.ValueType) {
		return true
	}

	return false
}

func structListContains(
	ss []*parser.Struct,
	s *parser.Struct,
) bool {
	for _, other := range ss {
		if s.Name == other.Name {
			return true
		}
	}
	return false
}
