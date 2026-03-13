package service

import (
	"testing"
	"time"

	"gin-quickstart/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) Create(u domain.User) (domain.User, error) {
	args := m.Called(u)
	if out, ok := args.Get(0).(domain.User); ok {
		return out, args.Error(1)
	}
	return domain.User{}, args.Error(1)
}

func TestAuthService_LoginSuccess(t *testing.T) {
	password := "secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &domain.User{ID: 7, Username: "alice", PasswordHash: string(hash)}
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "alice").Return(user, nil)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	token, expiry, err := svc.Login("alice", password)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, time.Hour, expiry)

	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("jwt-secret"), nil
	})
	require.NoError(t, err)
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	require.True(t, ok)
	assert.Equal(t, "7", claims.Subject)

	repo.AssertExpectations(t)
}

func TestAuthService_LoginUserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "missing").Return((*domain.User)(nil), domain.ErrNotFound)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, _, err := svc.Login("missing", "pw")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	repo.AssertExpectations(t)
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &domain.User{ID: 1, Username: "bob", PasswordHash: string(hash)}
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "bob").Return(user, nil)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, _, err = svc.Login("bob", "wrong")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	repo.AssertExpectations(t)
}

func TestAuthService_RegisterInvalidInput(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, err := svc.Register("", "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	repo.AssertNotCalled(t, "GetByUsername", mock.Anything)
}

func TestAuthService_RegisterConflict(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "alice").Return(&domain.User{ID: 1}, nil)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, err := svc.Register("alice", "pw")
	assert.ErrorIs(t, err, domain.ErrConflict)

	repo.AssertExpectations(t)
}

func TestAuthService_RegisterGetByUsernameError(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "alice").Return((*domain.User)(nil), domain.ErrInternal)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, err := svc.Register("alice", "pw")
	assert.ErrorIs(t, err, domain.ErrInternal)

	repo.AssertExpectations(t)
}

func TestAuthService_RegisterCreateError(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "alice").Return((*domain.User)(nil), domain.ErrNotFound)
	repo.On("Create", mock.Anything).Return(domain.User{}, domain.ErrInternal)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	_, err := svc.Register("alice", "pw")
	assert.ErrorIs(t, err, domain.ErrInternal)

	repo.AssertExpectations(t)
}

func TestAuthService_RegisterSuccess(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("GetByUsername", "alice").Return((*domain.User)(nil), domain.ErrNotFound)
	repo.On("Create", mock.MatchedBy(func(u domain.User) bool {
		if u.Username != "alice" {
			return false
		}
		return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("pw")) == nil
	})).Return(domain.User{ID: 9, Username: "alice"}, nil)

	svc := NewAuthService(repo, "jwt-secret", time.Hour)

	created, err := svc.Register("alice", "pw")
	require.NoError(t, err)
	assert.Equal(t, 9, created.ID)
	assert.Equal(t, "alice", created.Username)

	repo.AssertExpectations(t)
}
