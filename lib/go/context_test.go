package frugal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrelationID(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, corid, ctx.CorrelationID())
}

func TestOpID(t *testing.T) {
	corid := "fooid"
	opid := "12345"
	ctx := NewFContext(corid)
	ctx.requestHeaders[opID] = opid
	assert.Equal(t, uint64(12345), ctx.OpID())
}

func TestRequestHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	ctx.AddRequestHeader("foo", "bar")
	val, ok := ctx.RequestHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.RequestHeaders()["foo"])
	assert.Equal(t, corid, ctx.RequestHeaders()[cid])
	assert.NotEqual(t, "", ctx.RequestHeaders()[opID])
}

func TestResponseHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	ctx.AddResponseHeader("foo", "bar")
	val, ok := ctx.ResponseHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.ResponseHeaders()["foo"])
	assert.Equal(t, "", ctx.ResponseHeaders()[cid])
	assert.Equal(t, "", ctx.ResponseHeaders()[opID])
}
