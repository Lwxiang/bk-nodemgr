package apigw

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Identity    struct {
		UserType string `json:"user_type"`
		Username string `json:"username"`
	} `json:"identity"`
	RefreshToken string `json:"refresh_token"`
}

type AuthHeader struct {
	BKAppCode   string `json:"bk_app_code"`
	BKAppSecret string `json:"bk_app_secret"`
	BKUsername  string `json:"bk_username"`
}

type ReqSSMAccessToken struct {
	GrantType  string `json:"grant_type"`
	IDProvider string `json:"id_provider"`
	BKToken    string `json:"bk_token"`
}

type RespSSMAccessToken struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *AccessToken `json:"data"`
}
