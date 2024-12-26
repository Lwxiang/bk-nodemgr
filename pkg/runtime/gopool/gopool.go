// Package gopool provides a goroutine pool.
package gopool

import (
	"bytes"
	"fmt"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

// Pool is the interface of goroutine pool.
type Pool interface {
	// Go executes function f with recover.
	Go(f func() error)
	// Wait waits for all goroutines to finish and returns the first non-nil error (if any) from them.
	Wait() error
	// SetLimit sets the goroutine limit(capability) of the pool.
	SetLimit(limit int)
}

// NewPool returns a new goroutine pool.
func NewPool() Pool {
	p := &pool{
		eg: new(errgroup.Group),
	}

	return p
}

type pool struct {
	eg *errgroup.Group
}

// Go executes function f with recover.
func (p *pool) Go(fn func() error) {
	p.eg.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				// The first line of the stack trace is of the form "goroutine N [status]:",
				// but by the time the panic reaches Do the goroutine may no longer exist,
				// and its status will have changed. Trim out the misleading line.
				if line := bytes.IndexByte(stack[:], '\n'); line >= 0 {
					stack = stack[line+1:]
				}

				err = fmt.Errorf("go pool panic: %v, stack: %s", r, stack)
			}
		}()

		return fn()
	})
}

// Wait waits for all goroutines to finish and returns the first non-nil error (if any) from them.
func (p *pool) Wait() error {
	return p.eg.Wait()
}

// SetLimit sets the goroutine limit(capability) of the pool.
func (p *pool) SetLimit(limit int) {
	p.eg.SetLimit(limit)
}
