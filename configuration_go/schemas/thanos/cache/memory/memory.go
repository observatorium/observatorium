package memory

import "github.com/thanos-io/thanos/pkg/model"

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/store/cache/inmemory.go#L56

// InMemoryIndexCacheConfig holds the in-memory index cache config.
type MemoryCacheConfig struct {
	MaxSize     model.Bytes `yaml:"max_size,omitempty"`
	MaxItemSize model.Bytes `yaml:"max_item_size,omitempty"`
}

func (c MemoryCacheConfig) Type() string {
	return "MEMORY"
}
