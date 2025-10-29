package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	libauth "github.com/austiecodes/dws/internal/lib/auth"
	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already registered")
)

type AuthService struct {
}

var Auth = &AuthService{}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*libdb.User, error) {
	user, err := repository.Users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !libauth.ComparePassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := repository.Users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now
	return user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uint) (*libdb.User, error) {
	user, err := repository.Users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*libdb.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)
	if email == "" || displayName == "" {
		return nil, fmt.Errorf("email and display name required")
	}

	existing, err := repository.Users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check duplicate email: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := libauth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &libdb.User{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: hash,
		LastLoginAt:  &now,
	}

	if err := repository.Users.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := repository.Users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now
	return user, nil
}
