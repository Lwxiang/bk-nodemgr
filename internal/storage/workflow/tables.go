package workflow

import (
	"time"

	baseStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage"
)

// StoppingTask represents a workflow stopping task.
type StoppingTask struct {
	TaskID   string    `json:"task_id" bson:"task_id"`
	ExpireAt time.Time `json:"expire_at" bson:"expire_at"`
}

// TableStoppingTask represents the complete db structures of a stopping task.
type TableStoppingTask struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  *StoppingTask `json:"data" bson:"data"`
}

// ActionData represents a action data.
type ActionData struct {
	Name      string    `json:"name" bson:"name"`
	State     string    `json:"state" bson:"state"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	EndedAt   time.Time `json:"ended_at" bson:"ended_at"`
	StoppedAt time.Time `json:"stopped_at" bson:"stopped_at"`
	Messages  []string  `json:"messages" bson:"messages"`
	Content   string    `json:"content" bson:"content"`
}

// TaskData represents a task data.
type TaskData struct {
	TaskID         string                 `json:"task_id" bson:"task_id"`
	Actions        []string               `json:"actions" bson:"actions"`
	ActionData     map[string]*ActionData `json:"action_data" bson:"action_data"`
	Pipeline       string                 `json:"pipeline" bson:"pipeline"`
	ParentTaskID   string                 `json:"parent_task_id" bson:"parent_task_id"`
	TemplateTaskID string                 `json:"template_task_id" bson:"template_task_id"`

	Timeout time.Duration `json:"timeout" bson:"timeout"`

	IsPeriod   bool   `json:"is_period" bson:"is_period"`
	PeriodSpec string `json:"period_spec" bson:"period_spec"`

	InitContent string `json:"init_content" bson:"init_content"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	EndedAt   time.Time `json:"ended_at" bson:"ended_at"`
	StoppedAt time.Time `json:"stopped_at" bson:"stopped_at"`
}

// TableTaskData represents the complete db structures of a task data.
type TableTaskData struct {
	baseStorage.BasicInfo `json:"basic" bson:"basic"`
	Data                  *TaskData `json:"data" bson:"data"`
}

// TableTaskDataChangeEvent represents the complete db structures of a task data change event.
type TableTaskDataChangeEvent struct {
	FullDocument *TableTaskData `json:"fullDocument" bson:"fullDocument"`
}
