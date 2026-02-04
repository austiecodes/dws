package rbac

import (
	"fmt"
)

// toInterfaceSlice converts []string to []interface{} for Casbin API compatibility
func toInterfaceSlice(s []string) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// SeedDefaultPolicies seeds the default role hierarchy and permissions.
// This function is idempotent - safe to run multiple times.
func SeedDefaultPolicies() error {
	enforcer, err := Instance()
	if err != nil {
		return fmt.Errorf("get enforcer: %w", err)
	}

	// Define role hierarchy: admin inherits from user, super_admin inherits from admin
	roleHierarchy := [][]string{
		{RoleAdmin, RoleUser},
		{RoleSuperAdmin, RoleAdmin},
	}

	for _, rh := range roleHierarchy {
		exists, err := enforcer.HasGroupingPolicy(toInterfaceSlice(rh)...)
		if err != nil {
			return fmt.Errorf("check grouping policy: %w", err)
		}
		if !exists {
			if _, err := enforcer.AddGroupingPolicy(toInterfaceSlice(rh)...); err != nil {
				return fmt.Errorf("add grouping policy %v: %w", rh, err)
			}
		}
	}

	// Define permissions for each role
	userPermissions := [][]string{
		// Images
		{RoleUser, "images", "list"},
		// Containers
		{RoleUser, "containers", "create"},
		{RoleUser, "containers", "view_own"},
		{RoleUser, "containers", "manage_own"},
		// Tasks
		{RoleUser, "tasks", "create"},
		{RoleUser, "tasks", "view_own"},
		{RoleUser, "tasks", "manage_own"},
	}

	adminPermissions := [][]string{
		// Images (admin can manage)
		{RoleAdmin, "images", "create"},
		{RoleAdmin, "images", "update"},
		{RoleAdmin, "images", "delete"},
		// Containers (admin can view/manage all)
		{RoleAdmin, "containers", "view_all"},
		{RoleAdmin, "containers", "manage_any"},
		// Tasks (admin can view/manage all)
		{RoleAdmin, "tasks", "view_all"},
		{RoleAdmin, "tasks", "manage_any"},
		// Users (admin can view)
		{RoleAdmin, "users", "list"},
		{RoleAdmin, "users", "view"},
	}

	superAdminPermissions := [][]string{
		// Users (super_admin can manage roles)
		{RoleSuperAdmin, "users", "assign_roles"},
		{RoleSuperAdmin, "users", "delete"},
	}

	allPermissions := append(userPermissions, adminPermissions...)
	allPermissions = append(allPermissions, superAdminPermissions...)

	for _, p := range allPermissions {
		exists, err := enforcer.HasPolicy(toInterfaceSlice(p)...)
		if err != nil {
			return fmt.Errorf("check policy: %w", err)
		}
		if !exists {
			if _, err := enforcer.AddPolicy(toInterfaceSlice(p)...); err != nil {
				return fmt.Errorf("add policy %v: %w", p, err)
			}
		}
	}

	return nil
}

// ClearAllPolicies removes all policies and grouping policies.
// Use with caution - primarily for testing.
func ClearAllPolicies() error {
	enforcer, err := Instance()
	if err != nil {
		return err
	}

	// Clear all policies
	enforcer.ClearPolicy()

	// Save to adapter
	if err := enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("save policy: %w", err)
	}

	return nil
}

// ResetToDefaults clears all policies and re-seeds defaults.
// Use with caution - primarily for testing or recovery.
func ResetToDefaults() error {
	if err := ClearAllPolicies(); err != nil {
		return fmt.Errorf("clear policies: %w", err)
	}

	if err := SeedDefaultPolicies(); err != nil {
		return fmt.Errorf("seed defaults: %w", err)
	}

	return nil
}
