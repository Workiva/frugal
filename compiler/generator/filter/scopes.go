package filter

import (
	"github.com/Workiva/frugal/compiler/parser"
)

type scopesSpec struct {
	All *bool `yaml:"all"`
}

func (ss *scopesSpec) isSpecified(
	s *parser.Scope,
) bool {
	if ss == nil {
		return false
	}

	// Currently, we don't have the ability to filter at a per-scope level.
	// It's all or nothing.
	return ss.All != nil && *ss.All
}
