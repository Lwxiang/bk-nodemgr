package cmdb

// Page describe the page data in request.
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// HostInfo describe the information of single host.
type HostInfo struct {
	BKCloudID     int    `json:"bk_cloud_id"`
	BKHostID      int    `json:"bk_host_id"`
	BKHostInnerIP string `json:"bk_host_innerip"`
	BKMac         string `json:"bk_mac"`
	BKOSType      string `json:"bk_os_type"`
}

// BusinessInfo describe the information of single business.
type BusinessInfo struct {
	BKBizID           int    `json:"bk_biz_id"`
	BKBizName         string `json:"bk_biz_name"`
	BKBizMaintainer   string `json:"bk_biz_maintainer"`
	BKBizProducer     string `json:"bk_biz_producer"`
	BKBizDeveloper    string `json:"bk_biz_developer"`
	BKBizTester       string `json:"bk_biz_tester"`
	TimeZone          string `json:"time_zone"`
	Language          string `json:"language"`
	BKSupplierAccount string `json:"bk_supplier_account"`
	CreateTime        string `json:"create_time"`
	LastTime          string `json:"last_time"`

	// default field describes business type.
	Default     int    `json:"default"`
	Operator    string `json:"operator"`
	LifeCycle   string `json:"life_cycle"`
	BKCreatedAt string `json:"bk_created_at"`
	BKUpdatedAt string `json:"bk_updated_at"`
	BKCreatedBy string `json:"bk_created_by"`
}

type ObjectInfo struct {
	BKObjectID        string `json:"bk_obj_id"`
	BKObjectName      string `json:"bk_obj_name"`
	BKSupplierAccount string `json:"bk_supplier_account"`
}

// RespCommon describe the common part of response data.
type RespCommon struct {
	Result  bool   `json:"result"`
	Code    int    `json:"bk_error_code"`
	Message string `json:"bk_error_message"`
}

// ReqListBizHosts describe the request data of list_biz_hosts.
type ReqListBizHosts struct {
	// response page settings.
	Page Page `json:"page"`

	// biz id of this request.
	BKBizID int `json:"bk_biz_id"`

	// expected resposne fields.
	Fields []string `json:"fields"`
}

// RespListBizHosts describe the response data of list_biz_hosts.
type RespListBizHosts struct {
	RespCommon

	Data struct {
		Count int         `json:"count"`
		Info  []*HostInfo `json:"info"`
	} `json:"data"`
}

// ReqSearchBusiness describe the request data of search_business.
type ReqSearchBusiness struct {
	// response page settings.
	Page Page `json:"page"`

	// expected response fields.
	Fields []string `json:"fields"`
}

// RespSearchBusiness describe the response data of search_business.
type RespSearchBusiness struct {
	RespCommon

	Data struct {
		Count int             `json:"count"`
		Info  []*BusinessInfo `json:"info"`
	} `json:"data"`
}
