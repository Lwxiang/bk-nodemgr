package workflow

import (
	"context"
	"time"
)

// ActionContext represents the context of an action.
type ActionContext struct {
	// Context is the context of the action
	Ctx context.Context
	// Action is the action being executed
	Action *Action
}

// ActionDef represents a workflow action, which is a single basic step of work.
type ActionDef interface {
	// Name returns the name of the action
	Name() string

	// Version returns the version of the action
	Version() string

	// Description returns the description of the action
	Description() string

	// Timeout returns the timeout of the action
	Timeout() time.Duration

	// MaxRetryCount returns the max retry count of the action
	MaxRetryCount() uint

	// Do executes the action, with specified context
	Do(*ActionContext) error
}
