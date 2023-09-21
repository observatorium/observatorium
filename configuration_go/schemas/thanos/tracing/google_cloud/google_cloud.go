package google_cloud

// Taken from github.com/thanos-io/thanos/pkg/tracing/google_cloud v0.32.2

// Config - YAML configuration.
type Config struct {
	ServiceName  string `yaml:"service_name"`
	ProjectId    string `yaml:"project_id"`
	SampleFactor uint64 `yaml:"sample_factor"`
}
