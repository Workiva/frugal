package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Workiva/frugal/compiler/parser"
)

func TestShouldOmitGeneratedServiceServer(t *testing.T) {
	g := NewGenerator(nil).(*Generator)

	assert.False(t, g.shouldOmitGeneratedServiceServer())
	assert.NotEmpty(t, g.generateServer(&parser.Service{}))

	g = NewGenerator(map[string]string{
		shouldOmitServerGen: ``,
	}).(*Generator)

	assert.True(t, g.shouldOmitGeneratedServiceServer())
	assert.Empty(t, g.generateServer(nil))
}
