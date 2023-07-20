package jaeger

import "time"

// Taken from github.com/thanos-io/thanos/pkg/tracing/jaeger

const (
	SamplerTypeRemote        = "remote"
	SamplerTypeProbabilistic = "probabilistic"
	SamplerTypeConstant      = "const"
	SamplerTypeRateLimiting  = "ratelimiting"
)

// Config - YAML configuration. For details see to https://github.com/jaegertracing/jaeger-client-go#environment-variables.
// This also omits most empty fields (unlike Thanos upstream.)
type Config struct {
	ServiceName                        string                   `yaml:"service_name"`
	Disabled                           bool                     `yaml:"disabled,omitempty"`
	RPCMetrics                         bool                     `yaml:"rpc_metrics,omitempty"`
	Tags                               string                   `yaml:"tags,omitempty"`
	SamplerType                        string                   `yaml:"sampler_type"`
	SamplerParam                       float64                  `yaml:"sampler_param"`
	SamplerManagerHostPort             string                   `yaml:"sampler_manager_host_port,omitempty"`
	SamplerMaxOperations               int                      `yaml:"sampler_max_operations,omitempty"`
	SamplerRefreshInterval             time.Duration            `yaml:"sampler_refresh_interval,omitempty"`
	SamplerParentConfig                ParentBasedSamplerConfig `yaml:"sampler_parent_config,omitempty"`
	SamplingServerURL                  string                   `yaml:"sampling_server_url,omitempty"`
	OperationNameLateBinding           bool                     `yaml:"operation_name_late_binding,omitempty"`
	InitialSamplingRate                float64                  `yaml:"initial_sampler_rate,omitempty"`
	ReporterMaxQueueSize               int                      `yaml:"reporter_max_queue_size,omitempty"`
	ReporterFlushInterval              time.Duration            `yaml:"reporter_flush_interval,omitempty"`
	ReporterLogSpans                   bool                     `yaml:"reporter_log_spans,omitempty"`
	ReporterDisableAttemptReconnecting bool                     `yaml:"reporter_disable_attempt_reconnecting,omitempty"`
	ReporterAttemptReconnectInterval   time.Duration            `yaml:"reporter_attempt_reconnect_interval,omitempty"`
	Endpoint                           string                   `yaml:"endpoint,omitempty"`
	User                               string                   `yaml:"user,omitempty"`
	Password                           string                   `yaml:"password,omitempty"`
	AgentHost                          string                   `yaml:"agent_host,omitempty"`
	AgentPort                          int                      `yaml:"agent_port,omitempty"`
	Gen128Bit                          bool                     `yaml:"traceid_128bit,omitempty"`
	// Remove the above field. Ref: https://github.com/open-telemetry/opentelemetry-specification/issues/525#issuecomment-605519217
	// Ref: https://opentelemetry.io/docs/reference/specification/trace/api/#spancontext
}

type ParentBasedSamplerConfig struct {
	LocalParentSampled  bool `yaml:"local_parent_sampled"`
	RemoteParentSampled bool `yaml:"remote_parent_sampled"`
}
