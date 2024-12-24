package workflow

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ActionData struct {
	Name      string
	Index     int
	State     ActionState
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
	Messages  []string
	Content   string
}

type Action struct {
	data *ActionData

	task *Task
}

func (a *Action) Info() string {
	return fmt.Sprintf("task(%s), index(%d), action(%s)", a.task.data.TaskID, a.data.Index, a.data.Name)
}

func (a *Action) Log(messages ...string) {
	a.data.Messages = append(a.data.Messages, messages...)
}

func (a *Action) FlushLog() error {
	return a.task.saveMethod(a.task.data)
}

type TaskData struct {
	TaskID       string
	Pipeline     string
	Actions      []string
	ActionData   map[string]*ActionData
	ParentTaskID string
	Timeout      time.Duration

	IsPeriod   bool
	PeriodSpec string

	InitContent string

	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
	StoppedAt time.Time
}

type Task struct {
	pipeline *Pipeline

	data *TaskData

	saveMethod func(*TaskData) error

	mutex sync.RWMutex
}

func (t *Task) LastActionState(actionName string) ActionState {
	lastActionState := ActionStateSuccess

	for _, action := range t.data.Actions {
		if action != actionName {
			lastActionData, ok := t.data.ActionData[action]
			if !ok {
				lastActionState = ActionStateUnknown
			}

			lastActionState = lastActionData.State
			continue
		}

		return lastActionState
	}

	return ActionStateUnknown
}

func (t *Task) Action(actionName string) (*Action, error) {
	actionData, ok := t.data.ActionData[actionName]
	if !ok {
		return nil, fmt.Errorf("action %s not found", actionName)
	}

	return &Action{
		data: actionData,
		task: t,
	}, nil
}

func (t *Task) save() error {
	if t.saveMethod == nil {
		return errors.New("no save method implemented")
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.saveMethod(t.data)
}
