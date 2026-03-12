package http

import (
	_ "gin-quickstart/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// MapRoutes 負責定義所有的網址路徑
func MapRoutes(r *gin.Engine, h *ProductHandler) {
	api := r.Group("/api/v1")
	{
		p := api.Group("/products")
		{
			p.GET("", h.GetAll)
			p.GET("/:id", h.GetByID)
			p.POST("", h.Create)
			p.PUT("/:id", h.Update)
			p.DELETE("/:id", h.Delete)
		}
	}
}

func RegisterRoutes(r *gin.Engine, h *ProductHandler) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	MapRoutes(r, h)
}
