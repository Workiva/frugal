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

	if spec == nil || (spec.Included == nil && spec.Excluded == nil) {
		// nothing to do!
		return nil
	}

	debugPrintf("\n\nStarting filter of generated frugal in %q from %q...\n\n", f.Name, f.File)
	defer debugPrintf("\nCompleted filter of generated frugal in %q from %q.\n\n", f.Name, f.File)

	applyToServices(spec, f)
	applyToScopes(spec, f)
	applyToStructs(spec, f)
	applyToConstants(spec, f)
	applyToEnums(spec, f)
	applyToTypeDefs(spec, f)

	// FUTURE: filter out Namespaces and Includes

	return nil
}

func applyToServices(
	spec *filterSpec,
	f *parser.Frugal,
) {
	for i := 0; i < len(f.Services); i++ {
		service := f.Services[i]
		// if we're trying to exclude the entire service, and we don't have
		// it explicitly included for any reason, let's remove it entirely.
		if spec.Excluded.isEntireServiceSpecified(service) &&
			!spec.Included.isServiceSpecified(service) {
			debugPrintf("Excluding entire service %q\n", service.Name)
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
		scope := f.Scopes[i]
		if spec.Excluded.isEntireScopeSpecified(scope) {
			debugPrintf("Excluding entire scope %q\n", scope.Name)
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
		debugPrintln(`No structs excluded.`)
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
			debugPrintf("Excluding Struct %q\n", s.Name)
			f.Structs = append(f.Structs[:i], f.Structs[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(f.Exceptions); i++ {
		e := f.Exceptions[i]

		if structListContains(requiredStructs, e) {
			continue
		}

		if spec.Excluded.isStructSpecified(e) {
			debugPrintf("Excluding Exception %q\n", e.Name)
			f.Exceptions = append(f.Exceptions[:i], f.Exceptions[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(f.Unions); i++ {
		u := f.Unions[i]

		if structListContains(requiredStructs, u) {
			continue
		}

		if spec.Excluded.isStructSpecified(u) {
			debugPrintf("Excluding Union %q\n", u.Name)
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

func applyToEnums(
	spec *filterSpec,
	f *parser.Frugal,
) {
	if spec.Excluded.Enums == nil ||
		(spec.Excluded.Enums.All != nil &&
			*spec.Excluded.Enums.All != true) {
		// we have nothing to do if we're not specified in the excludes or if we
		// aren't excluding all
		debugPrintf("No Enums excluded\n")
		return
	}

	requiredEnums := getNeededEnums(spec, f)

	for i := 0; i < len(f.Enums); i++ {
		e := f.Enums[i]

		if enumListContains(requiredEnums, e) {
			continue
		}

		if spec.Excluded.isEnumSpecified(e) {
			debugPrintf("Excluding Enum %q\n", e.Name)
			f.Enums = append(f.Enums[:i], f.Enums[i+1:]...)
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
		debugPrintf("No Typedefs excluded\n")
		return
	}

	requiredTypes := getNeededTypeDefs(spec, f)

	for i := 0; i < len(f.Typedefs); i++ {
		t := f.Typedefs[i]

		if typeDefListContains(requiredTypes, t) {
			continue
		}

		if spec.Excluded.isTypedefSpecified(t) {
			debugPrintf("Excluding Type %q\n", t.Name)
			f.Typedefs = append(f.Typedefs[:i], f.Typedefs[i+1:]...)
			i--
		}
	}
}
