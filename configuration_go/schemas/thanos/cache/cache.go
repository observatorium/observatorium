package cache

import (
	"fmt"
	"time"

	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memcached"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/memory"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache/redis"
	"gopkg.in/yaml.v2"
)

type IndexCacheEnabledItem string

const (
	Postings         IndexCacheEnabledItem = "Postings"
	Series           IndexCacheEnabledItem = "Series"
	ExpandedPostings IndexCacheEnabledItem = "ExpandedPostings"
)

type CacheConfig interface {
	redis.RedisClientConfig | memcached.MemcachedClientConfig | memory.MemoryCacheConfig
	Type() string
}

func NewIndexCacheConfig[T CacheConfig](c T) *IndexCacheConfig {
	return &IndexCacheConfig{
		ConfigType: c.Type(),
		Config:     c,
	}
}

func NewBucketCacheConfig[T CacheConfig](c T) *BucketCacheConfig {
	return &BucketCacheConfig{
		Type:   c.Type(),
		Config: c,
	}
}

func NewResponseCacheConfig[T CacheConfig](c T) *ResponseCacheConfig {
	return &ResponseCacheConfig{
		Type:   c.Type(),
		Config: c,
	}
}

// IndexCacheConfig specifies the index cache config.
type IndexCacheConfig struct {
	ConfigType string      `yaml:"type"`
	Config     interface{} `yaml:"config"`
	// Available item types are Postings, Series and ExpandedPostings.
	EnabledItems []IndexCacheEnabledItem `yaml:"enabled_items,omitempty"`
	// TTL for storing items in remote cache. Not supported for inmemory cache.
	// Default value is 24h.
	TTL time.Duration `yaml:"ttl,omitempty"`
}

// String returns a string representation of the IndexCacheConfig as YAML.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
func (c IndexCacheConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("error mashalling IndexCacheConfig to yaml: %v", err))
	}
	return string(ret)
}

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/store/cache/caching_bucket_factory.go#L37

// CachingWithBackendConfig is a configuration of caching bucket used by Store component.
type BucketCacheConfig struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config"`

	// Basic unit used to cache chunks.
	ChunkSubrangeSize int64 `yaml:"chunk_subrange_size,omitempty"`

	// Maximum number of GetRange requests issued by this bucket for single GetRange call. Zero or negative value = unlimited.
	MaxChunksGetRangeRequests int `yaml:"max_chunks_get_range_requests,omitempty"`

	MetafileMaxSize string `yaml:"metafile_max_size,omitempty"`

	// TTLs for various cache items.
	ChunkObjectAttrsTTL time.Duration `yaml:"chunk_object_attrs_ttl,omitempty"`
	ChunkSubrangeTTL    time.Duration `yaml:"chunk_subrange_ttl,omitempty"`

	// How long to cache result of Iter call in root directory.
	BlocksIterTTL time.Duration `yaml:"blocks_iter_ttl,omitempty"`

	// Config for Exists and Get operations for metadata files.
	MetafileExistsTTL      time.Duration `yaml:"metafile_exists_ttl,omitempty"`
	MetafileDoesntExistTTL time.Duration `yaml:"metafile_doesnt_exist_ttl,omitempty"`
	MetafileContentTTL     time.Duration `yaml:"metafile_content_ttl,omitempty"`
}

// String returns a string representation of the BucketCacheConfig as YAML.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
func (c BucketCacheConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("error mashalling BucketCacheConfig to yaml: %v", err))
	}
	return string(ret)
}

type ResponseCacheConfig struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config"`
}

// String returns a string representation of the ResponseCacheConfig as YAML.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
func (c ResponseCacheConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("error mashalling ResponseCacheConfig to yaml: %v", err))
	}
	return string(ret)
}
