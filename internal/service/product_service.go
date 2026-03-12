package service

import (
	"gin-quickstart/internal/domain"
	"net/http"
)

// ProductService 定義業務邏輯介面
type ProductService interface {
	GetAllProducts() []domain.Product
	GetProduct(id int) (*domain.Product, error)
	CreateProduct(p domain.Product) (domain.Product, error)
	UpdateProduct(id int, p domain.Product) (*domain.Product, error)
	DeleteProduct(id int) error
}

// 實作結構體，它需要依賴 Repo
type productService struct {
	repo domain.ProductRepository
}

// NewProductService 初始化 Service
func NewProductService(repo domain.ProductRepository) ProductService {
	return &productService{repo: repo}
}

// --- 實作各個方法 ---

func (s *productService) GetAllProducts() []domain.Product {
	return s.repo.GetAll()
}

func (s *productService) GetProduct(id int) (*domain.Product, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) CreateProduct(p domain.Product) (domain.Product, error) {
	// 禁止建立價格低於 10 元的商品
	if p.Price < 10 {
		return domain.Product{}, domain.NewError(http.StatusUnprocessableEntity, "商品價格太低")
	}
	return s.repo.Create(p), nil
}

func (s *productService) UpdateProduct(id int, p domain.Product) (*domain.Product, error) {
	return s.repo.Update(id, p)
}

func (s *productService) DeleteProduct(id int) error {
	return s.repo.Delete(id)
}
