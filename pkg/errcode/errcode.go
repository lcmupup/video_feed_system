package errcode

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一错误码定义
const (
	CodeSuccess       = 0   // 成功
	CodeBadRequest    = 400 // 参数错误
	CodeUnauthorized  = 401 // 未授权
	CodeNotFound      = 404 // 资源不存在
	CodeConflict      = 409 // 冲突（如重复点赞）
	CodeInternalError = 500 // 服务器内部错误
)

// httpStatusFromCode 根据业务错误码映射 HTTP 状态码
func httpStatusFromCode(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

// Response 统一响应基础函数
func Response(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(httpStatusFromCode(code), gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// Success 成功响应（无 data）
func Success(c *gin.Context, msg string) {
	Response(c, CodeSuccess, msg, nil)
}

// SuccessWithData 成功响应（带 data）
func SuccessWithData(c *gin.Context, msg string, data interface{}) {
	Response(c, CodeSuccess, msg, data)
}

// SuccessWithPagination 成功响应（分页列表）
func SuccessWithPagination(c *gin.Context, msg string, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"code":      CodeSuccess,
		"msg":       msg,
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Error 错误响应（根据 code 自动映射 HTTP 状态码）
func Error(c *gin.Context, code int, msg string) {
	Response(c, code, msg, nil)
}

// ErrorWithData 错误响应（带 data）
func ErrorWithData(c *gin.Context, code int, msg string, data interface{}) {
	Response(c, code, msg, data)
}
