package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"todo-app/apperror"

	"github.com/gin-gonic/gin"
)

// RequestID 为每个请求注入唯一 ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := fmt.Sprintf("%08x", rand.Uint32())
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandler 全局统一错误响应处理
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		requestID, _ := c.Get("request_id")
		err := c.Errors.Last().Err

		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{
				"code":       appErr.Code,
				"message":    appErr.Message,
				"request_id": requestID,
			})
			return
		}

		// 未知错误统一返回 500，不暴露内部细节
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":       apperror.ErrInternal.Code,
			"message":    apperror.ErrInternal.Message,
			"request_id": requestID,
		})
	}
}
