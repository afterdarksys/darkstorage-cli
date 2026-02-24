package apikeys

import (
	"fmt"
	"strings"
	"time"
)

// PermissionChecker validates permissions
type PermissionChecker struct {
	key *APIKey
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(key *APIKey) *PermissionChecker {
	return &PermissionChecker{key: key}
}

// HasPermission checks if the API key has a specific permission
func (pc *PermissionChecker) HasPermission(required string) bool {
	// Admin has all permissions
	if pc.hasAdminPermission() {
		return true
	}

	// Check exact match
	for _, perm := range pc.key.Permissions {
		if perm == required {
			return true
		}

		// Check wildcard (e.g., "storage:*" grants "storage:read")
		if strings.HasSuffix(perm, ":*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}

	return false
}

// HasAnyPermission checks if the API key has any of the given permissions
func (pc *PermissionChecker) HasAnyPermission(permissions ...string) bool {
	for _, perm := range permissions {
		if pc.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the API key has all of the given permissions
func (pc *PermissionChecker) HasAllPermissions(permissions ...string) bool {
	for _, perm := range permissions {
		if !pc.HasPermission(perm) {
			return false
		}
	}
	return true
}

// hasAdminPermission checks if the key has admin:* permission
func (pc *PermissionChecker) hasAdminPermission() bool {
	for _, perm := range pc.key.Permissions {
		if perm == PermAdminAll {
			return true
		}
	}
	return false
}

// RequirePermission returns an error if permission is not granted
func (pc *PermissionChecker) RequirePermission(required string) error {
	if !pc.HasPermission(required) {
		return fmt.Errorf("permission denied: %s required", required)
	}
	return nil
}

// GetPermissions returns all permissions granted to this key
func (pc *PermissionChecker) GetPermissions() []string {
	if pc.hasAdminPermission() {
		return AllPermissions()
	}
	return pc.key.Permissions
}

// IsActive checks if the key is active and not expired
func (pc *PermissionChecker) IsActive() bool {
	if pc.key.Status != StatusActive {
		return false
	}

	if pc.key.ExpiresAt != nil && pc.key.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// CheckIPRestriction validates if the given IP is allowed
func (pc *PermissionChecker) CheckIPRestriction(ip string) error {
	// No restrictions = allow all
	if len(pc.key.IPRestrictions) == 0 {
		return nil
	}

	// Check against restrictions
	for _, allowedIP := range pc.key.IPRestrictions {
		if matchIPRestriction(ip, allowedIP) {
			return nil
		}
	}

	return fmt.Errorf("IP %s not allowed", ip)
}

// matchIPRestriction checks if IP matches restriction (supports CIDR)
func matchIPRestriction(ip, restriction string) bool {
	// Exact match
	if ip == restriction {
		return true
	}

	// TODO: Implement CIDR matching
	// For now, just exact match
	return false
}
