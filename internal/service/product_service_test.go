package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gin-quickstart/internal/domain"
	"gin-quickstart/internal/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) GetAll() []domain.Product {
	args := m.Called()
	if out, ok := args.Get(0).([]domain.Product); ok {
		return out
	}
	return nil
}

func (m *mockProductRepo) GetByID(id int) (*domain.Product, error) {
	args := m.Called(id)
	if out, ok := args.Get(0).(*domain.Product); ok {
		return out, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProductRepo) Create(p domain.Product) domain.Product {
	args := m.Called(p)
	if out, ok := args.Get(0).(domain.Product); ok {
		return out
	}
	return domain.Product{}
}

func (m *mockProductRepo) Update(id int, p domain.Product) (*domain.Product, error) {
	args := m.Called(id, p)
	if out, ok := args.Get(0).(*domain.Product); ok {
		return out, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProductRepo) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *mockCache) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type mockStockPublisher struct {
	mock.Mock
}

func (m *mockStockPublisher) PublishStockUpdate(ctx context.Context, event events.StockUpdateEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestProductService_GetAllProducts_CacheHit(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)

	products := []domain.Product{{ID: 1, Name: "A", Price: 10, Stock: 1}}
	payload, err := json.Marshal(products)
	require.NoError(t, err)

	cache.On("Get", mock.Anything, cacheAllProductsKey).Return(string(payload), nil)

	svc := NewProductService(repo, cache, nil)
	result := svc.GetAllProducts()

	assert.Equal(t, products, result)
	repo.AssertNotCalled(t, "GetAll")
	cache.AssertExpectations(t)
}

func TestProductService_GetAllProducts_CacheMiss(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)

	products := []domain.Product{{ID: 2, Name: "B", Price: 20, Stock: 2}}
	repo.On("GetAll").Return(products)
	cache.On("Get", mock.Anything, cacheAllProductsKey).Return("", errors.New("miss"))
	cache.On("Set", mock.Anything, cacheAllProductsKey, mock.Anything, mock.Anything).Return(nil)

	svc := NewProductService(repo, cache, nil)
	result := svc.GetAllProducts()

	assert.Equal(t, products, result)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestProductService_GetProduct_CacheHit(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)

	product := domain.Product{ID: 1, Name: "A", Price: 10, Stock: 1}
	payload, err := json.Marshal(product)
	require.NoError(t, err)

	cache.On("Get", mock.Anything, productByIDCacheKey(1)).Return(string(payload), nil)

	svc := NewProductService(repo, cache, nil)
	result, err := svc.GetProduct(1)

	require.NoError(t, err)
	assert.Equal(t, &product, result)
	repo.AssertNotCalled(t, "GetByID", mock.Anything)
	cache.AssertExpectations(t)
}

