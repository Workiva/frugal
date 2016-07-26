package parser

import (
	"fmt"
	"strings"
)

// Error object to keep track of errors and tree position
type Error struct {
	prefix   string
	children []Error
}

// NewError ...
func NewError(s string) Error {
	return Error{prefix: s}
}

// NewErrorf ...
func NewErrorf(s string, a ...interface{}) Error {
	return Error{prefix: fmt.Sprintf(s, a...)}
}

// Prefix defines position in tree
func (e *Error) Prefix(s string) {
	if e == nil || len(e.children) == 0 {
		e = nil
	} else {
		e.prefix = s
	}
}

// Append to the Error object
func (e *Error) Append(err Error) {
	if err.prefix == "" {
		return
	}
	if e == nil {
		*e = Error{children: make([]Error, 0)}
	}
	e.children = append(e.children, err)
}

// Implement the error interface so we can pass this as a golang error type
func (e *Error) Error() string {
	if e == nil {
		return ""
	} else if len(e.children) == 0 {
		return e.prefix
	}
	// get all the errors recursively
	var errors []string
	for _, child := range e.children {
		errors = append(errors, strings.Split(child.Error(), "\n")...)
	}
	// format the response
	for i, err := range errors {
		errors[i] = fmt.Sprintf("%s%s", e.prefix, err)
	}
	return strings.Join(errors, "\n")
}
