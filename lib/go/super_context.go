package frugal

import (
	"context"
	"time"
	"strings"
	"fmt"
)

const superFContextPrefix = "super-f-context-"
const superFContextDeadlinePrefix = superFContextPrefix + "deadline"
const superFContextErrorPrefix = superFContextPrefix + "error"

// Structs that implement this interface use inject and extract to propagate
// data from one struct to another. An example of using inject/extract to
// propagate data from one context to another:
// context_a.Inject(context_b.Extract())
// context_a can then continue to be propagated and pass anything new that was
// added to it with context_b.Inject(context_a.Extract())
type InjectExtractContext interface {
	// Inject stores the key value pairs in the map into the context.
	Inject(map[string]string)

	// Extract returns key value pairs that represent that context and can
	// be injected into another context to pass on certain properties.
	Extract() map[string]string
}

// TextMapCarriers have functions that allow for iterating over key, value
// string pairs and performing an action on them. It also allows for adding a
// new key value string pair.
type TextMapCarrier interface {
	// ForeachKey should iterate over key value string pairs and call the
	// function passed into it with each key value pair.
	ForeachKey(handler func(key, val string) error) error

	// Adds or updates a key value pair
	Set(string, string)
}

// SuperFContext is a hybrid of FContext and Golang's context. SuperFContext
// also implements the InjectExtractContext so that data can be passed along to
// other contexts, as well as the TextMapCarrier interface which is required
// for the propagation of span's in traces.
type SuperFContext struct {
	FContext
	context.Context
}

// Creates a new SuperFContext with a given a FContext
func NewSuperFContext(fContext FContext) SuperFContext {
	ctx := NewSuperFContextWithContext(fContext, context.Background())

	for key, val := range fContext.RequestHeaders() {
		if strings.HasPrefix(key, superFContextDeadlinePrefix) {
			if deadlineTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", val); err == nil {
				ctx, _ = WithDeadline(ctx, deadlineTime)
			} else {
				fmt.Println("not adding deadline")
			}
		} else if strings.HasPrefix(key, superFContextPrefix) {
			ctx.AddRequestHeader(string(key[len(superFContextPrefix)]), val)
		} else {
			ctx.AddRequestHeader(key, val)
		}
	}

	return ctx
}

// Creates a new SuperFContext with a given FContext and Golang context
func NewSuperFContextWithContext(fContext FContext, goContext context.Context) SuperFContext {
	return SuperFContext{fContext, goContext}
}

// Adds string key value pairs to the request headers
func (c *SuperFContext) Inject(contextMap map[string] string) {
	for key, val := range contextMap {
		if strings.HasPrefix(key, superFContextDeadlinePrefix) {
			if deadlineTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", val); err == nil {
				newCtx, _ := WithDeadline(*c, deadlineTime)
				c = &newCtx
				fmt.Println(c.Deadline())
			} else {
				fmt.Println("not adding deadline")
			}
		} else if strings.HasPrefix(key, superFContextPrefix) {
			c.AddRequestHeader(string(key[len(superFContextPrefix)]), val)
		} else {
			c.AddRequestHeader(key, val)
		}
	}
}

// Returns the key value pairs in the context's request headers.
func (c *SuperFContext) Extract() map[string] string {
	contextMap := make(map[string] string)

	for key, val := range c.RequestHeaders() {
		contextMap[key] = val;
	}

	if deadline, ok := c.Deadline(); ok {
		contextMap[superFContextDeadlinePrefix] = deadline.String()
	}

	//contextMap[superFContextPrefix] = string(c.Err().Error())

	return contextMap;
}

// Calls a function with each of the key value pairs in the request header
func (c SuperFContext) ForeachKey (handler func(key, val string) error) error {
	for k, v := range c.RequestHeaders() {
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Adds or sets a key value pair on the request header
func (c SuperFContext) Set (key, val string) {
	c.AddRequestHeader(key, val)
}

func WithDeadline(parent SuperFContext, deadline time.Time) (SuperFContext, context.CancelFunc) {
	deadlineContext, cancelFunc := context.WithDeadline(parent, deadline)
	ctx := NewSuperFContextWithContext(parent.FContext, deadlineContext)
	ctx.AddRequestHeader(superFContextDeadlinePrefix, deadline.String())
	return ctx, cancelFunc
}

func WithTimeout(parent SuperFContext, timeout time.Duration) (SuperFContext, context.CancelFunc) {
	timeoutContext, cancelFunc := context.WithTimeout(parent, timeout)
	ctx := NewSuperFContextWithContext(parent, timeoutContext)
	deadline, _ := ctx.Deadline()
	ctx.AddRequestHeader(superFContextDeadlinePrefix, deadline.String())
	return ctx, cancelFunc
}

func WithValue(parent SuperFContext, key, val string) SuperFContext {
	ctx := NewSuperFContextWithContext(parent, context.WithValue(parent, key, val))
	ctx.AddRequestHeader(superFContextPrefix + key, val)
	return ctx
}
