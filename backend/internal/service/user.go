package service

import (
	"context"
	"fmt"

	"github.com/meta-boy/mech-alligator/internal/domain/user"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
	"github.com/meta-boy/mech-alligator/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *postgres.UserRepository
}

func NewUserService(userRepo *postgres.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(username, password string) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &user.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(context.Background(), u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(context.Background(), username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	return jwt.GenerateToken(user.ID)
}
