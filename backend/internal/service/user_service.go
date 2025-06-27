package service

import (
	"context"
	"fmt"

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

func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(context.Background(), username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	return jwt.GenerateToken(user.ID.String())
}
