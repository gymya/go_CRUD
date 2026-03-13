package repository

import (
	"errors"
	"gin-quickstart/internal/domain"

	"gorm.io/gorm"
)

type sqliteRepo struct {
	db *gorm.DB
}

// NewSqliteRepository 初始化並自動建立資料表 (AutoMigrate)
func NewSqliteRepository(db *gorm.DB) domain.ProductRepository {
	// 自動根據結構建立或更新 Table
	db.AutoMigrate(&domain.Product{})
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) GetAll() []domain.Product {
	var products []domain.Product
	r.db.Find(&products)
	return products
}

func (r *sqliteRepo) GetByID(id int) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.ErrInternal
	}
	return &p, nil
}

func (r *sqliteRepo) Create(p domain.Product) domain.Product {
	r.db.Create(&p)
	return p
}

func (r *sqliteRepo) Update(id int, p domain.Product) (*domain.Product, error) {
	var existing domain.Product
	if err := r.db.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.ErrInternal
	}

	// 更新除了 ID 以外的欄位
	p.ID = id
	if err := r.db.Save(&p).Error; err != nil {
		return nil, domain.ErrInternal
	}
	return &p, nil
}

func (r *sqliteRepo) Delete(id int) error {
	result := r.db.Delete(&domain.Product{}, id)
	if result.Error != nil {
		return domain.ErrInternal
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
