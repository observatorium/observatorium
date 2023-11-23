package memory

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/store/cache/inmemory.go#L56

// InMemoryIndexCacheConfig holds the in-memory index cache config.
type MemoryCacheConfig struct {
	MaxSize     string `yaml:"max_size,omitempty"`
	MaxItemSize string `yaml:"max_item_size,omitempty"`
}

func (c MemoryCacheConfig) Type() string {
	return "MEMORY"
}
