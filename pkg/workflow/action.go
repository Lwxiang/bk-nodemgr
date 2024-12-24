package workflow

import (
	"context"
	"time"
)

type ActionContext struct {
	Ctx context.Context

	Action *Action
}

// Action represents a workflow action, which is a single basic step of work.
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
