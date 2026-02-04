package rbac

import (
	"fmt"
)

// Role constants
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// ValidRoles contains all valid role names
var ValidRoles = []string{RoleSuperAdmin, RoleAdmin, RoleUser}

// IsValidRole checks if a role name is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if a user (by role) has permission to perform an action on a resource.
func HasPermission(role, resource, action string) (bool, error) {
	enforcer, err := Instance()
	if err != nil {
		return false, err
	}

	ok, err := enforcer.Enforce(role, resource, action)
	if err != nil {
		return false, fmt.Errorf("enforce permission: %w", err)
	}
	return ok, nil
}

// HasPermissionByUserRole checks permission for a user's role.
// This is a convenience wrapper that takes a role string directly.
func HasPermissionByUserRole(userRole, resource, action string) bool {
	ok, err := HasPermission(userRole, resource, action)
	if err != nil {
		return false
	}
	return ok
}

// GetRolePermissions returns all permissions for a role (including inherited).
func GetRolePermissions(role string) ([][]string, error) {
	enforcer, err := Instance()
	if err != nil {
		return nil, err
	}

	permissions, err := enforcer.GetPermissionsForUser(role)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}
	return permissions, nil
}

// GetImplicitPermissions returns all permissions for a role including inherited ones.
func GetImplicitPermissions(role string) ([][]string, error) {
	enforcer, err := Instance()
	if err != nil {
		return nil, err
	}

	permissions, err := enforcer.GetImplicitPermissionsForUser(role)
	if err != nil {
		return nil, fmt.Errorf("get implicit permissions: %w", err)
	}
	return permissions, nil
}

// AddPermission adds a permission policy.
func AddPermission(role, resource, action string) error {
	enforcer, err := Instance()
	if err != nil {
		return err
	}

	_, err = enforcer.AddPolicy(role, resource, action)
	if err != nil {
		return fmt.Errorf("add policy: %w", err)
	}
	return nil
}

// RemovePermission removes a permission policy.
func RemovePermission(role, resource, action string) error {
	enforcer, err := Instance()
	if err != nil {
		return err
	}

	_, err = enforcer.RemovePolicy(role, resource, action)
	if err != nil {
		return fmt.Errorf("remove policy: %w", err)
	}
	return nil
}

// AddRoleInheritance adds role inheritance (child inherits from parent).
func AddRoleInheritance(child, parent string) error {
	enforcer, err := Instance()
	if err != nil {
		return err
	}

	_, err = enforcer.AddGroupingPolicy(child, parent)
	if err != nil {
		return fmt.Errorf("add grouping policy: %w", err)
	}
	return nil
}

// GetRoleHierarchy returns the roles that a role inherits from.
func GetRoleHierarchy(role string) ([]string, error) {
	enforcer, err := Instance()
	if err != nil {
		return nil, err
	}

	roles, err := enforcer.GetImplicitRolesForUser(role)
	if err != nil {
		return nil, fmt.Errorf("get implicit roles: %w", err)
	}
	return roles, nil
}

// HasRole checks if a role inherits from or is equal to the target role.
func HasRole(userRole, targetRole string) bool {
	if userRole == targetRole {
		return true
	}

	enforcer, err := Instance()
	if err != nil {
		return false
	}

	roles, err := enforcer.GetImplicitRolesForUser(userRole)
	if err != nil {
		return false
	}

	for _, r := range roles {
		if r == targetRole {
			return true
		}
	}
	return false
}

// IsAdmin checks if a role has admin or super_admin privileges.
func IsAdmin(role string) bool {
	return role == RoleAdmin || role == RoleSuperAdmin
}

// IsSuperAdmin checks if a role is super_admin.
func IsSuperAdmin(role string) bool {
	return role == RoleSuperAdmin
}
