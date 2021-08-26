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

	for i := 0; i < len(f.Services); i++ {
		service := f.Services[i]
		if shouldEntirelyRemoveService(spec, service) {
			f.Services = append(f.Services[:i], f.Services[i+1:]...)
			i--
			continue
		}

		applyFilterToService(spec, service)
	}

	for i := 0; i < len(f.Scopes); i++ {
		if shouldEntirelyRemoveScope(spec, f.Scopes[i]) {
			f.Scopes = append(f.Scopes[:i], f.Scopes[i+1:]...)
			i--
			continue
		}
	}

	// TODO do the same for structs...

	return nil
}

func shouldEntirelyRemoveService(
	gf *frugalFilterYaml,
	s *parser.Service,
) bool {
	return gf.Excluded.isEntireService(s)
}

func shouldEntirelyRemoveScope(
	gf *frugalFilterYaml,
	s *parser.Scope,
) bool {
	return gf.Excluded.shouldRemoveScope(s)
}
