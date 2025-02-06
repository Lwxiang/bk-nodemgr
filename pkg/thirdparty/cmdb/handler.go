package cmdb

import (
	"context"
	"errors"

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

	// SearchNetArea search net area
	SearchNetArea(ctx context.Context, page types.Page) ([]types.NetArea, error)
}

type handler struct {
	cli *cli
}

// New initialize a new cmdb handler.
func New(c *client.Capability, conf *Config) (Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	return &handler{cli: cli}, nil
}

// ListBizHosts list biz hosts
func (h *handler) ListBizHosts(ctx context.Context, biz types.Business, page types.Page) ([]types.Host, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	req := &ListBizHostsReq{
		BKBizID: biz.BizID,
		Page: Page{
			Start: page.Offset,
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
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	req := &SearchBusinessReq{
		BKSupplierAccount: "tencent",
		Page: Page{
			Start: page.Offset,
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

// SearchNetArea search net area
func (h *handler) SearchNetArea(ctx context.Context, page types.Page) ([]types.NetArea, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	req := &SearchCloudAreaReq{
		Page: Page{
			Start: page.Offset,
			Limit: page.Limit,
			Sort:  page.Sort,
		},
	}

	resp, err := h.cli.searchCloudArea(ctx, req)
	if err != nil {
		return nil, err
	}

	netAreas := make([]types.NetArea, len(resp.Info))
	for idx, netArea := range resp.Info {
		// TODO: 补充租户信息
		netAreas[idx] = types.NetArea{
			TenantID:  "",
			CloudID:   netArea.BkCloudID,
			CloudName: netArea.BkCloudName,
		}
	}

	return netAreas, nil
}
