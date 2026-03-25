package apperror

import "net/http"

// AppError 业务错误类型
type AppError struct {
	HTTPStatus int    // HTTP 状态码
	Code       string // 业务错误码
	Message    string // 错误描述
}

func (e *AppError) Error() string {
	return e.Message
}

// 预定义业务错误
var (
	ErrInvalidInput = &AppError{http.StatusBadRequest, "INVALID_INPUT", "invalid input"}
	ErrUnauthorized = &AppError{http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized"}
	ErrNotFound     = &AppError{http.StatusNotFound, "NOT_FOUND", "resource not found"}
	ErrInternal     = &AppError{http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"}
)

// New 创建自定义错误
func New(httpStatus int, code, message string) *AppError {
	return &AppError{httpStatus, code, message}
}
