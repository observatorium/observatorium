package otlp

import (
	"time"
)

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/tracing/otlp/config_yaml.go#L22

type retryConfig struct {
	RetryEnabled         bool          `yaml:"retry_enabled,omitempty"`
	RetryInitialInterval time.Duration `yaml:"retry_initial_interval,omitempty"`
	RetryMaxInterval     time.Duration `yaml:"retry_max_interval,omitempty"`
	RetryMaxElapsedTime  time.Duration `yaml:"retry_max_elapsed_time,omitempty"`
}

// Config - YAML configuration.
// This also omits most empty fields (unlike Thanos upstream.)
type Config struct {
	ClientType         string            `yaml:"client_type,omitempty"`
	ServiceName        string            `yaml:"service_name"`
	ReconnectionPeriod time.Duration     `yaml:"reconnection_period,omitempty"`
	Compression        string            `yaml:"compression,omitempty"`
	Insecure           bool              `yaml:"insecure,omitempty"`
	Endpoint           string            `yaml:"endpoint"`
	URLPath            string            `yaml:"url_path,omitempty"`
	Timeout            time.Duration     `yaml:"timeout,omitempty"`
	RetryConfig        retryConfig       `yaml:"retry_config,omitempty"`
	Headers            map[string]string `yaml:"headers,omitempty"`
	TLSConfig          TLSConfig         `yaml:"tls_config,omitempty"`
	SamplerType        string            `yaml:"sampler_type"`
	SamplerParam       string            `yaml:"sampler_param"`
}

// TLSConfig configures the options for TLS connections.
//
// Taken from github.com/thanos-io/thanos/pkg/exthttp
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`
}
