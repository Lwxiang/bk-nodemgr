package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NewPipeline creates a new pipeline
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		name:       name,
		actionDefs: make([]ActionDef, 0),
	}
}

// Pipeline represents a workflow pipeline
type Pipeline struct {
	name       string
	actionDefs []ActionDef
}

func (p *Pipeline) Name() string {
	return p.name
}

func (p *Pipeline) Next(actionDef ActionDef) *Pipeline {
	p.actionDefs = append(p.actionDefs, actionDef)

	return p
}

func (p *Pipeline) NewTask(timeout time.Duration) (*Task, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	actionNames := make([]string, 0, len(p.actionDefs))
	actionData := make(map[string]*ActionData)
	for index, action := range p.actionDefs {
		actionNames = append(actionNames, action.Name())
		actionData[action.Name()] = &ActionData{
			Name:     action.Name(),
			Index:    index,
			State:    ActionStatePending,
			Messages: make([]string, 0),
		}
	}

	return &Task{
		data: &TaskData{
			TaskID:     uuid.NewString(),
			Pipeline:   p.name,
			Actions:    actionNames,
			ActionData: actionData,
			Timeout:    timeout,
			IsPeriod:   false,
			PeriodSpec: "",
			CreatedAt:  time.Now().Local(),
		},
		pipeline: p,
	}, nil
}

// NewPeriodTask creates a new period task
func (p *Pipeline) NewPeriodTask(timeout time.Duration, spec string) (*Task, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	actionNames := make([]string, 0, len(p.actionDefs))
	for _, actionName := range p.actionDefs {
		actionNames = append(actionNames, actionName.Name())
	}

	content, err := json.Marshal(&ActionPeriodLauncherContent{
		Pipeline:    p.name,
		ActionNames: actionNames,
	})
	if err != nil {
		return nil, err
	}

	launcher := &ActionPeriodLauncher{}
	pipelineName := "period-" + p.name

	return &Task{
		data: &TaskData{
			TaskID:   uuid.NewString(),
			Pipeline: pipelineName,
			Actions:  []string{launcher.Name()},
			ActionData: map[string]*ActionData{
				launcher.Name(): {
					Name:     launcher.Name(),
					Index:    0,
					State:    ActionStatePending,
					Messages: make([]string, 0),
					Content:  string(content),
				},
			},
			Timeout:    timeout,
			IsPeriod:   true,
			PeriodSpec: spec,
			CreatedAt:  time.Now().Local(),
		},
		pipeline: NewPipeline(pipelineName).Next(launcher),
	}, nil
}

func (p *Pipeline) validate() error {
	if len(p.actionDefs) == 0 {
		return errors.New("empty pipeline")
	}

	for i, action := range p.actionDefs {
		if action == nil {
			return fmt.Errorf("action with index %d is nil", i)
		}
	}

	return nil
}
