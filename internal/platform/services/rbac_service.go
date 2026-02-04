package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/austiecodes/dws/internal/lib/rbac"
	"github.com/austiecodes/dws/internal/platform/repository"
)

var (
	ErrInvalidRole       = errors.New("invalid role")
	ErrLastSuperAdmin    = errors.New("cannot remove the last super_admin")
	ErrCannotAssignRoles = errors.New("permission denied: cannot assign roles")
	ErrCannotDowngrade   = errors.New("cannot downgrade your own role")
)

type RBACService struct{}

var RBAC = &RBACService{}

// GetUserRole returns the role of a user.
func (s *RBACService) GetUserRole(ctx context.Context, userID uint) (string, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return "", fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return "", ErrUserNotFound
	}
	return user.Role, nil
}

// AssignRole assigns a new role to a target user.
// Only super_admin can assign roles. Cannot remove the last super_admin.
func (s *RBACService) AssignRole(ctx context.Context, requesterID, targetID uint, newRole string) error {
	if !rbac.IsValidRole(newRole) {
		return ErrInvalidRole
	}

	// Get requester
	requester, err := repository.UserFindByID(ctx, nil, requesterID)
	if err != nil {
		return fmt.Errorf("lookup requester: %w", err)
	}
	if requester == nil {
		return ErrUserNotFound
	}

	// Only super_admin can assign roles
	if !rbac.IsSuperAdmin(requester.Role) {
		return ErrCannotAssignRoles
	}

	// Get target user
	target, err := repository.UserFindByID(ctx, nil, targetID)
	if err != nil {
		return fmt.Errorf("lookup target: %w", err)
	}
	if target == nil {
		return ErrUserNotFound
	}

	// Prevent downgrading self
	if requesterID == targetID && newRole != rbac.RoleSuperAdmin {
		return ErrCannotDowngrade
	}

	// Prevent removing the last super_admin
	if target.Role == rbac.RoleSuperAdmin && newRole != rbac.RoleSuperAdmin {
		count, err := s.countSuperAdmins(ctx)
		if err != nil {
			return fmt.Errorf("count super_admins: %w", err)
		}
		if count <= 1 {
			return ErrLastSuperAdmin
		}
	}

	// Update the role
	if err := repository.UserUpdateRole(ctx, nil, targetID, newRole); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	// Also update is_admin for backwards compatibility
	isAdmin := rbac.IsAdmin(newRole)
	if err := repository.UserSetAdmin(ctx, nil, targetID, isAdmin); err != nil {
		return fmt.Errorf("update admin flag: %w", err)
	}

	return nil
}

// countSuperAdmins returns the number of users with super_admin role.
func (s *RBACService) countSuperAdmins(ctx context.Context) (int64, error) {
	return repository.UserCountByRole(ctx, nil, rbac.RoleSuperAdmin)
}

// ListRoles returns all available roles.
func (s *RBACService) ListRoles() []string {
	return rbac.ValidRoles
}

// GetUserPermissions returns all effective permissions for a user based on their role.
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([][]string, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	permissions, err := rbac.GetImplicitPermissions(user.Role)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}
	return permissions, nil
}

// CanManageTask checks if a user can manage a specific task (owner or admin).
func (s *RBACService) CanManageTask(ctx context.Context, userID, taskID uint) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	// Check if user has manage_any permission
	if rbac.HasPermissionByUserRole(user.Role, "tasks", "manage_any") {
		return true, nil
	}

	// Check ownership
	task, err := repository.TaskGetByID(ctx, nil, taskID, false)
	if err != nil {
		return false, fmt.Errorf("lookup task: %w", err)
	}
	if task == nil {
		return false, nil
	}

	return task.UserID == userID, nil
}

// CanManageContainer checks if a user can manage a specific container (owner or admin).
func (s *RBACService) CanManageContainer(ctx context.Context, userID uint, containerUUID string) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	// Check if user has manage_any permission
	if rbac.HasPermissionByUserRole(user.Role, "containers", "manage_any") {
		return true, nil
	}

	// Check ownership
	container, err := repository.ContainerGetByUUID(ctx, nil, containerUUID)
	if err != nil {
		return false, fmt.Errorf("lookup container: %w", err)
	}
	if container == nil {
		return false, nil
	}

	return container.UserID == userID, nil
}

// CanManageImages checks if a user can manage images (create/update/delete).
func (s *RBACService) CanManageImages(ctx context.Context, userID uint) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	return rbac.HasPermissionByUserRole(user.Role, "images", "create"), nil
}

// CanViewAllTasks checks if a user can view all tasks (admin+).
func (s *RBACService) CanViewAllTasks(ctx context.Context, userID uint) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	return rbac.HasPermissionByUserRole(user.Role, "tasks", "view_all"), nil
}

// CanViewAllContainers checks if a user can view all containers (admin+).
func (s *RBACService) CanViewAllContainers(ctx context.Context, userID uint) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	return rbac.HasPermissionByUserRole(user.Role, "containers", "view_all"), nil
}

// HasPermission checks if a user has a specific permission.
func (s *RBACService) HasPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	user, err := repository.UserFindByID(ctx, nil, userID)
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}
	if user == nil {
		return false, nil
	}

	return rbac.HasPermissionByUserRole(user.Role, resource, action), nil
}

// GetEffectivePermissionsForRole returns permissions for a specific role.
func (s *RBACService) GetEffectivePermissionsForRole(role string) ([][]string, error) {
	return rbac.GetImplicitPermissions(role)
}
