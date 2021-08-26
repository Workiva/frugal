package filter

import (
	"github.com/Workiva/frugal/compiler/parser"
)

func Apply(
	filename string,
	f *parser.Frugal,
) error {
	spec, err := newYamlSpec(filename)
	if err != nil {
		return err
	}

	applyToServices(spec, f)
	applyToScopes(spec, f)
	applyToStructs(spec, f)
	applyToTypeDefs(spec, f)
	applyToConstants(spec, f)

	// TODO consider Enums, Namespaces, Includes?

	return nil
}

func applyToServices(
	spec *filterSpec,
	f *parser.Frugal,
) {
	for i := 0; i < len(f.Services); i++ {
		service := f.Services[i]
		if spec.Excluded.isEntireServiceSpecified(service) {
			f.Services = append(f.Services[:i], f.Services[i+1:]...)
			i--
			continue
		}

		applyFilterToService(spec, service)
	}
}

func applyToScopes(
	spec *filterSpec,
	f *parser.Frugal,
) {
	for i := 0; i < len(f.Scopes); i++ {
		if spec.Excluded.isEntireScopeSpecified(f.Scopes[i]) {
			f.Scopes = append(f.Scopes[:i], f.Scopes[i+1:]...)
			i--
			continue
		}
	}
}

func applyToStructs(
	spec *filterSpec,
	f *parser.Frugal,
) {
	if spec.Excluded.Structs == nil ||
		(spec.Excluded.Structs.All != nil &&
			*spec.Excluded.Structs.All != true) {
		// we have nothing to do if we're not specified in the excludes or if we
		// aren't excluding all
		return
	}

	requiredStructs := getNeededStructs(spec, f)

	for i := 0; i < len(f.Structs); i++ {
		s := f.Structs[i]

		if structListContains(requiredStructs, s) {
			continue
		}

		if spec.Excluded.isStructSpecified(s) {
			f.Structs = append(f.Structs[:i], f.Structs[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(f.Exceptions); i++ {
		s := f.Exceptions[i]

		if structListContains(requiredStructs, s) {
			continue
		}

		if spec.Excluded.isStructSpecified(s) {
			f.Exceptions = append(f.Exceptions[:i], f.Exceptions[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(f.Unions); i++ {
		s := f.Unions[i]

		if structListContains(requiredStructs, s) {
			continue
		}

		if spec.Excluded.isStructSpecified(s) {
			f.Unions = append(f.Unions[:i], f.Unions[i+1:]...)
			i--
		}
	}
}

func applyToConstants(
	spec *filterSpec,
	f *parser.Frugal,
) {
	if spec.Excluded.Constants == nil ||
		(spec.Excluded.Constants.All != nil &&
			*spec.Excluded.Constants.All != true) {
		// we have nothing to do if we're not specified in the excludes or if we
		// aren't excluding all
		return
	}

	// requiredStructs := getNeededStructs(spec, f)

	for i := 0; i < len(f.Constants); i++ {
		s := f.Constants[i]

		// if structListContains(requiredStructs, s) {
		// 	continue
		// }

		if spec.Excluded.isConstantSpecified(s) {
			f.Constants = append(f.Constants[:i], f.Constants[i+1:]...)
			i--
		}
	}
}

func applyToTypeDefs(
	spec *filterSpec,
	f *parser.Frugal,
) {
	if spec.Excluded.Typedefs == nil ||
		(spec.Excluded.Typedefs.All != nil &&
			*spec.Excluded.Typedefs.All != true) {
		// we have nothing to do if we're not specified in the excludes or if we
		// aren't excluding all
		return
	}

	// requiredStructs := getNeededStructs(spec, f)

	for i := 0; i < len(f.Typedefs); i++ {
		s := f.Typedefs[i]

		// if structListContains(requiredStructs, s) {
		// 	continue
		// }

		if spec.Excluded.isTypedefSpecified(s) {
			f.Typedefs = append(f.Typedefs[:i], f.Typedefs[i+1:]...)
			i--
		}
	}
}
