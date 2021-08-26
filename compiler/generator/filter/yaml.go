package filter

import (
	"io/ioutil"

	"github.com/Workiva/frugal/compiler/parser"
	"gopkg.in/yaml.v2"
)

type frugalFilterYaml struct {
	Included *frugalFilterSpec `yaml:"included"`
	Excluded *frugalFilterSpec `yaml:"excluded"`
}

func newYamlSpec(filename string) (*frugalFilterYaml, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	gf := &frugalFilterYaml{}

	err = yaml.Unmarshal(input, gf)
	if err != nil {
		return nil, err
	}

	return gf, nil
}

type frugalFilterSpec struct {
	Services []frugalFilterServiceSpec `yaml:"services"`
	Structs  []filterFrugalStruct      `yaml:"structs"`
	Scopes   *frugalFilterScopesSpec   `yaml:"scopes"`
}

func (ffs *frugalFilterSpec) isServiceSpecified(
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

func (ffs *frugalFilterSpec) isEntireService(
	s *parser.Service,
) bool {
	if ffs == nil {
		return false
	}

	for _, fs := range ffs.Services {
		if fs.isService(s) && fs.Entire != nil {
			return *fs.Entire
		}
	}

	return false
}

func (ffs *frugalFilterSpec) shouldRemoveScope(
	s *parser.Scope,
) bool {
	if ffs == nil || ffs.Scopes == nil {
		return false
	}

	// Currently, we don't have the ability to filter at a per-scope level.
	// It's all or nothing.
	return ffs.Scopes.All != nil && *ffs.Scopes.All
}
