package factories

import (
	"context"
	"fmt"

	"github.com/anasamu/microservices-library-go/cache"
	"github.com/anasamu/microservices-library-go/core/interfaces"
)

// CacheType represents the type of cache
type CacheType string

const (
	CacheTypeRedis      CacheType = "redis"
	CacheTypeMemcache   CacheType = "memcache"
	CacheTypeInMemory   CacheType = "inmemory"
	CacheTypeMultiLevel CacheType = "multilevel"
)

// CacheFactory creates cache instances
type CacheFactory struct {
	configs map[CacheType]interface{}
}

// NewCacheFactory creates a new cache factory
func NewCacheFactory() *CacheFactory {
	return &CacheFactory{
		configs: make(map[CacheType]interface{}),
	}
}

// RegisterConfig registers a cache configuration
func (f *CacheFactory) RegisterConfig(cacheType CacheType, config interface{}) {
	f.configs[cacheType] = config
}

// CreateCache creates a cache instance based on type
func (f *CacheFactory) CreateCache(ctx context.Context, cacheType CacheType) (interfaces.Cache, error) {
	config, exists := f.configs[cacheType]
	if !exists {
		return nil, fmt.Errorf("no configuration found for cache type: %s", cacheType)
	}

	switch cacheType {
	case CacheTypeRedis:
		redisConfig, ok := config.(*cache.RedisConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for Redis cache")
		}
		return cache.NewRedisCache(redisConfig)

	case CacheTypeMemcache:
		memcacheConfig, ok := config.(*cache.MemcacheConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for Memcache")
		}
		return cache.NewMemcacheCache(memcacheConfig)

	case CacheTypeInMemory:
		inMemoryConfig, ok := config.(*cache.InMemoryConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for InMemory cache")
		}
		return cache.NewInMemoryCache(inMemoryConfig)

	case CacheTypeMultiLevel:
		multiLevelConfig, ok := config.(*cache.MultiLevelConfig)
		if !ok {
			return nil, fmt.Errorf("invalid configuration type for MultiLevel cache")
		}
		return cache.NewMultiLevelCache(multiLevelConfig)

	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}
}

// CreateMultipleCaches creates multiple cache instances
func (f *CacheFactory) CreateMultipleCaches(ctx context.Context, cacheTypes []CacheType) (map[CacheType]interfaces.Cache, error) {
	caches := make(map[CacheType]interfaces.Cache)

	for _, cacheType := range cacheTypes {
		cache, err := f.CreateCache(ctx, cacheType)
		if err != nil {
			// Close already created caches
			for _, createdCache := range caches {
				createdCache.Close()
			}
			return nil, fmt.Errorf("failed to create cache %s: %w", cacheType, err)
		}
		caches[cacheType] = cache
	}

	return caches, nil
}

// GetSupportedTypes returns supported cache types
func (f *CacheFactory) GetSupportedTypes() []CacheType {
	return []CacheType{
		CacheTypeRedis,
		CacheTypeMemcache,
		CacheTypeInMemory,
		CacheTypeMultiLevel,
	}
}

// IsTypeSupported checks if a cache type is supported
func (f *CacheFactory) IsTypeSupported(cacheType CacheType) bool {
	supportedTypes := f.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if supportedType == cacheType {
			return true
		}
	}
	return false
}
