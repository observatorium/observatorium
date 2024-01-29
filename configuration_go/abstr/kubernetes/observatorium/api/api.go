package api

import (
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultHTTPPublicPort   int = 8080
	defaultHTTPInternalPort int = 8081
	defaultGRPCPort         int = 8090
)

// NewRbacConfig returns a new RBAC config file option.
func NewRbacConfig(value *RBAC) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/observatorium/rbac", "config.yaml", "rbac-config", "observatorium-rbac")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewTenantsConfig returns a new tenants config file option.
func NewTenantsConfig(value *Tenants) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/observatorium/tenants", "config.yaml", "tenants", "observatorium-tenants")
	ret.AsSecret() // Is a secret by default.
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type ObservatoriumAPIOptions struct {
	DebugBlockProfileRate                       int                            `opt:"debug.block-profile-rate"`
	DebugMutexProfileRate                       int                            `opt:"debug.mutex-profile-fraction"`
	DebugName                                   string                         `opt:"debug.name"`
	GrpcListen                                  *net.TCPAddr                   `opt:"grpc.listen"`
	InternalTracingEndpoint                     string                         `opt:"internal.tracing.endpoint"`
	InternalTracingEndpointType                 string                         `opt:"internal.tracing.endpoint-type"`
	InternalTracingSamplingFraction             float64                        `opt:"internal.tracing.sampling-fraction"`
	InternalTracingServiceName                  string                         `opt:"internal.tracing.service-name"`
	LogFormat                                   log.Format                     `opt:"log.format"`
	LogLevel                                    log.Level                      `opt:"log.level"`
	LogsAuthExtractSelectors                    string                         `opt:"logs.auth.extract-selectors"`
	LogsReadEndpoint                            string                         `opt:"logs.read.endpoint"`
	LogsRulesEndpoint                           string                         `opt:"logs.rules.endpoint"`
	LogsRulesLabelFilters                       string                         `opt:"logs.rules.label-filters"`
	LogsRulesReadOnly                           bool                           `opt:"logs.rules.read-only,noval"`
	LogsRulesTenantLabel                        string                         `opt:"logs.rules.tenant-label"`
	LogsTailEndpoint                            string                         `opt:"logs.tail.endpoint"`
	LogsTenantHeader                            string                         `opt:"logs.tenant-header"`
	LogsTlsCaFile                               string                         `opt:"logs.tls.ca-file"`
	LogsTlsCertFile                             string                         `opt:"logs.tls.cert-file"`
	LogsTlsKeyFile                              string                         `opt:"logs.tls.key-file"`
	LogsWriteTimeout                            time.Duration                  `opt:"logs.write-timeout"`
	LogsWriteEndpoint                           string                         `opt:"logs.write.endpoint"`
	MetricsAlertmanagerEndpoint                 string                         `opt:"metrics.alertmanager.endpoint"`
	MetricsReadEndpoint                         string                         `opt:"metrics.read.endpoint"`
	MetricsRulesEndpoint                        string                         `opt:"metrics.rules.endpoint"`
	MetricsTenantHeader                         string                         `opt:"metrics.tenant-header"`
	MetricsTenantLabel                          string                         `opt:"metrics.tenant-label"`
	MetricsTlsCaFile                            string                         `opt:"metrics.tls.ca-file"`
	MetricsTlsCertFile                          string                         `opt:"metrics.tls.cert-file"`
	MetricsTlsKeyFile                           string                         `opt:"metrics.tls.key-file"`
	MetricsWriteTimeout                         time.Duration                  `opt:"metrics.write-timeout"`
	MetricsWriteEndpoint                        string                         `opt:"metrics.write.endpoint"`
	MiddlewareBacklogDurationConcurrentRequests time.Duration                  `opt:"middleware.backlog-duration-concurrent-requests"`
	MiddlewareBacklogLimitConcurrentRequests    int                            `opt:"middleware.backlog-limit-concurrent-requests"`
	MiddlewareConcurrentRequestLimit            int                            `opt:"middleware.concurrent-request-limit"`
	MiddlewareRateLimiterGrpcAddress            string                         `opt:"middleware.rate-limiter.grpc-address"`
	RbacConfig                                  containeropts.ContainerUpdater `opt:"rbac.config"`
	ServerReadHeaderTimeout                     time.Duration                  `opt:"server.read-header-timeout"`
	ServerReadTimeout                           time.Duration                  `opt:"server.read-timeout"`
	ServerWriteTimeout                          time.Duration                  `opt:"server.write-timeout"`
	TenantsConfig                               containeropts.ContainerUpdater `opt:"tenants.config"`
	TlsCipherSuites                             string                         `opt:"tls.cipher-suites"`
	TlsClientAuthType                           string                         `opt:"tls.client-auth-type"`
	TlsHealthchecksServerCaFile                 string                         `opt:"tls.healthchecks.server-ca-file"`
	TlsHealthchecksServerName                   string                         `opt:"tls.healthchecks.server-name"`
	TlsInternalServerCertFile                   string                         `opt:"tls.internal.server.cert-file"`
	TlsInternalServerKeyFile                    string                         `opt:"tls.internal.server.key-file"`
	TlsMaxVersion                               string                         `opt:"tls.max-version"`
	TlsMinVersion                               string                         `opt:"tls.min-version"`
	TlsReloadInterval                           time.Duration                  `opt:"tls.reload-interval"`
	TlsServerCertFile                           string                         `opt:"tls.server.cert-file"`
	TlsServerKeyFile                            string                         `opt:"tls.server.key-file"`
	TracesReadEndpoint                          string                         `opt:"traces.read.endpoint"`
	TracesTempoEndpoint                         string                         `opt:"traces.tempo.endpoint"`
	TracesTenantHeader                          string                         `opt:"traces.tenant-header"`
	TracesTlsCaFile                             string                         `opt:"traces.tls.ca-file"`
	TracesTlsCertFile                           string                         `opt:"traces.tls.cert-file"`
	TracesTlsKeyFile                            string                         `opt:"traces.tls.key-file"`
	TracesWriteTimeout                          time.Duration                  `opt:"traces.write-timeout"`
	TracesWriteEndpoint                         string                         `opt:"traces.write.endpoint"`
	WebHealthchecksURL                          string                         `opt:"web.healthchecks.url"`
	WebInternalListen                           *net.TCPAddr                   `opt:"web.internal.listen"`
	WebListen                                   *net.TCPAddr                   `opt:"web.listen"`

	// For setting extra options not listed above.
	cmdopt.ExtraOpts
}

