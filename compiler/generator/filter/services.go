package filter

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type serviceSpec struct {
	Name    string   `yaml:"name"`
	Entire  *bool    `yaml:"all"`
	Methods []string `yaml:"methods"`
}

func (ffs *serviceSpec) matches(s *parser.Service) bool {
	return strings.EqualFold(s.Name, ffs.Name)
}

func (ffs *serviceSpec) isMethodSpecified(
	m *parser.Method,
) bool {
	for _, fm := range ffs.Methods {
		if strings.EqualFold(m.Name, fm) {
			return true
		}
	}
	return false
}

func applyFilterToService(
	gf *filterSpec,
	s *parser.Service,
) {

	isIncludesSpecified := gf.Included.isServiceSpecified(s)
	isExcludesSpecified := gf.Excluded.isServiceSpecified(s)
	if !isIncludesSpecified && !isExcludesSpecified {
		// nothing to do!
		return
	}

	msc := s.Methods

	if isIncludesSpecified {
		msc = msc[:0]
		for _, sf := range gf.Included.Services {
			if sf.Name != s.Name {
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
		for _, sf := range gf.Excluded.Services {
			if sf.Name != s.Name {
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
