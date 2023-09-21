package redis

import (
	"time"

	"github.com/thanos-io/thanos/pkg/model"
)

// Taken from github.com/thanos-io/thanos/pkg/cacheutil/redis_client.go v0.32.2

var DefaultRedisClientConfig = RedisClientConfig{
	DialTimeout:            time.Second * 5,
	ReadTimeout:            time.Second * 3,
	WriteTimeout:           time.Second * 3,
	MaxGetMultiConcurrency: 100,
	GetMultiBatchSize:      100,
	MaxSetMultiConcurrency: 100,
	SetMultiBatchSize:      100,
	MaxAsyncConcurrency:    20,
	MaxAsyncBufferSize:     10000,
}

// RedisClientConfig is the config accepted by RedisClient.
type RedisClientConfig struct {
	// Addr specifies the addresses of redis server.
	Addr string `yaml:"addr"`

	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string `yaml:"username,omitempty"`
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string `yaml:"password,omitempty"`

	// DB Database to be selected after connecting to the server.
	DB int `yaml:"db,omitempty"`

	// DialTimeout specifies the client dial timeout.
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`

	// ReadTimeout specifies the client read timeout.
	ReadTimeout time.Duration `yaml:"read_timeout,omitempty"`

	// WriteTimeout specifies the client write timeout.
	WriteTimeout time.Duration `yaml:"write_timeout,omitempty"`

	// MaxGetMultiConcurrency specifies the maximum number of concurrent GetMulti() operations.
	// If set to 0, concurrency is unlimited.
	MaxGetMultiConcurrency int `yaml:"max_get_multi_concurrency,omitempty"`

	// GetMultiBatchSize specifies the maximum size per batch for mget.
	GetMultiBatchSize int `yaml:"get_multi_batch_size,omitempty"`

	// MaxSetMultiConcurrency specifies the maximum number of concurrent SetMulti() operations.
	// If set to 0, concurrency is unlimited.
	MaxSetMultiConcurrency int `yaml:"max_set_multi_concurrency,omitempty"`

	// SetMultiBatchSize specifies the maximum size per batch for pipeline set.
	SetMultiBatchSize int `yaml:"set_multi_batch_size,omitempty"`

	// TLSEnabled enable tls for redis connection.
	TLSEnabled bool `yaml:"tls_enabled,omitempty"`

	// TLSConfig to use to connect to the redis server.
	TLSConfig TLSConfig `yaml:"tls_config,omitempty"`

	// If not zero then client-side caching is enabled.
	// Client-side caching is when data is stored in memory
	// instead of fetching data each time.
	// See https://redis.io/docs/manual/client-side-caching/ for info.
	CacheSize model.Bytes `yaml:"cache_size,omitempty"`

	// MasterName specifies the master's name. Must be not empty
	// for Redis Sentinel.
	MasterName string `yaml:"master_name,omitempty"`

	// MaxAsyncBufferSize specifies the queue buffer size for SetAsync operations.
	MaxAsyncBufferSize int `yaml:"max_async_buffer_size,omitempty"`

	// MaxAsyncConcurrency specifies the maximum number of SetAsync goroutines.
	MaxAsyncConcurrency int `yaml:"max_async_concurrency,omitempty"`
}

func (c RedisClientConfig) Type() string {
	return "REDIS"
}

// TLSConfig configures TLS connections.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"key_file,omitempty"`
	// Used to verify the hostname for the targets. See https://tools.ietf.org/html/rfc4366#section-3.1
	ServerName string `yaml:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`
}
