package v1

// UserAuthRequest 用户认证请求
// @Description 用户登录/注册请求参数
type UserAuthRequest struct {
	Username string `json:"username" vd:"$len($)>0&&$len($)<20" example:"testuser" binding:"required" description:"用户名，长度1-20字符"`
	Password string `json:"password" vd:"$len($)>0&&$len($)<20" example:"password123" binding:"required" description:"密码，长度1-20字符"`
}

// UserAuthResponseBody 用户认证响应体
// @Description 用户认证成功返回的用户信息
type UserAuthResponseBody struct {
	UserId      string                      `json:"user_id" example:"123456789" description:"用户ID"`
	Username    string                      `json:"username" example:"testuser" description:"用户名"`
	Nickname    string                      `json:"nickname" example:"测试用户" description:"用户昵称"`
	AvatarUrl   string                      `json:"avatar_url" example:"https://example.com/avatar.jpg" description:"用户头像URL"`
	Certificate UserCertificateResponseBody `json:"certificate" description:"访问凭证信息"`
}

// UserAuthResponse 用户认证响应
// @Description 用户登录/注册响应
type UserAuthResponse struct {
	Resp Response
	Data UserAuthResponseBody `json:"data" description:"用户认证响应数据"`
}

// UserCertificateResponseBody 用户凭证响应体
// @Description 用户访问凭证信息
type UserCertificateResponseBody struct {
	Certificate string `json:"certificate" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." description:"JWT访问令牌"`
	ExpiresIn   int64  `json:"expires_in" example:"3600" description:"令牌过期时间（秒）"`
}

// RefreshResponse 刷新令牌响应
// @Description 刷新访问令牌响应
type RefreshResponse struct {
	Resp Response
	Data UserCertificateResponseBody `json:"data" description:"新的访问凭证"`
}
