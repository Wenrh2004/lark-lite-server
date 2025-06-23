package v1

type UserAuthRequest struct {
	Username string `json:"username" vd:"$len($)>0&&$len($)<20"`
	Password string `json:"password" vd:"$len($)>0&&$len($)<20"`
}

type UserAuthResponseBody struct {
	UserId      string                      `json:"user_id"`
	Username    string                      `json:"username"`
	Nickname    string                      `json:"nickname"`
	AvatarUrl   string                      `json:"avatar_url"`
	Certificate UserCertificateResponseBody `json:"certificate"`
}

type UserAuthResponse struct {
	Resp Response
	Data UserAuthResponseBody `json:"data"`
}

type UserCertificateResponseBody struct {
	Certificate string `json:"certificate"`
	ExpiresIn   int64  `json:"expires_in"`
}

type RefreshResponse struct {
	Resp Response
	Data UserCertificateResponseBody `json:"data"`
}
