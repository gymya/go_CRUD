package main

import (
	"context"
	"gin-quickstart/configs"
	"gin-quickstart/internal/cache"
	"gin-quickstart/internal/delivery/http"
	"gin-quickstart/internal/events"
	kafkapub "gin-quickstart/internal/messaging/kafka"
	repository "gin-quickstart/internal/repository/SQLite"
	"gin-quickstart/internal/service"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           CRUD API
// @version         1.0
// @description     Gin CRUD
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("未找到 .env，將使用系統環境變數: %v", err)
	}

	appCfg, err := configs.LoadConfig("configs/app.yaml")
	if err != nil {
		log.Fatalf("無法載入設定檔: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET 未設定")
	}

	// 1. 初始化 SQLite 連線
	// 這會在專案根目錄產生一個 shop.db 檔案
	db, err := gorm.Open(sqlite.Open("shop.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("無法連接資料庫:", err)
	}

	// 2. 初始化資料層 (切換為 SQLite)
	repo := repository.NewSqliteRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 3. 初始化 Redis 快取
	var cacheStore cache.Cache
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := getEnvInt("REDIS_DB", 0)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	cacheStore = cache.NewRedisCache(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis 連線失敗，快取將在 Redis 可用時自動恢復: %v", err)
	}

	// 4. 初始化業務層
	var stockPublisher events.StockEventPublisher
	kafkaBrokersRaw := strings.TrimSpace(getEnv("KAFKA_BROKERS", ""))
	kafkaTopic := getEnv("KAFKA_STOCK_TOPIC", "product.stock.updated")
	if kafkaBrokersRaw != "" && strings.TrimSpace(kafkaTopic) != "" {
		kafkaBrokers := strings.Split(kafkaBrokersRaw, ",")
		producer, err := kafkapub.NewStockProducer(kafkaBrokers, kafkaTopic, "product-service")
		if err != nil {
			log.Printf("Kafka producer 初始化失敗，Stock 更新將無法送出: %v", err)
		} else {
			stockPublisher = producer
		}

		consumerGroup := getEnv("KAFKA_CONSUMER_GROUP", "product-stock-consumer")
		consumer, err := kafkapub.NewStockConsumer(kafkaBrokers, kafkaTopic, consumerGroup, repo, cacheStore)
		if err != nil {
			log.Printf("Kafka consumer 初始化失敗，Stock 更新將無法同步到 DB: %v", err)
		} else {
			go func() {
				if err := consumer.Start(context.Background()); err != nil {
					log.Printf("Kafka consumer 停止: %v", err)
				}
			}()
		}
	}

	svc := service.NewProductService(repo, cacheStore, stockPublisher)
	authSvc := service.NewAuthService(userRepo, jwtSecret, appCfg.JWTExpiry())

	// 5. 初始化 Handler (最上層，注入 Service)
	handler := &http.ProductHandler{Svc: svc}
	authHandler := &http.AuthHandler{Auth: authSvc}

	// 6. 設定路由與啟動
	router := gin.Default()
	router.Use(http.ErrorHandler())
	authMiddleware := http.AuthMiddleware(jwtSecret)
	http.RegisterRoutes(router, handler, authHandler, authMiddleware)

	router.Run("localhost:8080")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
