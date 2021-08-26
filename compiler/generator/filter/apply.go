package filter

import "github.com/Workiva/frugal/compiler/parser"

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

	// TODO do the same for structs...

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
