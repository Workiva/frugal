package generator

import (
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

type generatorFilter struct {
	Included *filterFrugalSpec `yaml:"included"`
	Excluded *filterFrugalSpec `yaml:"excluded"`
}

type filterFrugalSpec struct {
	Services []filterFrugalService `yaml:"services"`
	Structs  []filterFrugalStruct  `yaml:"structs"`
}

func (ffs *filterFrugalSpec) isServiceSpecified(
	s *parser.Service,
) bool {
	if ffs == nil {
		return false
	}

	for _, fs := range ffs.Services {
		if fs.isService(s) {
			return true
		}
	}
	return false
}

type filterFrugalService struct {
	Name    string   `yaml:"name"`
	Methods []string `yaml:"methods"`
}

func (ffs *filterFrugalService) isService(s *parser.Service) bool {
	return strings.EqualFold(s.Name, ffs.Name)
}

func (ffs *filterFrugalService) hasMethod(
	m *parser.Method,
) bool {
	for _, fm := range ffs.Methods {
		if strings.EqualFold(m.Name, fm) {
			return true
		}
	}
	return false
}

type filterFrugalStruct struct {
	Name string `yaml:"name"`
}

func applyFilterToService(
	gf *generatorFilter,
	s *parser.Service,
) {

	isIncludesSpecified := gf.Included.isServiceSpecified(s)
	isExcludesSpecified := gf.Included.isServiceSpecified(s)
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
				if !methodSliceIncludes(msc, m) && sf.hasMethod(m) {
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
				if sf.hasMethod(msc[i]) {
					// add methods if they have not already been included
					// and if the included spec has them
					msc = append(msc[:i], msc[i+1:]...)
					i--
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
