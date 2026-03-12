package main

import (
	"gin-quickstart/internal/delivery/http"
	repository "gin-quickstart/internal/repository/SQLite"
	"gin-quickstart/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           CRUD API
// @version         1.0
// @description     Gin CRUD
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// 1. 初始化 SQLite 連線
	// 這會在專案根目錄產生一個 shop.db 檔案
	db, err := gorm.Open(sqlite.Open("shop.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("無法連接資料庫:", err)
	}

	// 2. 初始化資料層 (切換為 SQLite)
	repo := repository.NewSqliteRepository(db)

	// 3. 初始化業務層
	svc := service.NewProductService(repo)

	// 4. 初始化 Handler (最上層，注入 Service)
	handler := &http.ProductHandler{Svc: svc}

	// 4. 設定路由與啟動
	router := gin.Default()
	router.Use(http.ErrorHandler())
	http.RegisterRoutes(router, handler)

	router.Run("localhost:8080")
}