func TestProductService_GetProduct_CacheMiss(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)

	product := &domain.Product{ID: 2, Name: "B", Price: 20, Stock: 2}
	repo.On("GetByID", 2).Return(product, nil)
	cache.On("Get", mock.Anything, productByIDCacheKey(2)).Return("", errors.New("miss"))
	cache.On("Set", mock.Anything, productByIDCacheKey(2), mock.Anything, mock.Anything).Return(nil)

	svc := NewProductService(repo, cache, nil)
	result, err := svc.GetProduct(2)

	require.NoError(t, err)
	assert.Equal(t, product, result)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestProductService_CreateProduct_InvalidPrice(t *testing.T) {
	repo := new(mockProductRepo)
	svc := NewProductService(repo, nil, nil)

	_, err := svc.CreateProduct(domain.Product{Price: 5})
	require.Error(t, err)
	if appErr, ok := err.(*domain.AppError); ok {
		assert.Equal(t, http.StatusUnprocessableEntity, appErr.Code)
	}
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestProductService_CreateProduct_Success(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)

	created := domain.Product{ID: 3, Name: "C", Price: 30, Stock: 3}
	repo.On("Create", mock.Anything).Return(created)
	cache.On("Del", mock.Anything, cacheAllProductsKey).Return(nil)
	cache.On("Del", mock.Anything, productByIDCacheKey(created.ID)).Return(nil)

	svc := NewProductService(repo, cache, nil)
	result, err := svc.CreateProduct(domain.Product{Name: "C", Price: 30, Stock: 3})

	require.NoError(t, err)
	assert.Equal(t, created, result)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestProductService_UpdateProduct(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		input         domain.Product
		setupMocks    func(repo *mockProductRepo, cache *mockCache)
		expectErr     error
		expectResult  *domain.Product
		expectCacheOp bool
	}{
		{
			name:  "not found",
			id:    1,
			input: domain.Product{Name: "X", Price: 10},
			setupMocks: func(repo *mockProductRepo, cache *mockCache) {
				repo.On("GetByID", 1).Return((*domain.Product)(nil), domain.ErrNotFound)
			},
			expectErr:     domain.ErrNotFound,
			expectResult:  nil,
			expectCacheOp: false,
		},
		{
			name:  "success without stock change",
			id:    4,
			input: domain.Product{Name: "D", Price: 40},
			setupMocks: func(repo *mockProductRepo, cache *mockCache) {
				existing := &domain.Product{ID: 4, Name: "D", Price: 40, Stock: 4}
				updated := &domain.Product{ID: 4, Name: "D", Price: 40, Stock: 4}
				repo.On("GetByID", 4).Return(existing, nil)
				repo.On("Update", 4, mock.MatchedBy(func(p domain.Product) bool {
					return p.Stock == 4 && p.Name == "D" && p.Price == 40
				})).Return(updated, nil)
				cache.On("Del", mock.Anything, cacheAllProductsKey).Return(nil)
				cache.On("Del", mock.Anything, productByIDCacheKey(4)).Return(nil)
			},
			expectErr:     nil,
			expectResult:  &domain.Product{ID: 4, Name: "D", Price: 40, Stock: 4},
			expectCacheOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockProductRepo)
			cache := new(mockCache)
			tt.setupMocks(repo, cache)

			svc := NewProductService(repo, cache, nil)
			result, err := svc.UpdateProduct(tt.id, tt.input)

			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
				cache.AssertNotCalled(t, "Del", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectResult, result)
			repo.AssertExpectations(t)
			if tt.expectCacheOp {
				cache.AssertExpectations(t)
			}
		})
	}
}

func TestProductService_AdjustStock(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		stock         int
		setupMocks    func(repo *mockProductRepo, cache *mockCache, publisher *mockStockPublisher)
		expectErr     error
		expectCacheOp bool
	}{
		{
			name:          "invalid stock",
			id:            1,
			stock:         -1,
			setupMocks:    func(repo *mockProductRepo, cache *mockCache, publisher *mockStockPublisher) {},
			expectErr:     domain.ErrInvalidInput,
			expectCacheOp: false,
		},
		{
			name:  "not found",
			id:    2,
			stock: 3,
			setupMocks: func(repo *mockProductRepo, cache *mockCache, publisher *mockStockPublisher) {
				repo.On("GetByID", 2).Return((*domain.Product)(nil), domain.ErrNotFound)
			},
			expectErr:     domain.ErrNotFound,
			expectCacheOp: false,
		},
		{
			name:  "publisher not configured",
			id:    3,
			stock: 10,
			setupMocks: func(repo *mockProductRepo, cache *mockCache, publisher *mockStockPublisher) {
				repo.On("GetByID", 3).Return(&domain.Product{ID: 3}, nil)
			},
			expectErr:     domain.NewError(http.StatusServiceUnavailable, "Kafka producer not configured"),
			expectCacheOp: false,
		},
		{
			name:  "success",
			id:    4,
			stock: 8,
			setupMocks: func(repo *mockProductRepo, cache *mockCache, publisher *mockStockPublisher) {
				repo.On("GetByID", 4).Return(&domain.Product{ID: 4}, nil)
				publisher.On("PublishStockUpdate", mock.Anything, mock.MatchedBy(func(e events.StockUpdateEvent) bool {
					return e.ProductID == 4 && e.Stock == 8 && e.EventType == events.EventTypeProductStockUpdated
				})).Return(nil)
				cache.On("Del", mock.Anything, cacheAllProductsKey).Return(nil)
				cache.On("Del", mock.Anything, productByIDCacheKey(4)).Return(nil)
			},
			expectErr:     nil,
			expectCacheOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockProductRepo)
			cache := new(mockCache)
			publisher := new(mockStockPublisher)
			tt.setupMocks(repo, cache, publisher)

			var pub events.StockEventPublisher
			if tt.name != "publisher not configured" {
				pub = publisher
			}

			svc := NewProductService(repo, cache, pub)
			err := svc.AdjustStock(tt.id, tt.stock)

			if tt.expectErr != nil {
				assert.Error(t, err)
				cache.AssertNotCalled(t, "Del", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			repo.AssertExpectations(t)
			if tt.expectCacheOp {
				cache.AssertExpectations(t)
				publisher.AssertExpectations(t)
			}
		})
	}
}

