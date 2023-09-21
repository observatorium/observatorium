package cache

import (
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memcached"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memory"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/redis"
)

type cacheConfig interface {
	redis.RedisClientConfig | memcached.MemcachedClientConfig | memory.MemoryCacheConfig
	Type() string
}

// IndexCacheConfig specifies the index cache config.
type IndexCacheConfig[T cacheConfig] struct {
	ConfigType string `yaml:"type"`
	Config     T      `yaml:"config"`
}

func NewConfig[T cacheConfig](c T) *IndexCacheConfig[T] {
	return &IndexCacheConfig[T]{
		ConfigType: c.Type(),
		Config:     c,
	}
}
