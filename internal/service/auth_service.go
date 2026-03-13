package service

import (
	"errors"
	"gin-quickstart/internal/domain"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username, password string) (token string, expiresIn time.Duration, err error)
	Register(username, password string) (domain.User, error)
}

type authService struct {
	userRepo  domain.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewAuthService(userRepo domain.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Login(username, password string) (string, time.Duration, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil || user == nil {
		return "", 0, domain.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, domain.ErrUnauthorized
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.ID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, domain.ErrInternal
	}
	return signed, s.jwtExpiry, nil
}

func (s *authService) Register(username, password string) (domain.User, error) {
	if username == "" || password == "" {
		return domain.User{}, domain.ErrInvalidInput
	}
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return domain.User{}, domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, domain.ErrInternal
	}

	created, err := s.userRepo.Create(domain.User{
		Username:     username,
		PasswordHash: string(hash),
	})
	if err != nil {
		return domain.User{}, err
	}
	return created, nil
}
