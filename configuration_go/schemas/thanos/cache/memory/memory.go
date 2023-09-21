package memory

import "github.com/thanos-io/thanos/pkg/model"

// Taken from github.com/thanos-io/thanos/pkg/store/cache/inmemory.go v0.32.2

// InMemoryIndexCacheConfig holds the in-memory index cache config.
type MemoryCacheConfig struct {
	MaxSize     model.Bytes `yaml:"max_size,omitempty"`
	MaxItemSize model.Bytes `yaml:"max_item_size,omitempty"`
}

func (c MemoryCacheConfig) Type() string {
	return "MEMORY"
}
