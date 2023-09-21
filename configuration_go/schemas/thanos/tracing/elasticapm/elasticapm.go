package elasticapm

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/tracing/elasticapm/elastic_apm.go#L19

type Config struct {
	ServiceName        string  `yaml:"service_name"`
	ServiceVersion     string  `yaml:"service_version"`
	ServiceEnvironment string  `yaml:"service_environment"`
	SampleRate         float64 `yaml:"sample_rate"`
}
