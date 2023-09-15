package cache

type IndexCacheProvider string

// Taken from github.com/thanos-io/pkg/store/cache/factory.go

const (
	INMEMORY  IndexCacheProvider = "IN-MEMORY"
	MEMCACHED IndexCacheProvider = "MEMCACHED"
	REDIS     IndexCacheProvider = "REDIS"
)

// IndexCacheConfig specifies the index cache config.
type IndexCacheConfig struct {
	Type   IndexCacheProvider `yaml:"type"`
	Config interface{}        `yaml:"config"`
}