type ObservatoriumAPIDeployment struct {
	options *ObservatoriumAPIOptions
	workload.DeploymentWorkload
}

func NewObservatoriumAPI(opts *ObservatoriumAPIOptions, namespace, imageTag string) *ObservatoriumAPIDeployment {
	if opts == nil {
		opts = &ObservatoriumAPIOptions{}
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "observatorium-api",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "api",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	httpInternalPort := kghelpers.GetPortOrDefault(defaultHTTPInternalPort, opts.WebInternalListen)

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Image:                "quay.io/observatorium/api",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-api",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("100m", "1", "1Gi", "4Gi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: kghelpers.NewProbe("/live", httpInternalPort, kghelpers.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: kghelpers.NewProbe("/ready", httpInternalPort, kghelpers.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}

	return &ObservatoriumAPIDeployment{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (o *ObservatoriumAPIDeployment) Objects() []runtime.Object {
	container := o.makeContainer()
	return o.DeploymentWorkload.Objects(container)
}

func (o *ObservatoriumAPIDeployment) makeContainer() *workload.Container {
	httpPublicPort := kghelpers.GetPortOrDefault(defaultHTTPPublicPort, o.options.WebListen)
	httpInternalPort := kghelpers.GetPortOrDefault(defaultHTTPInternalPort, o.options.WebInternalListen)
	grpcPort := kghelpers.GetPortOrDefault(defaultGRPCPort, o.options.GrpcListen)

	kghelpers.CheckProbePort(httpInternalPort, o.LivenessProbe)
	kghelpers.CheckProbePort(httpInternalPort, o.ReadinessProbe)

	ret := o.ToContainer()
	ret.Name = "observatorium-api"
	ret.Args = cmdopt.GetOpts(o.options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http-public",
			ContainerPort: int32(httpPublicPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "http-internal",
			ContainerPort: int32(httpInternalPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "grpc",
			ContainerPort: int32(grpcPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http-public", httpPublicPort, httpPublicPort),
		kghelpers.NewServicePort("http-internal", httpInternalPort, httpInternalPort),
		kghelpers.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http-internal",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if o.options.RbacConfig != nil {
		o.options.RbacConfig.Update(ret)
	}

	if o.options.TenantsConfig != nil {
		o.options.TenantsConfig.Update(ret)
	}

	return ret
}
