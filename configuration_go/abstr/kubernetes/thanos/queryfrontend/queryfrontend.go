package queryfrontend

import (
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultHTTPPort int = 10902
)

type CacheCompressionType string

const (
	CacheCompressionTypeSnappy CacheCompressionType = "snappy"
)

// NewTracingConfigFile returns a new tracing config file option.
func NewTracingConfigFile(value *trclient.TracingConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/tracing", "config.yaml", "tracing", "observatorium-thanos-query-tracing")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewRequestLoggingConfigFile returns a new request logging config file option.
func NewRequestLoggingConfigFile(value *reqlogging.RequestConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/request-logging", "config.yaml", "request-logging", "observatorium-thanos-query-request-logging")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewLabelsResponseCacheConfigFile returns a new labels response cache config file option.
func NewLabelsResponseCacheConfigFile(value *cache.ResponseCacheConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/labels-response-cache", "config.yaml", "labels-response-cache", "observatorium-thanos-query-labels-response-cache")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewQueryRangeResponseCacheConfigFile returns a new query range response cache config file option.
func NewQueryRangeResponseCacheConfigFile(value *cache.ResponseCacheConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/query-range-response-cache", "config.yaml", "query-range-response-cache", "observatorium-thanos-query-query-range-response-cache")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type QueryFrontendOptions struct {
	CacheCompressionType                 CacheCompressionType           `opt:"cache-compression-type"`
	HttpAddress                          *net.TCPAddr                   `opt:"http-address"`
	HttpGracePeriod                      time.Duration                  `opt:"http-grace-period"`
	HttpConfig                           string                         `opt:"http.config"`
	LabelsDefaultTimeRange               time.Duration                  `opt:"labels.default-time-range"`
	LabelsMaxQueryParallelism            int                            `opt:"labels.max-query-parallelism"`
	LabelsMaxRetriesPerRequest           *int                           `opt:"labels.max-retries-per-request"`
	LabelsPartialResponse                bool                           `opt:"labels.partial-response,noval"`
	LabelsResponseCacheConfig            *cache.ResponseCacheConfig     `opt:"labels.response-cache-config"`
	LabelsResponseCacheConfigFile        containeropts.ContainerUpdater `opt:"labels.response-cache-config-file"`
	LabelsResponseMaxFreshness           string                         `opt:"labels.response-cache-max-freshness"`
	LabelsSplitInterval                  time.Duration                  `opt:"labels.split-interval"`
	LogFormat                            log.Format                     `opt:"log.format"`
	LogLevel                             log.Level                      `opt:"log.level"`
	QueryFrontendCompressResponses       bool                           `opt:"query-frontend.compress-responses,noval"`
	QueryFrontendDownstreamTripperConfig *DownstreamTripperConfig       `opt:"query-frontend.downstream-tripper-config"`
	QueryFrontendDownstreamURL           string                         `opt:"query-frontend.downstream-url"`
	QueryFrontendForwardHeader           []string                       `opt:"query-frontend.forward-header"`
	QueryFrontendLogQueriesLongerThan    time.Duration                  `opt:"query-frontend.log-queries-longer-than"`
	QueryFrontendVerticalShards          int                            `opt:"query-frontend.vertical-shards"`
	QueryRangeAlignRangeWithStep         bool                           `opt:"query-range.align-range-with-step,noval"`
	QueryRangeHorizontalShards           int                            `opt:"query-range.horizontal-shards"`
	QueryRangeMaxQueryLength             time.Duration                  `opt:"query-range.max-query-length"`
	QueryRangeMaxQueryParallelism        int                            `opt:"query-range.max-query-parallelism"`
	QueryRangeMaxRetriesPerRequest       *int                           `opt:"query-range.max-retries-per-request"`
	QueryRangeMaxSplitInterval           time.Duration                  `opt:"query-range.max-split-interval"`
	QueryRangeMinSplitInterval           time.Duration                  `opt:"query-range.min-split-interval"`
	QueryRangePartialResponse            bool                           `opt:"query-range.partial-response,noval"`
	QueryRangeRequestDownsampled         bool                           `opt:"query-range.request-downsampled,noval"`
	QueryRangeResponseCacheConfig        *cache.ResponseCacheConfig     `opt:"query-range.response-cache-config"`
	QueryRangeResponseCacheConfigFile    containeropts.ContainerUpdater `opt:"query-range.response-cache-config-file"`
	QueryRangeResponseCacheMaxFreshness  time.Duration                  `opt:"query-range.response-cache-max-freshness"`
	QueryRangeSplitInterval              time.Duration                  `opt:"query-range.split-interval"`
	RequestLoggingConfig                 *reqlogging.RequestConfig      `opt:"request.logging-config"`
	RequestLoggingConfigFile             containeropts.ContainerUpdater `opt:"request.logging-config-file"`
	TracingConfig                        *trclient.TracingConfig        `opt:"tracing.config"`
	TracingConfigFile                    containeropts.ContainerUpdater `opt:"tracing.config-file"`
	WebDisableCORS                       bool                           `opt:"web.disable-cors,noval"`

	// Extra options not officially supported.
	cmdopt.ExtraOpts
}

type QueryFrontendDeployment struct {
	options *QueryFrontendOptions
	workload.DeploymentWorkload
}

func NewDefaultOptions() *QueryFrontendOptions {
	return &QueryFrontendOptions{
		LogLevel:  "warn",
		LogFormat: "logfmt",
	}
}

func NewQueryFrontend(opts *QueryFrontendOptions, namespace, imageTag string) *QueryFrontendDeployment {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-query-frontend",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "query-cache",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-query-frontend",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("500m", "2", "1Gi", "2Gi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: kghelpers.NewProbe("/-/healthy", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: kghelpers.NewProbe("/-/ready", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				kghelpers.NewEnvFromField("HOST_IP_ADDRESS", "status.hostIP"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}

	return &QueryFrontendDeployment{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (q *QueryFrontendDeployment) Objects() []runtime.Object {
	container := q.makeContainer()
	return q.DeploymentWorkload.Objects(container)
}

func (q *QueryFrontendDeployment) makeContainer() *workload.Container {
	httpPort := kghelpers.GetPortOrDefault(defaultHTTPPort, q.options.HttpAddress)
	kghelpers.CheckProbePort(httpPort, q.LivenessProbe)
	kghelpers.CheckProbePort(httpPort, q.ReadinessProbe)

	ret := q.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"query-frontend"}, cmdopt.GetOpts(q.options)...)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http", httpPort, httpPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if q.options.RequestLoggingConfigFile != nil {
		q.options.RequestLoggingConfigFile.Update(ret)
	}

	if q.options.TracingConfigFile != nil {
		q.options.TracingConfigFile.Update(ret)
	}

	if q.options.LabelsResponseCacheConfigFile != nil {
		q.options.LabelsResponseCacheConfigFile.Update(ret)
	}

	if q.options.QueryRangeResponseCacheConfigFile != nil {
		q.options.QueryRangeResponseCacheConfigFile.Update(ret)
	}

	return ret
}
