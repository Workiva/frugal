package frugal

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensures NewFContext creates an FContext with the given correlation id.
func TestCorrelationID(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, corid, ctx.CorrelationID())
}

// Ensures NewFContext creates an FContext and generates a correlation id if
// one is not supplied.
func TestNewCorrelationID(t *testing.T) {
	cid := "abc"
	oldCID := generateCorrelationID
	defer func() { generateCorrelationID = oldCID }()
	generateCorrelationID = func() string { return cid }

	ctx := NewFContext("")

	assert.Equal(t, cid, ctx.CorrelationID())
}

// Ensures the "_opid" request header for an FContext is returned for calls to
// getOpID.
func TestOpID(t *testing.T) {
	assert := assert.New(t)
	corid := "fooid"
	opid := "12345"
	ctx := NewFContext(corid)
	ctx.AddRequestHeader(opIDHeader, opid)
	actOpID, err := getOpID(ctx)
	assert.Nil(err)
	assert.Equal(uint64(12345), actOpID)

	delete(ctx.(*FContextImpl).requestHeaders, opIDHeader)
	_, err = getOpID(ctx)
	assert.Equal(fmt.Errorf("FContext does not have the required %s request header", opIDHeader), err)

	opIDStr := "-123"
	ctx.(*FContextImpl).requestHeaders[opIDHeader] = opIDStr
	_, err = getOpID(ctx)
	assert.Equal(fmt.Errorf("FContext has an opid that is not a non-negative integer: %s", opIDStr), err)
}

// Ensures the "_timeout" request header is correctly set and calls to Timeout
// return the correct Duration.
func TestTimeout(t *testing.T) {
	// Check default timeout (5 seconds).
	ctx := NewFContext("")
	timeoutStr, _ := ctx.RequestHeader(timeoutHeader)
	assert.Equal(t, "5000", timeoutStr)
	assert.Equal(t, defaultTimeout, ctx.Timeout())

	// Set timeout and check expected values.
	ctx.SetTimeout(10 * time.Second)
	timeoutStr, _ = ctx.RequestHeader(timeoutHeader)
	assert.Equal(t, "10000", timeoutStr)
	assert.Equal(t, 10*time.Second, ctx.Timeout())
}

// Ensures AddRequestHeader properly adds the key-value pair to the context
// RequestHeaders.
func TestRequestHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddRequestHeader("foo", "bar"))
	val, ok := ctx.RequestHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.RequestHeaders()["foo"])
	assert.Equal(t, corid, ctx.RequestHeaders()[cidHeader])
	assert.NotEqual(t, "", ctx.RequestHeaders()[opIDHeader])

	assert.Equal(t, ctx, ctx.AddRequestHeader(cidHeader, "baz"))
	assert.Equal(t, ctx, ctx.AddRequestHeader(opIDHeader, "123"))

	assert.Equal(t, "baz", ctx.CorrelationID())
	actOpID, err := getOpID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), actOpID)
}

// Ensures AddResponseHeader properly adds the key-value pair to the context
// ResponseHeaders.
func TestResponseHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddResponseHeader("foo", "bar"))
	val, ok := ctx.ResponseHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.ResponseHeaders()["foo"])
	assert.Equal(t, "", ctx.ResponseHeaders()[cidHeader])
	assert.Equal(t, "", ctx.ResponseHeaders()[opIDHeader])

	assert.Equal(t, ctx, ctx.AddResponseHeader(opIDHeader, "1"))
	assert.Equal(t, "1", ctx.ResponseHeaders()[opIDHeader])
}

// Ensures AddTag, Tag, and Tags work as expected.
func TestTags(t *testing.T) {
	assert := assert.New(t)
	ctx := NewFContext("")
	val, ok := ctx.Tag("foo")
	assert.Nil(val)
	assert.False(ok)
	assert.Empty(ctx.Tags())
	ctx.AddTag("foo", "bar").AddTag("baz", "qux")
	val, ok = ctx.Tag("foo")
	assert.Equal("bar", val)
	assert.True(ok)
	val, ok = ctx.Tag("baz")
	assert.Equal("qux", val)
	assert.True(ok)
	tags := ctx.Tags()
	assert.Len(tags, 2)
	assert.Equal("bar", tags["foo"])
	assert.Equal("qux", tags["baz"])
}
