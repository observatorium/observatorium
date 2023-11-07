package queryfrontend

import (
	"fmt"

	prommodel "github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type DownstreamTripperConfig struct {
	IdleConnTimeout       prommodel.Duration `yaml:"idle_conn_timeout,omitempty"`
	ResponseHeaderTimeout prommodel.Duration `yaml:"response_header_timeout,omitempty"`
	TLSHandshakeTimeout   prommodel.Duration `yaml:"tls_handshake_timeout,omitempty"`
	ExpectContinueTimeout prommodel.Duration `yaml:"expect_continue_timeout,omitempty"`
	MaxIdleConns          *int               `yaml:"max_idle_conns,omitempty"`
	MaxIdleConnsPerHost   *int               `yaml:"max_idle_conns_per_host,omitempty"`
	MaxConnsPerHost       *int               `yaml:"max_conns_per_host,omitempty"`
}

// String returns a string representation of the IndexCacheConfig as YAML.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
func (c DownstreamTripperConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("error mashalling DownstreamTripperConfig to yaml: %v", err))
	}
	return string(ret)
}
