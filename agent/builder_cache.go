package agent

import (
	"context"
	"time"
)

// Cache configuration methods for Builder
// This file contains all methods related to cache management,
// including memory cache, Redis cache, and cache operations.

func (b *Builder) WithCache(cache Cache) *Builder {
	b.cache = cache
	b.cacheEnabled = true
	return b
}

func (b *Builder) WithMemoryCache(maxSize int, defaultTTL time.Duration) *Builder {
	b.cache = NewMemoryCache(maxSize, defaultTTL)
	b.cacheEnabled = true
	return b
}

func (b *Builder) WithRedisCache(addr, password string, db int) *Builder {
	cache, err := NewRedisCache(addr, password, db, 5*time.Minute)
	if err != nil {
		// Log error but don't fail - fall back to no caching
		return b
	}
	b.cache = cache
	b.cacheEnabled = true
	return b
}

func (b *Builder) WithRedisCacheOptions(opts *RedisCacheOptions) *Builder {
	cache, err := NewRedisCacheWithOptions(opts)
	if err != nil {
		// Log error but don't fail - fall back to no caching
		return b
	}
	b.cache = cache
	b.cacheEnabled = true
	return b
}

func (b *Builder) WithCacheTTL(ttl time.Duration) *Builder {
	b.cacheTTL = ttl
	return b
}

func (b *Builder) DisableCache() *Builder {
	b.cacheEnabled = false
	return b
}

func (b *Builder) EnableCache() *Builder {
	if b.cache != nil {
		b.cacheEnabled = true
	}
	return b
}

func (b *Builder) GetCacheStats() CacheStats {
	logger := b.getLogger()
	if b.cache != nil {
		stats := b.cache.Stats()
		hitRate := 0.0
		if stats.Hits+stats.Misses > 0 {
			hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
		}
		logger.Debug(context.Background(), "Cache stats retrieved",
			F("hits", stats.Hits),
			F("misses", stats.Misses),
			F("size", stats.Size),
			F("hit_rate", hitRate))
		return stats
	}
	logger.Debug(context.Background(), "No cache configured")
	return CacheStats{}
}

func (b *Builder) ClearCache(ctx context.Context) error {
	logger := b.getLogger()
	if b.cache != nil {
		logger.Info(ctx, "Clearing cache")
		err := b.cache.Clear(ctx)
		if err != nil {
			logger.Error(ctx, "Failed to clear cache", F("error", err.Error()))
			return err
		}
		logger.Info(ctx, "Cache cleared successfully")
		return nil
	}
	logger.Debug(ctx, "No cache to clear")
	return nil
}
