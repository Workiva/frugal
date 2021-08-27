package filter

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type servicesSpec struct {
	All   *bool         `yaml:"all"`
	Specs []serviceSpec `yaml:"specs"`
}

func (ss *servicesSpec) isEntireServiceSpecified(s *parser.Service) bool {
	if ss == nil {
		return false
	}

	if ss.All != nil && *ss.All {
		return true
	}

	for _, ss := range ss.Specs {
		if ss.isEntireServiceSpecified(s) {
			return true
		}
	}

	return false
}

func (ss *servicesSpec) isServiceSpecified(s *parser.Service) bool {
	if ss == nil {
		return false
	}

	if ss.All != nil && *ss.All {
		return true
	}

	for _, ss := range ss.Specs {
		if ss.matches(s) {
			return true
		}
	}

	return false
}

type serviceSpec struct {
	Name    string   `yaml:"name"`
	Entire  *bool    `yaml:"all"`
	Methods []string `yaml:"methods"`
}

func (ss *serviceSpec) isEntireServiceSpecified(s *parser.Service) bool {
	return ss.matches(s) && ss.Entire != nil && *ss.Entire
}

func (ss *serviceSpec) matches(s *parser.Service) bool {
	return strings.EqualFold(s.Name, ss.Name)
}

func (ss *serviceSpec) isMethodSpecified(
	m *parser.Method,
) bool {
	for _, fm := range ss.Methods {
		if strings.EqualFold(m.Name, fm) {
			return true
		}
	}
	return false
}

func applyFilterToService(
	fs *filterSpec,
	s *parser.Service,
) {

	isIncludesSpecified := fs.Included.isServiceSpecified(s)
	isExcludesSpecified := fs.Excluded.isServiceSpecified(s)
	if !isIncludesSpecified && !isExcludesSpecified {
		// nothing to do!
		return
	}

	msc := s.Methods

	if isIncludesSpecified {
		// reset the msc so that we only start including the desired "includes"
		msc = msc[:0]
		for _, sf := range fs.Included.Services.Specs {
			if !sf.matches(s) {
				continue
			}

			for _, m := range s.Methods {
				if !methodSliceIncludes(msc, m) && sf.isMethodSpecified(m) {
					// add methods if they have not already been included
					// and if the included spec has them
					msc = append(msc, m)
				}
			}
		}
	}

	if isExcludesSpecified {
		for _, sf := range fs.Excluded.Services.Specs {
			if !sf.matches(s) {
				continue
			}

			for i := 0; i < len(msc); i++ {
				if sf.isMethodSpecified(msc[i]) {
					// add methods if they have not already been included
					// and if the included spec has them
					msc = append(msc[:i], msc[i+1:]...)
					i--
					continue
				}
			}
		}
	}

	s.Methods = msc
}

func methodSliceIncludes(
	ms []*parser.Method,
	other *parser.Method,
) bool {
	for _, m := range ms {
		if m.Name == other.Name {
			return true
		}
	}
	return false
}
