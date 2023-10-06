package client

import "gopkg.in/yaml.v2"

// Taken from github.com/thanos-io/thanos/pkg/tracing/client/client.go v0.32.2

type TracingProvider string

const (
	Stackdriver           TracingProvider = "STACKDRIVER"
	GoogleCloud           TracingProvider = "GOOGLE_CLOUD"
	Jaeger                TracingProvider = "JAEGER"
	ElasticAPM            TracingProvider = "ELASTIC_APM"
	Lightstep             TracingProvider = "LIGHTSTEP"
	OpenTelemetryProtocol TracingProvider = "OTLP"
)

type TracingConfig struct {
	Type   TracingProvider `yaml:"type"`
	Config interface{}     `yaml:"config"`
}

func (c TracingConfig) String() string {
	// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
	ret, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(ret)
}
