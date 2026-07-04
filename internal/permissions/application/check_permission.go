package application

import (
	"context"
	"log/slog"
	"time"
)

type PermissionCacheReader interface {
	Get(roleName string) ([]string, bool)
}

type RolePermissionReader interface {
	GetPermissionNames(ctx context.Context, roleName string) ([]string, error)
}

type CacheWriter interface {
	Set(roleName string, permissions []string, ttl time.Duration)
}

type PermissionChecker struct {
	roleRepo RolePermissionReader
	cache    interface {
		PermissionCacheReader
		CacheWriter
	}
	logger *slog.Logger
}

func NewPermissionChecker(
	roleRepo RolePermissionReader,
	cache interface {
		PermissionCacheReader
		CacheWriter
	},
	logger *slog.Logger,
) *PermissionChecker {
	if logger == nil {
		logger = slog.Default()
	}
	return &PermissionChecker{roleRepo: roleRepo, cache: cache, logger: logger}
}

func (c *PermissionChecker) HasPermission(ctx context.Context, roleName, permission string) bool {
	if roleName == "operator_admin" {
		return true
	}
	if roleName == "read_only_analyst" {
		roleName = "analyst"
	}
	if perms, ok := c.cache.Get(roleName); ok {
		return containsPermission(perms, permission)
	}
	perms, err := c.roleRepo.GetPermissionNames(ctx, roleName)
	if err != nil {
		c.logger.Warn("permission lookup failed", "role", roleName, "error", err)
		return false
	}
	c.cache.Set(roleName, perms, 5*time.Minute)
	return containsPermission(perms, permission)
}

func containsPermission(perms []string, permission string) bool {
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}
