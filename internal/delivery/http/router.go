package http

import (
	_ "gin-quickstart/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// MapProductRoutes 負責定義所有的網址路徑
func MapProductRoutes(r *gin.Engine, h *ProductHandler, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api/v1")
	{
		p := api.Group("/products")
		{
			p.GET("", h.GetAll)
			p.GET("/:id", h.GetByID)
		}
		auth := api.Group("")
		auth.Use(authMiddleware)
		{
			pp := auth.Group("/products")
			{
				pp.POST("", h.Create)
				pp.PUT("/:id", h.Update)
				pp.PUT("/:id/stock", h.AdjustStock)
				pp.DELETE("/:id", h.Delete)
			}
		}
	}
}

// MapAuthRoutes 定義認證相關的路由
func MapAuthRoutes(r *gin.Engine, authHandler *AuthHandler, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api/v1")
	{
		api.POST("/login", authHandler.Login)
		auth := api.Group("")
		auth.Use(authMiddleware)
		{
			auth.POST("/users", authHandler.Register)
		}

	}
}

func RegisterRoutes(r *gin.Engine, h *ProductHandler, authHandler *AuthHandler, authMiddleware gin.HandlerFunc) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	MapAuthRoutes(r, authHandler, authMiddleware)
	MapProductRoutes(r, h, authMiddleware)
}
