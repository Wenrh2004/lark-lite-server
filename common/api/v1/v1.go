package v1

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Response 通用响应结构
// @Description API通用响应格式
type Response struct {
	Code    int         `json:"code" example:"0" description:"响应状态码，0表示成功"`
	Message string      `json:"message" example:"success" description:"响应消息"`
	Data    interface{} `json:"data,omitempty" description:"响应数据"`
}

// ErrorResponse 错误响应
// @Description API错误响应格式
type ErrorResponse struct {
	Code    int    `json:"code" example:"400" description:"错误状态码"`
	Message string `json:"message" example:"请求参数错误" description:"错误消息"`
}

func HandlerSuccess(c *app.RequestContext, data interface{}) {
	resp := Response{Code: errorCodeMap[ErrSuccess], Message: ErrSuccess.Error(), Data: data}
	c.JSON(consts.StatusOK, resp)
}

func HandlerError(c *app.RequestContext, err error) {
	resp := Response{Code: errorCodeMap[err], Message: err.Error()}
	if _, ok := errorCodeMap[err]; !ok {
		resp = Response{Code: 500, Message: "unknown error"}
	}
	c.JSON(consts.StatusOK, resp)
}

type Error struct {
	Code    int
	Message string
}

var errorCodeMap = map[error]int{}

func newError(code int, msg string) error {
	err := errors.New(msg)
	errorCodeMap[err] = code
	return err
}

func (e Error) Error() string {
	return e.Message
}
