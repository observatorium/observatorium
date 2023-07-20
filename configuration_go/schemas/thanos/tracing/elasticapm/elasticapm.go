package elasticapm

// Taken from github.com/thanos-io/thanos/pkg/tracing/elasticapm

type Config struct {
	ServiceName        string  `yaml:"service_name"`
	ServiceVersion     string  `yaml:"service_version"`
	ServiceEnvironment string  `yaml:"service_environment"`
	SampleRate         float64 `yaml:"sample_rate"`
}
