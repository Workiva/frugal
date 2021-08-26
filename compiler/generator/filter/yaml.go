package filter

import (
	"io/ioutil"

	"github.com/Workiva/frugal/compiler/parser"
	"gopkg.in/yaml.v2"
)

type filterSpec struct {
	Included *definitionsSpec `yaml:"included"`
	Excluded *definitionsSpec `yaml:"excluded"`
}

func newYamlSpec(filename string) (*filterSpec, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	gf := &filterSpec{}

	err = yaml.Unmarshal(input, gf)
	if err != nil {
		return nil, err
	}

	return gf, nil
}

type definitionsSpec struct {
	Services  []serviceSpec  `yaml:"services"`
	Structs   *structSpec    `yaml:"structs"`
	Scopes    *scopesSpec    `yaml:"scopes"`
	Constants *constantsSpec `yaml:"constants"`
	Typedefs  *typedefsSpec  `yaml:"types"`
}

func (ffs *definitionsSpec) isServiceSpecified(
	s *parser.Service,
) bool {
	if ffs == nil {
		return false
	}

	for _, fs := range ffs.Services {
		if fs.matches(s) {
			return true
		}
	}
	return false
}

func (ffs *definitionsSpec) isEntireServiceSpecified(
	s *parser.Service,
) bool {
	if ffs == nil {
		return false
	}

	for _, fs := range ffs.Services {
		if fs.matches(s) && fs.Entire != nil {
			return *fs.Entire
		}
	}

	return false
}

func (ffs *definitionsSpec) isEntireScopeSpecified(
	s *parser.Scope,
) bool {
	if ffs == nil || ffs.Scopes == nil {
		return false
	}

	// Currently, we don't have the ability to filter at a per-scope level.
	// It's all or nothing.
	return ffs.Scopes.All != nil && *ffs.Scopes.All
}

func (ffs *definitionsSpec) isStructSpecified(
	s *parser.Struct,
) bool {
	if ffs == nil {
		return false
	}

	return ffs.Structs.isStructSpecified(s)
}

func (ffs *definitionsSpec) isConstantSpecified(
	s *parser.Constant,
) bool {
	if ffs == nil {
		return false
	}

	return ffs.Constants.isSpecified(s)
}

func (ffs *definitionsSpec) isTypedefSpecified(
	s *parser.TypeDef,
) bool {
	if ffs == nil {
		return false
	}

	return ffs.Typedefs.isSpecified(s)
}
