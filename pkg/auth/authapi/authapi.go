package authapi

type UserAuth struct {
	BKTicket string `json:"bk_ticket"`
	RTX      string `json:"rtx"`
}
