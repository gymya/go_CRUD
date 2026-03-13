package repository

import (
	"errors"
	"gin-quickstart/internal/domain"

	"gorm.io/gorm"
)

type sqliteUserRepo struct {
	db *gorm.DB
}

// NewUserRepository 初始化並自動建立使用者資料表 (AutoMigrate)
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	db.AutoMigrate(&domain.User{})
	return &sqliteUserRepo{db: db}
}

func (r *sqliteUserRepo) GetByUsername(username string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, domain.ErrInternal
	}
	return &u, nil
}

func (r *sqliteUserRepo) Create(u domain.User) (domain.User, error) {
	if err := r.db.Create(&u).Error; err != nil {
		return domain.User{}, domain.ErrInternal
	}
	return u, nil
}
