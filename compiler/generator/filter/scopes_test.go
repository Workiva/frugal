package filter

import (
	"testing"

	"github.com/Workiva/frugal/compiler/parser"
	"github.com/stretchr/testify/assert"
)

func TestScopesSpecIsSpecified(t *testing.T) {
	s := &parser.Scope{Name: `updatesScope`}
	var ss *scopesSpec
	assert.False(t, ss.isSpecified(s))

	ss = &scopesSpec{}
	assert.False(t, ss.isSpecified(s))

	b := false
	ss.All = &b
	assert.False(t, ss.isSpecified(s))

	b = true
	assert.True(t, ss.isSpecified(s))
}
