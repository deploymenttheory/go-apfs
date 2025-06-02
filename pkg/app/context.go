package app

import (
	"context"
	"time"
)

// Context holds application-wide configuration and state
type Context struct {
	context.Context

	// Output preferences
	OutputFormat string
	Verbose      bool
	Quiet        bool
	NoColor      bool

	// Common timeouts
	DefaultTimeout time.Duration

	// Progress reporting
	ProgressCallback func(message string, percent int)
}

// NewContext creates a new application context
func NewContext() *Context {
	return &Context{
		Context:        context.Background(),
		DefaultTimeout: 30 * time.Second,
	}
}

// WithTimeout creates a context with timeout
func (c *Context) WithTimeout(timeout time.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	newCtx := *c
	newCtx.Context = ctx
	return &newCtx, cancel
}

// WithCancel creates a cancellable context
func (c *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	newCtx := *c
	newCtx.Context = ctx
	return &newCtx, cancel
}

// SetProgress sets the progress callback function
func (c *Context) SetProgress(callback func(string, int)) {
	c.ProgressCallback = callback
}

// Progress reports progress if callback is set
func (c *Context) Progress(message string, percent int) {
	if c.ProgressCallback != nil {
		c.ProgressCallback(message, percent)
	}
}

// Log outputs a message based on verbosity settings
func (c *Context) Log(message string) {
	if !c.Quiet && c.Verbose {
		println(message)
	}
}

// Error outputs an error message unless quiet
func (c *Context) Error(message string) {
	if !c.Quiet {
		println("Error:", message)
	}
}
