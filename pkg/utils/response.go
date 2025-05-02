package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

const (
	SuccessCode      = 0
	BadRequestCode   = 400
	UnauthorizedCode = 401
	NotFoundCode     = 404
	ServerErrorCode  = 500
	DuplicateCode    = 409
)

// Success 成功返回
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: SuccessCode,
		Msg:  "Success",
		Data: data,
	})
}

// Fail 失败返回
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
