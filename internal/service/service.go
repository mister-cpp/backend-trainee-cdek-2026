package service

// Содержит основную структуру сервиса и методы для работы с пользователями
// (регистрация, аутентификация и работа с JWT).

import (
	"context"
	"time"

	"cdek/internal/domain"
	"cdek/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      repository.PostgresRepository
	jwtSecret string
}

func NewService(repo repository.PostgresRepository, secret string) *Service {
	return &Service{repo: repo, jwtSecret: secret}
}

func (s *Service) Register(ctx context.Context, req domain.RegisterReq) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	_, err = s.repo.CreateUser(ctx, domain.User{
		Email:    req.Email,
		Password: string(hash),
	})

	return err
}

func (s *Service) Login(ctx context.Context, req domain.RegisterReq) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)

	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}
