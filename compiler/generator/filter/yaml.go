package filter

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/Workiva/frugal/compiler/parser"
)

// filterSpec is the top-level definition of what should be filtered
type filterSpec struct {
	Included *definitionsSpec `yaml:"included"`
	Excluded *definitionsSpec `yaml:"excluded"`
}

func newYamlSpec(filename string) (*filterSpec, error) {
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fs := &filterSpec{}

	err = yaml.Unmarshal(input, fs)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// definitionsSpec is all of the Services, Scopes, and Structs that should be included/excluded.
// Eventually this may include constants, typedefs, and others.
type definitionsSpec struct {
<<<<<<< HEAD
	Services  []serviceSpec  `yaml:"services"`
	Structs   *structSpec    `yaml:"structs"`
	Scopes    *scopesSpec    `yaml:"scopes"`
	Constants *constantsSpec `yaml:"constants"`
	Typedefs  *typedefsSpec  `yaml:"types"`
=======
	Services *servicesSpec `yaml:"services"`
	Structs  *structSpec   `yaml:"structs"`
	Scopes   *scopesSpec   `yaml:"scopes"`
>>>>>>> 88cdd501d986d105e84d7c527c2bf5427e12233f
}

func (ds *definitionsSpec) isServiceSpecified(s *parser.Service) bool {
	return ds != nil && ds.Services.isServiceSpecified(s)
}

func (ds *definitionsSpec) isEntireServiceSpecified(s *parser.Service) bool {
	return ds != nil && ds.Services.isEntireServiceSpecified(s)
}

func (ds *definitionsSpec) isEntireScopeSpecified(s *parser.Scope) bool {
	return ds != nil && ds.Scopes.isSpecified(s)
}

func (ds *definitionsSpec) isStructSpecified(s *parser.Struct) bool {
	return ds != nil && ds.Structs.isStructSpecified(s)
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
