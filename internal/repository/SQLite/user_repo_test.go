package repository

import (
	"testing"

	"gin-quickstart/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(domain.User{Username: "Test_Name", PasswordHash: "hash"})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	found, err := repo.GetByUsername("Test_Name")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Test_Name", found.Username)
	assert.Equal(t, "hash", found.PasswordHash)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	user, err := repo.GetByUsername("missing")
	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.Create(domain.User{Username: "dup", PasswordHash: "hash"})
	require.NoError(t, err)

	_, err = repo.Create(domain.User{Username: "dup", PasswordHash: "hash2"})
	assert.ErrorIs(t, err, domain.ErrInternal)
}
