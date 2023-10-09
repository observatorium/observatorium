package cache

import (
	"fmt"

	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memcached"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memory"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/redis"
	"gopkg.in/yaml.v2"
)

// IndexCacheConfig specifies the index cache config.
type IndexCacheConfig struct {
	ConfigType string      `yaml:"type"`
	Config     interface{} `yaml:"config"`
}

func (c IndexCacheConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("error mashalling IndexCacheConfig to yaml: %v", err))
	}
	return string(ret)
}

type CacheConfig interface {
	redis.RedisClientConfig | memcached.MemcachedClientConfig | memory.MemoryCacheConfig
	Type() string
}

func NewConfig[T CacheConfig](c T) *IndexCacheConfig {
	return &IndexCacheConfig{
		ConfigType: c.Type(),
		Config:     c,
	}
}
