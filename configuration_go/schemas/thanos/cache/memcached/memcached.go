package memcached

import (
	"time"

	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/units"
)

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/cacheutil/memcached_client.go#L106

var DefaultMemcachedClientConfig = MemcachedClientConfig{
	Timeout:                   500 * time.Millisecond,
	MaxIdleConnections:        100,
	MaxAsyncConcurrency:       20,
	MaxAsyncBufferSize:        10000,
	MaxItemSize:               units.Bytes(1024 * 1024),
	MaxGetMultiConcurrency:    100,
	MaxGetMultiBatchSize:      0,
	DNSProviderUpdateInterval: 10 * time.Second,
	AutoDiscovery:             false,
}

// MemcachedClientConfig is the config accepted by RemoteCacheClient.
type MemcachedClientConfig struct {
	// Addresses specifies the list of memcached addresses. The addresses get
	// resolved with the DNS provider.
	Addresses []string `yaml:"addresses"`

	// Timeout specifies the socket read/write timeout.
	Timeout time.Duration `yaml:"timeout,omitempty"`

	// MaxIdleConnections specifies the maximum number of idle connections that
	// will be maintained per address. For better performances, this should be
	// set to a number higher than your peak parallel requests.
	MaxIdleConnections int `yaml:"max_idle_connections,omitempty"`

	// MaxAsyncConcurrency specifies the maximum number of SetAsync goroutines.
	MaxAsyncConcurrency int `yaml:"max_async_concurrency,omitempty"`

	// MaxAsyncBufferSize specifies the queue buffer size for SetAsync operations.
	MaxAsyncBufferSize int `yaml:"max_async_buffer_size,omitempty"`

	// MaxGetMultiConcurrency specifies the maximum number of concurrent GetMulti() operations.
	// If set to 0, concurrency is unlimited.
	MaxGetMultiConcurrency int `yaml:"max_get_multi_concurrency,omitempty"`

	// MaxItemSize specifies the maximum size of an item stored in memcached.
	// Items bigger than MaxItemSize are skipped.
	// If set to 0, no maximum size is enforced.
	MaxItemSize units.Bytes `yaml:"max_item_size,omitempty"`

	// MaxGetMultiBatchSize specifies the maximum number of keys a single underlying
	// GetMulti() should run. If more keys are specified, internally keys are splitted
	// into multiple batches and fetched concurrently, honoring MaxGetMultiConcurrency parallelism.
	// If set to 0, the max batch size is unlimited.
	MaxGetMultiBatchSize int `yaml:"max_get_multi_batch_size,omitempty"`

	// DNSProviderUpdateInterval specifies the DNS discovery update interval.
	DNSProviderUpdateInterval time.Duration `yaml:"dns_provider_update_interval,omitempty"`

	// AutoDiscovery configures memached client to perform auto-discovery instead of DNS resolution
	AutoDiscovery bool `yaml:"auto_discovery,omitempty"`
}

func (c MemcachedClientConfig) Type() string {
	return "MEMCACHED"
}
