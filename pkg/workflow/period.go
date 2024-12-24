package workflow

import (
	"encoding/json"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
)

func NewActionPeriodLauncher(workflowMgr *Manager) *ActionPeriodLauncher {
	return &ActionPeriodLauncher{workflowMgr: workflowMgr}
}

const (
	// ActionNamePeriodLauncher defines the name of this action.
	ActionNamePeriodLauncher = "period_launcher"
)

type ActionPeriodLauncher struct {
	workflowMgr *Manager
}

func (a *ActionPeriodLauncher) Name() string {
	return ActionNamePeriodLauncher
}

func (a *ActionPeriodLauncher) Version() string {
	return "v1"
}

func (a *ActionPeriodLauncher) Description() string {
	return "period task launcher"
}

func (a *ActionPeriodLauncher) Timeout() time.Duration {
	return 1 * time.Minute
}

func (a *ActionPeriodLauncher) MaxRetryCount() uint {
	return 0
}

func (a *ActionPeriodLauncher) Do(ctx *ActionContext) error {
	blog.Info("begin to do period task")

	var content ActionPeriodLauncherContent
	if err := json.Unmarshal([]byte(ctx.Action.data.Content), &content); err != nil {
		blog.Errorf("failed to unmarshal content, err: %v", err)

		return err
	}

	pipeline := NewPipeline(content.Pipeline)
	for _, actionName := range content.ActionNames {
		pipeline = pipeline.Next(a.workflowMgr.GetRegisteredAction(actionName))
	}

	task, err := pipeline.NewTask(ctx.Action.task.data.Timeout)
	if err != nil {
		blog.Errorf("failed to generate a new task, err: %v", err)

		return err
	}

	if err = a.workflowMgr.DispatchTask(task); err != nil {
		blog.Errorf("failed to dispatch a new task, err: %v", err)

		return err
	}

	blog.Infof("successfully dispatched a new task(%s) from period launcher(%s)", task.data.TaskID, ctx.Action.task.data.TaskID)

	return nil
}

// ActionPeriodLauncherContent defines the content of ActionPeriodLauncher.
type ActionPeriodLauncherContent struct {
	Pipeline    string   `json:"pipeline"`
	ActionNames []string `json:"action_names"`
}
