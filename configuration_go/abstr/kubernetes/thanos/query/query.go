package query

import (
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"

	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type GrpcCompressionType string

const (
	defaultHTTPPort       int                 = 10902
	defaultGRPCPort       int                 = 10901
	GrpcCompressionSnappy GrpcCompressionType = "snappy"
	GrpcCompressionNone   GrpcCompressionType = "none"
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

type QueryOptions struct {
	AlertQueryURL                                 string                         `opt:"alert.query-url"`
	EnableFeature                                 string                         `opt:"enable-feature"`
	Endpoint                                      []string                       `opt:"endpoint"`
	EndpointGroup                                 []string                       `opt:"endpoint-group"`
	EndpointStrict                                []string                       `opt:"endpoint-strict"`
	EndpointGroupStrict                           []string                       `opt:"endpoint-group-strict"`
	GrpcAddress                                   *net.TCPAddr                   `opt:"grpc-address"`
	GrpcClientsServerName                         string                         `opt:"grpc-client-server-name"`
	GrpcClientsTLSCA                              string                         `opt:"grpc-client-tls-ca"`
	GrpcClientsTLSCert                            string                         `opt:"grpc-client-tls-cert"`
	GrpcClientsTLSKey                             string                         `opt:"grpc-client-tls-key"`
	GrpcClientsTLSSecure                          bool                           `opt:"grpc-client-tls-secure,noval"`
	GrpcClientsTLSSkipVerify                      bool                           `opt:"grpc-client-tls-skip-verify,noval"`
	GrpcClientsCompression                        GrpcCompressionType            `opt:"grpc-compression"`
	GrpcGracePeriod                               time.Duration                  `opt:"grpc-grace-period"`
	GrpcMMaxConnectionAge                         time.Duration                  `opt:"grpc-server-max-connection-age"`
	GrpcServerTLSCert                             string                         `opt:"grpc-server-tls-cert"`
	GrpcServerTLSClientCA                         string                         `opt:"grpc-server-tls-client-ca"`
	GrpcServerTLSKey                              string                         `opt:"grpc-server-tls-key"`
	HttpAddress                                   *net.TCPAddr                   `opt:"http-address"`
	HttpGracePeriod                               time.Duration                  `opt:"http-grace-period"`
	HttpConfig                                    string                         `opt:"http-config"`
	LogFormat                                     log.Format                     `opt:"log.format"`
	LogLevel                                      log.Level                      `opt:"log.level"`
	QueryActiveQueryPath                          string                         `opt:"query.active-query-path"`
	QueryAutoDownsampling                         bool                           `opt:"query.auto-downsampling,noval"`
	QueryConnMetricLabel                          []string                       `opt:"query.conn-metric.label"`
	QueryDefaultEvaluationInterval                time.Duration                  `opt:"query.default-evaluation-interval"`
	QueryDefaultStep                              time.Duration                  `opt:"query.default-step"`
	QueryDefaultTenantID                          string                         `opt:"query.default-tenant-id"`
	QueryLookbackDelta                            time.Duration                  `opt:"query.lookback-delta"`
	QueryMaxConcurrent                            int                            `opt:"query.max-concurrent"`
	QueryMaxConcurrentSelect                      int                            `opt:"query.max-concurrent-select"`
	QueryMetadataDefaultTimeRange                 time.Duration                  `opt:"query.metadata.default-time-range"`
	QueryPartialResponse                          bool                           `opt:"query.partial-response,noval"`
	QueryPromQLEngine                             string                         `opt:"query.promql-engine"`
	QueryReplicaLabel                             []string                       `opt:"query.replica-label"`
	QueryTelemetryRequestDurationSecondsQuantiles []float64                      `opt:"query.telemetry.request-duration-seconds-quantiles"`
	QueryTelemetryRequestSamplesQuantiles         []float64                      `opt:"query.telemetry.request-samples-quantiles"`
	QueryTelemetryRequestSeriesSecondsQuantiles   []float64                      `opt:"query.telemetry.request-series-seconds-quantiles"`
	QueryTenantCertificateField                   string                         `opt:"query.tenant-certificate-field"`
	QueryTenantHeader                             string                         `opt:"query.tenant-header"`
	QueryTimeout                                  time.Duration                  `opt:"query.timeout"`
	RequestLoggingConfig                          *reqlogging.RequestConfig      `opt:"request.logging-config"`
	RequestLoggingConfigFile                      containeropts.ContainerUpdater `opt:"request.logging-config-file"`
	SelectorLabel                                 []string                       `opt:"selector-label"`
	StoreLimitsRequestSamples                     int                            `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries                      int                            `opt:"store.limits.request-series"`
	StoreResponseTimeout                          time.Duration                  `opt:"store.response-timeout"`
	StoreSDDNSInterval                            time.Duration                  `opt:"store.sd-dns-interval"`
	StoreSDFiles                                  []string                       `opt:"store.sd-files"`
	StoreSDInterval                               time.Duration                  `opt:"store.sd-interval"`
	StoreUnhealthyTimeout                         time.Duration                  `opt:"store.unhealthy-timeout"`
	TracingConfig                                 *trclient.TracingConfig        `opt:"tracing.config"`
	TracingConfigFile                             containeropts.ContainerUpdater `opt:"tracing.config-file"`
	WebDisableCORS                                bool                           `opt:"web.disable-cors,noval"`
	WebExternalPrefix                             string                         `opt:"web.external-prefix"`
	WebPrefixHeader                               string                         `opt:"web.prefix-header"`
	WebRoutePrefix                                string                         `opt:"web.route-prefix"`

	// Extra options not officially supported.
	cmdopt.ExtraOpts
}

type QueryDeployment struct {
	options *QueryOptions
	workload.DeploymentWorkload
}

func NewDefaultOptions() *QueryOptions {
	return &QueryOptions{
		LogLevel:  "warn",
		LogFormat: "logfmt",
	}
}

func NewQuery(opts *QueryOptions, namespace, imageTag string) *QueryDeployment {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-query",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "query-layer",
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
			Name:                 "observatorium-thanos-query",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("500m", "2", "1Gi", "8Gi"),
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

	return &QueryDeployment{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (q *QueryDeployment) Objects() []runtime.Object {
	container := q.makeContainer()
	return q.DeploymentWorkload.Objects(container)
}

func (q *QueryDeployment) makeContainer() *workload.Container {
	httpPort := kghelpers.GetPortOrDefault(defaultHTTPPort, q.options.HttpAddress)
	kghelpers.CheckProbePort(httpPort, q.LivenessProbe)
	kghelpers.CheckProbePort(httpPort, q.ReadinessProbe)

	grpcPort := kghelpers.GetPortOrDefault(defaultGRPCPort, q.options.GrpcAddress)

	ret := q.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"query"}, cmdopt.GetOpts(q.options)...)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "grpc",
			ContainerPort: int32(grpcPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http", httpPort, httpPort),
		kghelpers.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if q.options.RequestLoggingConfig != nil {
		q.options.RequestLoggingConfigFile.Update(ret)
	}

	if q.options.TracingConfigFile != nil {
		q.options.TracingConfigFile.Update(ret)
	}

	return ret
}
