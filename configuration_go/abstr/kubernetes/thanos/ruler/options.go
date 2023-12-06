package ruler

import "github.com/prometheus/common/model"

type AlertingConfig struct {
	Alertmanagers []AlertmanagerConfig `yaml:"alertmanagers"`
}

// AlertmanagerConfig represents a client to a cluster of Alertmanager endpoints.
type AlertmanagerConfig struct {
	HTTPClientConfig ClientConfig `yaml:"http_config,omitempty"`
	// List of addresses with DNS prefixes.
	StaticAddresses []string `yaml:"static_configs,omitempty"`
	// List of file  configurations (our FileSD supports different DNS lookups).
	FileSDConfigs []FileSDConfig `yaml:"file_sd_configs,omitempty"`

	// The URL scheme to use when talking to targets.
	Scheme string `yaml:"scheme,omitempty"`

	// Path prefix to add in front of the endpoint path.
	PathPrefix string         `yaml:"path_prefix,omitempty"`
	Timeout    model.Duration `yaml:"timeout,omitempty"`
	APIVersion APIVersion     `yaml:"api_version,omitempty"`
}

// APIVersion represents the API version of the Alertmanager endpoint.
type APIVersion string

const (
	APIv1 APIVersion = "v1"
	APIv2 APIVersion = "v2"
)

// FileSDConfig represents a file service discovery configuration.
type FileSDConfig struct {
	Files           []string       `yaml:"files,omitempty"`
	RefreshInterval model.Duration `yaml:"refresh_interval,omitempty"`
}

// ClientConfig configures an HTTP client.
type ClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth BasicAuth `yaml:"basic_auth,omitempty"`
	// The bearer token for the targets.
	BearerToken string `yaml:"bearer_token,omitempty"`
	// The bearer token file for the targets.
	BearerTokenFile string `yaml:"bearer_token_file,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL string `yaml:"proxy_url,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig TLSConfig `yaml:"tls_config,omitempty"`
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

// BasicAuth configures basic authentication for HTTP clients.
type BasicAuth struct {
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty"`
}

// Label represents a single label configuration.
type Label struct {
	Key   string
	Value string
}

func (l Label) String() string {
	return l.Key + "=" + l.Value
}
