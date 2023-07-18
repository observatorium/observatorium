package jaeger

import "time"

const (
	SamplerTypeRemote        = "remote"
	SamplerTypeProbabilistic = "probabilistic"
	SamplerTypeConstant      = "const"
	SamplerTypeRateLimiting  = "ratelimiting"
)

type Config struct {
	ServiceName                        string                   `yaml:"service_name"`
	Disabled                           bool                     `yaml:"disabled"`
	RPCMetrics                         bool                     `yaml:"rpc_metrics"`
	Tags                               string                   `yaml:"tags"`
	SamplerType                        string                   `yaml:"sampler_type"`
	SamplerParam                       float64                  `yaml:"sampler_param"`
	SamplerManagerHostPort             string                   `yaml:"sampler_manager_host_port"`
	SamplerMaxOperations               int                      `yaml:"sampler_max_operations"`
	SamplerRefreshInterval             time.Duration            `yaml:"sampler_refresh_interval"`
	SamplerParentConfig                ParentBasedSamplerConfig `yaml:"sampler_parent_config"`
	SamplingServerURL                  string                   `yaml:"sampling_server_url"`
	OperationNameLateBinding           bool                     `yaml:"operation_name_late_binding"`
	InitialSamplingRate                float64                  `yaml:"initial_sampler_rate"`
	ReporterMaxQueueSize               int                      `yaml:"reporter_max_queue_size"`
	ReporterFlushInterval              time.Duration            `yaml:"reporter_flush_interval"`
	ReporterLogSpans                   bool                     `yaml:"reporter_log_spans"`
	ReporterDisableAttemptReconnecting bool                     `yaml:"reporter_disable_attempt_reconnecting"`
	ReporterAttemptReconnectInterval   time.Duration            `yaml:"reporter_attempt_reconnect_interval"`
	Endpoint                           string                   `yaml:"endpoint"`
	User                               string                   `yaml:"user"`
	Password                           string                   `yaml:"password"`
	AgentHost                          string                   `yaml:"agent_host"`
	AgentPort                          int                      `yaml:"agent_port"`
	Gen128Bit                          bool                     `yaml:"traceid_128bit"`
	// Remove the above field. Ref: https://github.com/open-telemetry/opentelemetry-specification/issues/525#issuecomment-605519217
	// Ref: https://opentelemetry.io/docs/reference/specification/trace/api/#spancontext
}

type ParentBasedSamplerConfig struct {
	LocalParentSampled  bool `yaml:"local_parent_sampled"`
	RemoteParentSampled bool `yaml:"remote_parent_sampled"`
}
