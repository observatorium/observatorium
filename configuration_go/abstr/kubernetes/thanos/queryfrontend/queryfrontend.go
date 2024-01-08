package queryfrontend

import (
	"fmt"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache"
	thanoslog "github.com/observatorium/observatorium/configuration_go/schemas/thanos/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultNamespace string = "observatorium"
	defaultHTTPPort  int    = 10902
)

type CacheCompressionType string

const (
	CacheCompressionTypeSnappy CacheCompressionType = "snappy"
)

type tracingConfigFile = k8sutil.ConfigFile

// NewReceiveLimitsConfigFile returns a new tracing config file option.
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
	LogFormat                            thanoslog.LogFormat            `opt:"log.format"`
	LogLevel                             thanoslog.LogLevel             `opt:"log.level"`
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
	Options *QueryFrontendOptions

	k8sutil.DeploymentGenericConfig
}

func NewQueryFrontend() *QueryFrontendDeployment {
	opts := &QueryFrontendOptions{
		LogLevel:  "warn",
		LogFormat: "logfmt",
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-query-frontend",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "query-cache", // TODO
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &QueryFrontendDeployment{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-query-frontend",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("500m", "2", "1Gi", "2Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", defaultHTTPPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", defaultHTTPPort, k8sutil.ProbeConfig{
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

func (s *QueryFrontendDeployment) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      s.Name,
		Labels:    s.CommonLabels,
		Namespace: s.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &s.TerminationGracePeriodSeconds,
		Affinity:                      s.Affinity,
		SecurityContext:               s.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container}, s.Sidecars...),
	}

	statefulset := &k8sutil.Deployment{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   s.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"query-fe-statefulSet": statefulset.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["query-fe-service"] = service.MakeManifest()

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["query-fe-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["query-fe-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["query-fe-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["query-fe-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *QueryFrontendDeployment) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &QueryFrontendOptions{}
	}

	httpPort := defaultHTTPPort
	if s.Options.HttpAddress != nil && s.Options.HttpAddress.Port != 0 {
		httpPort = s.Options.HttpAddress.Port
	}

	livenessPort := s.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(httpPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, httpPort))
	}

	readinessPort := s.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(httpPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, httpPort))
	}

	ret := s.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"query-frontend"}, cmdopt.GetOpts(s.Options)...)
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

	if s.Options.RequestLoggingConfigFile != nil {
		s.Options.RequestLoggingConfigFile.AddToContainer(ret)
	}

	if s.Options.TracingConfigFile != nil {
		s.Options.TracingConfigFile.AddToContainer(ret)
	}

	if s.Options.LabelsResponseCacheConfigFile != nil {
		s.Options.LabelsResponseCacheConfigFile.AddToContainer(ret)
	}

	if s.Options.QueryRangeResponseCacheConfigFile != nil {
		s.Options.QueryRangeResponseCacheConfigFile.AddToContainer(ret)
	}

	return ret
}
