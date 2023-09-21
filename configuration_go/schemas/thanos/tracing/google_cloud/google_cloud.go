package google_cloud

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/tracing/google_cloud/google_cloud.go#L24

// Config - YAML configuration.
type Config struct {
	ServiceName  string `yaml:"service_name"`
	ProjectId    string `yaml:"project_id"`
	SampleFactor uint64 `yaml:"sample_factor"`
}
