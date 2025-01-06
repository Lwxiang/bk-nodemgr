package cmdb

import (
	"context"

	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/client"
	"git.woa.com/bk-gse/bk-nodeman/pkg/types"
)

// This document is responsible for processing the conversion of original requests and responses
// from the third-party system and the internal data of the nodeman system.

// Handler the handler of cmdb.
type Handler interface {
	// ListBizHosts list biz hosts
	ListBizHosts(ctx context.Context, biz types.Business, page types.Page) ([]types.Host, error)
	// SearchBusiness search business
	SearchBusiness(ctx context.Context, page types.Page) ([]types.Business, error)
}

type handler struct {
	cli *cmdbCli
}

// NewHandler initialize a new cmdb handler.
func NewHandler(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := NewClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBizHosts list biz hosts
func (h *handler) ListBizHosts(ctx context.Context, biz types.Business, page types.Page) ([]types.Host, error) {
	// TODO: 补充参数构建信息
	req := &ReqListBizHosts{
		BKBizID: biz.BizID,
		Page: Page{
			Start: page.Start,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.listBizHosts(ctx, req)
	if err != nil {
		return nil, err
	}

	hosts := make([]types.Host, len(resp.Info))
	for idx, host := range resp.Info {
		hosts[idx] = types.Host{
			// TODO: 补充租户信息
			TenantID: "",
			CloudID:  host.BKCloudID,
			BizID:    req.BKBizID,
			HostID:   host.BKHostID,
			InnerIP:  host.BKHostInnerIPV4,
			Mac:      host.BKMac,
			OSType:   host.BKOsType,
		}
	}

	return hosts, nil
}

// SearchBusiness search business
func (h *handler) SearchBusiness(ctx context.Context, page types.Page) ([]types.Business, error) {
	req := &ReqSearchBusiness{
		Page: Page{
			Start: page.Start,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
		Fields: nil,
	}

	resp, err := h.cli.searchBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	bizs := make([]types.Business, len(resp.Info))
	for idx, business := range resp.Info {
		// TODO: 补充租户信息
		bizs[idx] = types.Business{
			TenantID: "",
			BizID:    business.BKBizID,
			BizName:  business.BKBizName,
		}
	}

	return bizs, nil
}
