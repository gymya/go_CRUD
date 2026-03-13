package repository

import (
	"testing"

	"gin-quickstart/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository_CreateAndGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	created := repo.Create(domain.Product{Name: "A", Price: 10, Stock: 1})
	assert.NotZero(t, created.ID)

	found, err := repo.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "A", found.Name)
	assert.Equal(t, 10.0, found.Price)
	assert.Equal(t, 1, found.Stock)
}

func TestProductRepository_GetAll(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	_ = repo.Create(domain.Product{Name: "A", Price: 10, Stock: 1})
	_ = repo.Create(domain.Product{Name: "B", Price: 20, Stock: 2})

	all := repo.GetAll()
	assert.Len(t, all, 2)
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	p, err := repo.GetByID(999)
	assert.Nil(t, p)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestProductRepository_Update_Success(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	created := repo.Create(domain.Product{Name: "A", Price: 10, Stock: 1})

	updated, err := repo.Update(created.ID, domain.Product{Name: "A2", Price: 15, Stock: 3})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "A2", updated.Name)
	assert.Equal(t, 15.0, updated.Price)
	assert.Equal(t, 3, updated.Stock)
}

func TestProductRepository_Update_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	updated, err := repo.Update(123, domain.Product{Name: "X", Price: 10, Stock: 1})
	assert.Nil(t, updated)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestProductRepository_Delete_Success(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	created := repo.Create(domain.Product{Name: "A", Price: 10, Stock: 1})

	err := repo.Delete(created.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(created.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestProductRepository_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewSqliteRepository(db)

	err := repo.Delete(999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
