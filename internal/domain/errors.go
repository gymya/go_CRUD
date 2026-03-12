package domain

import "net/http"

// AppError 自定義的錯誤結構
type AppError struct {
	Code    int    `json:"-"`       // HTTP 狀態碼 (不顯示在 JSON 裡)
	Message string `json:"message"` // 給前端看的錯誤訊息
}

func (e *AppError) Error() string {
	return e.Message
}

// 預定義常用的錯誤，方便 Service 直接使用
var (
	ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "資源不存在"}
	ErrInvalidInput = &AppError{Code: http.StatusUnprocessableEntity, Message: "無效的輸入參數"}
	ErrInternal     = &AppError{Code: http.StatusInternalServerError, Message: "伺服器內部錯誤"}
)

// NewError 方便快速建立自定義訊息的錯誤
func NewError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
