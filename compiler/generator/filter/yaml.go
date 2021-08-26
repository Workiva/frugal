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
	Services []serviceSpec `yaml:"services"`
	Structs  *structSpec   `yaml:"structs"`
	Scopes   *scopesSpec   `yaml:"scopes"`
}

func (ds *definitionsSpec) isServiceSpecified(s *parser.Service) bool {
	if ds == nil {
		return false
	}

	for _, ss := range ds.Services {
		if ss.matches(s) {
			return true
		}
	}

	return false
}

func (ds *definitionsSpec) isEntireServiceSpecified(s *parser.Service) bool {
	if ds == nil {
		return false
	}

	for _, ss := range ds.Services {
		if ss.isEntireServiceSpecified(s) {
			return true
		}
	}

	return false
}

func (ds *definitionsSpec) isEntireScopeSpecified(s *parser.Scope) bool {
	return ds != nil && ds.Scopes.isSpecified(s)
}

func (ds *definitionsSpec) isStructSpecified(s *parser.Struct) bool {
	return ds != nil && ds.Structs.isStructSpecified(s)
}
