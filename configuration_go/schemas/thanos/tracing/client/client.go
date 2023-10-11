package client

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
