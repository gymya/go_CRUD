package http

import (
	"errors"
	"gin-quickstart/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 執行後續的 Handler

		// 檢查是否有錯誤被存入 c.Errors
		if len(c.Errors) > 0 {
			// 如果回應已經寫入，就不再處理
			if c.Writer.Written() {
				return
			}

			err := c.Errors.Last().Err
			if err == nil {
				return
			}

			// 允許被包裝的 AppError 或非指標錯誤
			var appErr *domain.AppError
			if errors.As(err, &appErr) && appErr != nil {
				c.AbortWithStatusJSON(appErr.Code, appErr)
				return
			}

			// 如果是未知的錯誤回傳 500
			c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrInternal)
		}
	}
}