func TestProductService_UpdateProduct_IgnoresStockAndDoesNotPublishEvent(t *testing.T) {
	repo := new(mockProductRepo)
	cache := new(mockCache)
	publisher := new(mockStockPublisher)

	existing := &domain.Product{ID: 10, Name: "S", Price: 100, Stock: 5}
	updated := &domain.Product{ID: 10, Name: "S", Price: 100, Stock: 5}

	repo.On("GetByID", 10).Return(existing, nil)
	repo.On("Update", 10, mock.MatchedBy(func(p domain.Product) bool {
		return p.Name == "S" && p.Price == 100 && p.Stock == existing.Stock
	})).Return(updated, nil)
	cache.On("Del", mock.Anything, cacheAllProductsKey).Return(nil)
	cache.On("Del", mock.Anything, productByIDCacheKey(10)).Return(nil)

	svc := NewProductService(repo, cache, publisher)
	result, err := svc.UpdateProduct(10, domain.Product{Name: "S", Price: 100, Stock: 99})

	require.NoError(t, err)
	assert.Equal(t, updated, result)
	repo.AssertExpectations(t)
	publisher.AssertNotCalled(t, "PublishStockUpdate", mock.Anything, mock.Anything)
	cache.AssertExpectations(t)
}

func TestProductService_DeleteProduct(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMocks    func(repo *mockProductRepo, cache *mockCache)
		expectErr     error
		expectCacheOp bool
	}{
		{
			name: "not found",
			id:   1,
			setupMocks: func(repo *mockProductRepo, cache *mockCache) {
				repo.On("Delete", 1).Return(domain.ErrNotFound)
			},
			expectErr:     domain.ErrNotFound,
			expectCacheOp: false,
		},
		{
			name: "success",
			id:   5,
			setupMocks: func(repo *mockProductRepo, cache *mockCache) {
				repo.On("Delete", 5).Return(nil)
				cache.On("Del", mock.Anything, cacheAllProductsKey).Return(nil)
				cache.On("Del", mock.Anything, productByIDCacheKey(5)).Return(nil)
			},
			expectErr:     nil,
			expectCacheOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockProductRepo)
			cache := new(mockCache)
			tt.setupMocks(repo, cache)

			svc := NewProductService(repo, cache, nil)
			err := svc.DeleteProduct(tt.id)

			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
				cache.AssertNotCalled(t, "Del", mock.Anything, mock.Anything)
				return
			}

			require.NoError(t, err)
			repo.AssertExpectations(t)
			if tt.expectCacheOp {
				cache.AssertExpectations(t)
			}
		})
	}
}
