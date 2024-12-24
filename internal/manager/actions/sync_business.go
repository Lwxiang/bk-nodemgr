package actions

import (
	"sync"
	"time"

	"git.woa.com/bk-gse/bk-nodeman/internal/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/cmdb"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
	"git.woa.com/bk-gse/bk-nodeman/pkg/workflow"
)

// NewActionSyncBusinessFromCMDB creates a new ActionSyncBusinessFromCMDB.
func NewActionSyncBusinessFromCMDB(cmdbHandler cmdb.Handler, topoStorage topo.Storage) *ActionSyncBusinessFromCMDB {
	return &ActionSyncBusinessFromCMDB{
		cmdbHandler: cmdbHandler,
		topoStorage: topoStorage,
	}
}

const (
	// ActionNameSyncBusinessFromCMDB defines the name of this action.
	ActionNameSyncBusinessFromCMDB = "sync_business_from_cmdb"
)

// ActionSyncBusinessFromCMDB sync business info from cmdb.
type ActionSyncBusinessFromCMDB struct {
	cmdbHandler cmdb.Handler
	topoStorage topo.Storage
}

// Name returns the name of the action.
func (a *ActionSyncBusinessFromCMDB) Name() string {
	return ActionNameSyncBusinessFromCMDB
}

// Version returns the version of the action.
func (a *ActionSyncBusinessFromCMDB) Version() string {
	return "v1"
}

// Description returns the description of the action.
func (a *ActionSyncBusinessFromCMDB) Description() string {
	return "sync business info from cmdb and update to storage"
}

// Timeout returns the timeout of this action.
func (a *ActionSyncBusinessFromCMDB) Timeout() time.Duration {
	return 1 * time.Minute
}

// MaxRetryCount returns the max retry count of this action.
func (a *ActionSyncBusinessFromCMDB) MaxRetryCount() uint {
	return 0
}

// Do executes the action.
func (a *ActionSyncBusinessFromCMDB) Do(ctx *workflow.ActionContext) error {
	blog.Infof("start syncing business info from cmdb. info: %s", ctx.Action.Info())
	ctx.Action.Log("start syncing business info from cmdb")

	var wg sync.WaitGroup

	page := cmdb.Page{
		Start: 0,
		Limit: 500,
	}
	for {
		businesses, _, err := a.cmdbHandler.Tenant(1).User(nil).SearchBusiness(page)
		if err != nil {
			blog.Errorf("failed to get business info from cmdb. err: %v", err)
			ctx.Action.Log("failed to get business info from cmdb. err: " + err.Error())

			return err
		}

		if len(businesses) == 0 {
			break
		}

		for _, business := range businesses {
			wg.Add(1)

			func(biz *types.Business) {
				defer wg.Done()

				if err := a.topoStorage.UpsertBusiness(biz); err != nil {
					blog.Errorf("failed to upsert business info. biz-id(%d), err: %v", biz.BizID, err)
				}
			}(business)
		}

		page.Start += page.Limit
	}

	wg.Wait()
	blog.Infof("succeed to sync business info from cmdb. info: %s", ctx.Action.Info())

	return nil
}
