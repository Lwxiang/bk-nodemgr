// Package types define all common types used in nodeman runtime.
// Everything from API or Database should be converted into types in this package before using.
package types

// Tenant represents a blueking tenant.
type Tenant struct {
	// tenant-id is the unique identifier for a tenant in a Blueking environment.
	TenantID string

	// there is only one admin tenant in a Blueking environment.
	// others are all normal tenants.
	IsAdmin bool
}

// Business represents a cmdb business under a tenant.
type Business struct {
	// belongs to.
	TenantID string

	// biz-id is the unique identifier for a business.
	BizID   int
	BizName string
}

// NetArea represents a cmdb net-area. In which IPs will not be duplicated.
type NetArea struct {
	// belongs to
	TenantID string

	// cloud-id is the unique identifier for a net-area.
	CloudID int
}

// Host represents a cmdb host.
type Host struct {
	// belongs to
	TenantID string
	CloudID  int
	BizID    int

	// host-id is the unique identifier for a host.
	HostID int

	// host informations.
	InnerIP string
	Mac     string
	OSType  string
}
