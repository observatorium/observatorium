package lightstep

// Taken from github.com/thanos-io/thanos/pkg/tracing/lightstep

// Config - YAML configuration.
// This also omits most empty fields (unlike Thanos upstream.)
type Config struct {
	// AccessToken is the unique API key for your LightStep project. It is
	// available on your account page at https://app.lightstep.com/account.
	AccessToken string `yaml:"access_token"`

	// Collector is the host, port, and plaintext option to use
	// for the collector.
	Collector Endpoint `yaml:"collector"`

	// Tags is a string comma-delimited of key value pairs that holds metadata that will be sent to lightstep
	Tags string `yaml:"tags,omitempty"`
}

// Taken from github.com/lightstep/lightstep-tracer-go/options.go
//
// Endpoint describes a collector or web API host/port and whether or
// not to use plaintext communication.
type Endpoint struct {
	Scheme           string `yaml:"scheme" json:"scheme" usage:"scheme to use for the endpoint, defaults to appropriate one if no custom one is required"`
	Host             string `yaml:"host" json:"host" usage:"host on which the endpoint is running"`
	Port             int    `yaml:"port" json:"port" usage:"port on which the endpoint is listening"`
	Plaintext        bool   `yaml:"plaintext,omitempty" json:"plaintext,omitempty" usage:"whether or not to encrypt data send to the endpoint"`
	CustomCACertFile string `yaml:"custom_ca_cert_file,omitempty" json:"custom_ca_cert_file,omitempty" usage:"path to a custom CA cert file, defaults to system defined certs if omitted"`
}
