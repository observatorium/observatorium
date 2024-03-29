package reqlogging

import "gopkg.in/yaml.v2"

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/logging/yaml_parser.go

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARNING"
	LogLevelError LogLevel = "ERROR"
)

// RequestConfig is the configuration for request logging.
type RequestConfig struct {
	HTTP    HTTPProtocolConfigs `yaml:"http,omitempty"`
	GRPC    GRPCProtocolConfigs `yaml:"grpc,omitempty"`
	Options OptionsConfig       `yaml:"options,omitempty"`
}

// String returns a string representation of the RequestConfig as YAML.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
func (c RequestConfig) String() string {
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(ret)
}

type HTTPProtocolConfigs struct {
	Options OptionsConfig        `yaml:"options,omitempty"`
	Config  []HTTPProtocolConfig `yaml:"config,omitempty"`
}

type GRPCProtocolConfigs struct {
	Options OptionsConfig        `yaml:"options,omitempty"`
	Config  []GRPCProtocolConfig `yaml:"config,omitempty"`
}

type OptionsConfig struct {
	Level    LogLevel       `yaml:"level,omitempty"`
	Decision DecisionConfig `yaml:"decision,omitempty"`
}

type DecisionConfig struct {
	LogStart bool `yaml:"log_start,omitempty"`
	LogEnd   bool `yaml:"log_end,omitempty"`
}

type HTTPProtocolConfig struct {
	Path string `yaml:"path,omitempty"`
	Port uint64 `yaml:"port,omitempty"`
}

type GRPCProtocolConfig struct {
	Service string `yaml:"service,omitempty"`
	Method  string `yaml:"method,omitempty"`
}
