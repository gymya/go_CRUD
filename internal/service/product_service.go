package service

import (
	"context"
	"encoding/json"
	"gin-quickstart/internal/cache"
	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/events"
	"net/http"
	"strconv"
	"time"
)

// ProductService 定義業務邏輯介面
type ProductService interface {
	GetAllProducts() []domain.Product
	GetProduct(id int) (*domain.Product, error)
	CreateProduct(p domain.Product) (domain.Product, error)
	UpdateProduct(id int, p domain.Product) (*domain.Product, error)
	AdjustStock(id int, stock int) error
	DeleteProduct(id int) error
}

// 實作結構體，它需要依賴 Repo
type productService struct {
	repo      domain.ProductRepository
	cache     cache.Cache
	publisher events.StockEventPublisher
}

// NewProductService 初始化 ProductService，並注入 Repository 和 Cache
func NewProductService(repo domain.ProductRepository, cache cache.Cache, publisher events.StockEventPublisher) ProductService {
	return &productService{repo: repo, cache: cache, publisher: publisher}
}

// --- 實作各個方法 ---

func (s *productService) GetAllProducts() []domain.Product {
	ctx := context.Background()
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheAllProductsKey); err == nil {
			var products []domain.Product
			if err := json.Unmarshal([]byte(cached), &products); err == nil {
				return products
			}
		}
	}

	products := s.repo.GetAll()
	if s.cache != nil {
		if payload, err := json.Marshal(products); err == nil {
			_ = s.cache.Set(ctx, cacheAllProductsKey, payload, cacheTTL)
		}
	}
	return products
}

func (s *productService) GetProduct(id int) (*domain.Product, error) {
	ctx := context.Background()
	if s.cache != nil {
		cacheKey := productByIDCacheKey(id)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			var product domain.Product
			if err := json.Unmarshal([]byte(cached), &product); err == nil {
				return &product, nil
			}
		}
	}

	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if payload, err := json.Marshal(p); err == nil {
			_ = s.cache.Set(ctx, productByIDCacheKey(id), payload, cacheTTL)
		}
	}
	return p, nil
}

func (s *productService) CreateProduct(p domain.Product) (domain.Product, error) {
	// 禁止建立價格低於 10 元的商品
	if p.Price < 10 {
		return domain.Product{}, domain.NewError(http.StatusUnprocessableEntity, "商品價格太低")
	}
	created := s.repo.Create(p)
	if s.cache != nil {
		ctx := context.Background()
		_ = s.cache.Del(ctx, cacheAllProductsKey)
		_ = s.cache.Del(ctx, productByIDCacheKey(created.ID))
	}
	return created, nil
}

func (s *productService) UpdateProduct(id int, p domain.Product) (*domain.Product, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updatedInput := *existing
	if p.Name != "" {
		updatedInput.Name = p.Name
	}
	if p.Price != 0 {
		updatedInput.Price = p.Price
	}
	updatedInput.Stock = existing.Stock

	updated, err := s.repo.Update(id, updatedInput)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		ctx := context.Background()
		_ = s.cache.Del(ctx, cacheAllProductsKey)
		_ = s.cache.Del(ctx, productByIDCacheKey(id))
	}
	return updated, nil
}

func (s *productService) AdjustStock(id int, stock int) error {
	if stock < 0 {
		return domain.ErrInvalidInput
	}

	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if s.publisher == nil {
		return domain.NewError(http.StatusServiceUnavailable, "Kafka producer not configured")
	}

	event := events.StockUpdateEvent{
		EventType:  events.EventTypeProductStockUpdated,
		OccurredAt: time.Now().UTC(),
		ProductID:  id,
		Stock:      stock,
		Source:     "product-service",
	}

	ctx := context.Background()
	if err := s.publisher.PublishStockUpdate(ctx, event); err != nil {
		return domain.NewError(http.StatusServiceUnavailable, "Failed to publish stock update")
	}

	if s.cache != nil {
		_ = s.cache.Del(ctx, cacheAllProductsKey)
		_ = s.cache.Del(ctx, productByIDCacheKey(id))
	}

	return nil
}

func (s *productService) DeleteProduct(id int) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	if s.cache != nil {
		ctx := context.Background()
		_ = s.cache.Del(ctx, cacheAllProductsKey)
		_ = s.cache.Del(ctx, productByIDCacheKey(id))
	}
	return nil
}

const (
	cacheTTL              = 60 * time.Second
	cacheAllProductsKey   = "products:all"
	cacheProductKeyPrefix = "products:id:"
)

func productByIDCacheKey(id int) string {
	return cacheProductKeyPrefix + strconv.Itoa(id)
}
