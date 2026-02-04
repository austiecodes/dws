package services

import (
	"context"
	"errors"
	"fmt"

	libdb "github.com/austiecodes/dws/internal/lib/db"
	"github.com/austiecodes/dws/internal/platform/repository"
	"gorm.io/gorm"
)

var (
	ErrForbiddenUserManagement = errors.New("admin permission required")
	ErrLastAdmin               = errors.New("cannot remove the last admin user")
)

type UserService struct{}

var UsersService = &UserService{}

func (s *UserService) requireAdmin(ctx context.Context, userID uint) error {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	if user == nil || !user.IsAdmin {
		return ErrForbiddenUserManagement
	}
	return nil
}

func (s *UserService) List(ctx context.Context, requesterID uint) ([]libdb.User, error) {
	if err := s.requireAdmin(ctx, requesterID); err != nil {
		return nil, err
	}
	users, err := repository.UserList(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *UserService) SetAdmin(ctx context.Context, requesterID, targetID uint, isAdmin bool) error {
	if err := s.requireAdmin(ctx, requesterID); err != nil {
		return err
	}

	target, err := repository.UserFindByID(ctx, nil, targetID)
	if err != nil {
		return fmt.Errorf("lookup target user: %w", err)
	}
	if target == nil {
		return ErrUserNotFound
	}

	if !isAdmin && target.IsAdmin {
		adminCount, err := repository.UserCountAdmins(ctx, nil)
		if err != nil {
			return fmt.Errorf("count admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	if err := repository.UserSetAdmin(ctx, nil, targetID, isAdmin); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("update admin flag: %w", err)
	}
	return nil
}
