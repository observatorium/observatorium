package queryfrontend

import (
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultHTTPPort int = 10902
)

type CacheCompressionType string

const (
	CacheCompressionTypeSnappy CacheCompressionType = "snappy"
)

type tracingConfigFile = k8sutil.ConfigFile

// NewTracingConfigFile returns a new tracing config file option.
func NewTracingConfigFile(value *trclient.TracingConfig) *tracingConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/tracing", "config.yaml", "tracing", "observatorium-thanos-query-tracing")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type requestLoggingConfigFile = k8sutil.ConfigFile

// NewRequestLoggingConfigFile returns a new request logging config file option.
func NewRequestLoggingConfigFile(value *reqlogging.RequestConfig) *requestLoggingConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/request-logging", "config.yaml", "request-logging", "observatorium-thanos-query-request-logging")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type labelsResponseCacheConfig = k8sutil.ConfigFile

// NewLabelsResponseCacheConfigFile returns a new labels response cache config file option.
func NewLabelsResponseCacheConfigFile(value *cache.ResponseCacheConfig) *labelsResponseCacheConfig {
	ret := k8sutil.NewConfigFile("/etc/thanos/labels-response-cache", "config.yaml", "labels-response-cache", "observatorium-thanos-query-labels-response-cache")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type queryRangeResponseCacheConfig = k8sutil.ConfigFile

// NewQueryRangeResponseCacheConfigFile returns a new query range response cache config file option.
func NewQueryRangeResponseCacheConfigFile(value *cache.ResponseCacheConfig) *queryRangeResponseCacheConfig {
	ret := k8sutil.NewConfigFile("/etc/thanos/query-range-response-cache", "config.yaml", "query-range-response-cache", "observatorium-thanos-query-query-range-response-cache")
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
	LabelsResponseCacheConfigFile        *labelsResponseCacheConfig     `opt:"labels.response-cache-config-file"`
	LabelsResponseMaxFreshness           string                         `opt:"labels.response-cache-max-freshness"`
	LabelsSplitInterval                  time.Duration                  `opt:"labels.split-interval"`
	LogFormat                            log.LogFormat                  `opt:"log.format"`
	LogLevel                             log.LogLevel                   `opt:"log.level"`
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
	QueryRangeResponseCacheConfigFile    *queryRangeResponseCacheConfig `opt:"query-range.response-cache-config-file"`
	QueryRangeResponseCacheMaxFreshness  time.Duration                  `opt:"query-range.response-cache-max-freshness"`
	QueryRangeSplitInterval              time.Duration                  `opt:"query-range.split-interval"`
	RequestLoggingConfig                 *reqlogging.RequestConfig      `opt:"request.logging-config"`
	RequestLoggingConfigFile             *requestLoggingConfigFile      `opt:"request.logging-config-file"`
	TracingConfig                        *trclient.TracingConfig        `opt:"tracing.config"`
	TracingConfigFile                    *tracingConfigFile             `opt:"tracing.config-file"`
	WebDisableCORS                       bool                           `opt:"web.disable-cors,noval"`
}

type QueryFrontendDeployment struct {
	options *QueryFrontendOptions

	k8sutil.DeploymentGenericConfig
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
		k8sutil.NameLabel:      "thanos-query-frontend",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "query-cache",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	return &QueryFrontendDeployment{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-query-frontend",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("500m", "2", "1Gi", "2Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", probePort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", probePort, k8sutil.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				k8sutil.NewEnvFromField("HOST_IP_ADDRESS", "status.hostIP"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}
}

func (q *QueryFrontendDeployment) Manifests() k8sutil.ObjectMap {

	container := q.makeContainer()

	ret := k8sutil.ObjectMap{}
	ret.AddAll(q.GenerateObjects(container))

	return ret
}

func (q *QueryFrontendDeployment) makeContainer() *k8sutil.Container {
	httpPort := k8sutil.GetPortOrDefault(defaultHTTPPort, q.options.HttpAddress)
	k8sutil.CheckProbePort(httpPort, q.LivenessProbe)
	k8sutil.CheckProbePort(httpPort, q.ReadinessProbe)

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
		k8sutil.NewServicePort("http", httpPort, httpPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if q.options.RequestLoggingConfigFile != nil {
		q.options.RequestLoggingConfigFile.AddToContainer(ret)
	}

	if q.options.TracingConfigFile != nil {
		q.options.TracingConfigFile.AddToContainer(ret)
	}

	if q.options.LabelsResponseCacheConfigFile != nil {
		q.options.LabelsResponseCacheConfigFile.AddToContainer(ret)
	}

	if q.options.QueryRangeResponseCacheConfigFile != nil {
		q.options.QueryRangeResponseCacheConfigFile.AddToContainer(ret)
	}

	return ret
}
