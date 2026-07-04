package bootstrap

import (
	"log/slog"

	permApp "skykin-platform/internal/permissions/application"
	permInfra "skykin-platform/internal/permissions/infrastructure"
	permHTTP "skykin-platform/internal/permissions/interfaces/http"
	"skykin-platform/internal/platform/messaging"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewPermissionSystem wires the permissions module (composition root).
func NewPermissionSystem(
	db *gorm.DB,
	rdb *redis.Client,
	bus *messaging.Bus,
	logger *slog.Logger,
) (*permApp.PermissionChecker, *permHTTP.Handler) {
	if logger == nil {
		logger = slog.Default()
	}

	permissionRepo := permInfra.NewPermissionRepository(db)
	roleRepo := permInfra.NewRoleRepository(db)

	var cache *permInfra.RedisPermissionCache
	var memCache *permInfra.InMemoryPermissionCache
	if rdb != nil {
		cache = permInfra.NewRedisPermissionCache(rdb, logger)
		logger.Info("permissions cache: redis")
	} else {
		memCache = permInfra.NewInMemoryPermissionCache()
		logger.Info("permissions cache: in-memory")
	}

	assignUC := permApp.NewAssignPermissionUseCase(roleRepo, permissionRepo, pickCache(cache, memCache), bus, logger)
	revokeUC := permApp.NewRevokePermissionUseCase(roleRepo, permissionRepo, pickCache(cache, memCache), bus, logger)
	createUC := permApp.NewCreateRoleUseCase(roleRepo, bus, logger)
	listRolesUC := permApp.NewListRolesUseCase(roleRepo)
	listPermsUC := permApp.NewListPermissionsUseCase(permissionRepo)

	handler := permHTTP.NewHandler(listPermsUC, listRolesUC, createUC, assignUC, revokeUC)
	checker := permApp.NewPermissionChecker(roleRepo, pickCheckerCache(cache, memCache), logger)
	return checker, handler
}

type cacheInvalidator interface {
	Invalidate(roleName string)
}

func pickCache(redis *permInfra.RedisPermissionCache, mem *permInfra.InMemoryPermissionCache) cacheInvalidator {
	if redis != nil {
		return redis
	}
	return mem
}

func pickCheckerCache(redis *permInfra.RedisPermissionCache, mem *permInfra.InMemoryPermissionCache) interface {
	permApp.PermissionCacheReader
	permApp.CacheWriter
} {
	if redis != nil {
		return redis
	}
	return mem
}
