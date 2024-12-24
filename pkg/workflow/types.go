package workflow

type ActionState string

const (
	ActionStatePending ActionState = "pending"
	ActionStateRunning ActionState = "running"
	ActionStateSuccess ActionState = "success"
	ActionStateFailed  ActionState = "failed"
	ActionStateTimeout ActionState = "timeout"
	ActionStateSkipped ActionState = "skipped"
	ActionStateStopped ActionState = "stopped"
	ActionStateUnknown ActionState = "unknown"
)

type TaskState string

const (
	TaskStatePending ActionState = "pending"
	TaskStateRunning ActionState = "running"
	TaskStateSuccess ActionState = "success"
	TaskStateFailed  ActionState = "failed"
	TaskStateTimeout ActionState = "timeout"
	TaskStateSkipped ActionState = "skipped"
	TaskStateStopped ActionState = "stopped"
	TaskStateUnknown ActionState = "unknown"
)
