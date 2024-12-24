package workflow

import (
	"context"
)

// Storage represents a workflow storage handler.
// it will be used to store the custom records during task scheduling.
type Storage interface {
	CreateTaskData(data *TaskData) error

	GetTaskData(taskID string) (*TaskData, error)

	UpdateTaskData(data *TaskData) error

	MarkTaskStopping(taskID string) error

	WatchTaskStopping(ctx context.Context, taskID string) <-chan struct{}
}
