package api

import (
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultHTTPPublicPort   int = 8080
	defaultHTTPInternalPort int = 8081
	defaultGRPCPort         int = 8090
)

type rbacConfig = k8sutil.ConfigFile

func NewRbacConfig(value *RBAC) *rbacConfig {
	ret := k8sutil.NewConfigFile("/etc/observatorium/rbac", "config.yaml", "rbac-config", "observatorium-rbac")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type tenantsConfig = k8sutil.ConfigFile

func NewTenantsConfig(value *Tenants) *tenantsConfig {
	ret := k8sutil.NewConfigFile("/etc/observatorium/tenants", "config.yaml", "tenants", "observatorium-tenants")
	ret.AsSecret() // Is a secret by default.
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type ObservatoriumAPIOptions struct {
	DebugBlockProfileRate                       int            `opt:"debug.block-profile-rate"`
	DebugMutexProfileRate                       int            `opt:"debug.mutex-profile-fraction"`
	DebugName                                   string         `opt:"debug.name"`
	GrpcListen                                  *net.TCPAddr   `opt:"grpc.listen"`
	InternalTracingEndpoint                     string         `opt:"internal.tracing.endpoint"`
	InternalTracingEndpointType                 string         `opt:"internal.tracing.endpoint-type"`
	InternalTracingSamplingFraction             float64        `opt:"internal.tracing.sampling-fraction"`
	InternalTracingServiceName                  string         `opt:"internal.tracing.service-name"`
	LogFormat                                   log.LogFormat  `opt:"log.format"`
	LogLevel                                    log.LogLevel   `opt:"log.level"`
	LogsAuthExtractSelectors                    string         `opt:"logs.auth.extract-selectors"`
	LogsReadEndpoint                            string         `opt:"logs.read.endpoint"`
	LogsRulesEndpoint                           string         `opt:"logs.rules.endpoint"`
	LogsRulesLabelFilters                       string         `opt:"logs.rules.label-filters"`
	LogsRulesReadOnly                           bool           `opt:"logs.rules.read-only,noval"`
	LogsRulesTenantLabel                        string         `opt:"logs.rules.tenant-label"`
	LogsTailEndpoint                            string         `opt:"logs.tail.endpoint"`
	LogsTenantHeader                            string         `opt:"logs.tenant-header"`
	LogsTlsCaFile                               string         `opt:"logs.tls.ca-file"`
	LogsTlsCertFile                             string         `opt:"logs.tls.cert-file"`
	LogsTlsKeyFile                              string         `opt:"logs.tls.key-file"`
	LogsWriteTimeout                            time.Duration  `opt:"logs.write-timeout"`
	LogsWriteEndpoint                           string         `opt:"logs.write.endpoint"`
	MetricsAlertmanagerEndpoint                 string         `opt:"metrics.alertmanager.endpoint"`
	MetricsReadEndpoint                         string         `opt:"metrics.read.endpoint"`
	MetricsRulesEndpoint                        string         `opt:"metrics.rules.endpoint"`
	MetricsTenantHeader                         string         `opt:"metrics.tenant-header"`
	MetricsTenantLabel                          string         `opt:"metrics.tenant-label"`
	MetricsTlsCaFile                            string         `opt:"metrics.tls.ca-file"`
	MetricsTlsCertFile                          string         `opt:"metrics.tls.cert-file"`
	MetricsTlsKeyFile                           string         `opt:"metrics.tls.key-file"`
	MetricsWriteTimeout                         time.Duration  `opt:"metrics.write-timeout"`
	MetricsWriteEndpoint                        string         `opt:"metrics.write.endpoint"`
	MiddlewareBacklogDurationConcurrentRequests time.Duration  `opt:"middleware.backlog-duration-concurrent-requests"`
	MiddlewareBacklogLimitConcurrentRequests    int            `opt:"middleware.backlog-limit-concurrent-requests"`
	MiddlewareConcurrentRequestLimit            int            `opt:"middleware.concurrent-request-limit"`
	MiddlewareRateLimiterGrpcAddress            string         `opt:"middleware.rate-limiter.grpc-address"`
	RbacConfig                                  *rbacConfig    `opt:"rbac.config"`
	ServerReadHeaderTimeout                     time.Duration  `opt:"server.read-header-timeout"`
	ServerReadTimeout                           time.Duration  `opt:"server.read-timeout"`
	ServerWriteTimeout                          time.Duration  `opt:"server.write-timeout"`
	TenantsConfig                               *tenantsConfig `opt:"tenants.config"`
	TlsCipherSuites                             string         `opt:"tls.cipher-suites"`
	TlsClientAuthType                           string         `opt:"tls.client-auth-type"`
	TlsHealthchecksServerCaFile                 string         `opt:"tls.healthchecks.server-ca-file"`
	TlsHealthchecksServerName                   string         `opt:"tls.healthchecks.server-name"`
	TlsInternalServerCertFile                   string         `opt:"tls.internal.server.cert-file"`
	TlsInternalServerKeyFile                    string         `opt:"tls.internal.server.key-file"`
	TlsMaxVersion                               string         `opt:"tls.max-version"`
	TlsMinVersion                               string         `opt:"tls.min-version"`
	TlsReloadInterval                           time.Duration  `opt:"tls.reload-interval"`
	TlsServerCertFile                           string         `opt:"tls.server.cert-file"`
	TlsServerKeyFile                            string         `opt:"tls.server.key-file"`
	TracesReadEndpoint                          string         `opt:"traces.read.endpoint"`
	TracesTempoEndpoint                         string         `opt:"traces.tempo.endpoint"`
	TracesTenantHeader                          string         `opt:"traces.tenant-header"`
	TracesTlsCaFile                             string         `opt:"traces.tls.ca-file"`
	TracesTlsCertFile                           string         `opt:"traces.tls.cert-file"`
	TracesTlsKeyFile                            string         `opt:"traces.tls.key-file"`
	TracesWriteTimeout                          time.Duration  `opt:"traces.write-timeout"`
	TracesWriteEndpoint                         string         `opt:"traces.write.endpoint"`
	WebHealthchecksURL                          string         `opt:"web.healthchecks.url"`
	WebInternalListen                           *net.TCPAddr   `opt:"web.internal.listen"`
	WebListen                                   *net.TCPAddr   `opt:"web.listen"`

	// For setting extra options not listed above.
	cmdopt.ExtraOpts
}

type ObservatoriumAPIDeployment struct {
	options *ObservatoriumAPIOptions
	k8sutil.DeploymentGenericConfig
}

func NewObservatoriumAPI(opts *ObservatoriumAPIOptions, namespace, imageTag string) *ObservatoriumAPIDeployment {
	if opts == nil {
		opts = &ObservatoriumAPIOptions{}
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "observatorium-api",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "api",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	httpInternalPort := k8sutil.GetPortOrDefault(defaultHTTPInternalPort, opts.WebInternalListen)

	return &ObservatoriumAPIDeployment{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/observatorium/api",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-api",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("100m", "1", "1Gi", "4Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/live", httpInternalPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/ready", httpInternalPort, k8sutil.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}
}

func (s *ObservatoriumAPIDeployment) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()
	ret := k8sutil.ObjectMap{}
	ret.AddAll(s.GenerateObjectsDeployment(container))

	return ret
}

func (s *ObservatoriumAPIDeployment) makeContainer() *k8sutil.Container {
	httpPublicPort := k8sutil.GetPortOrDefault(defaultHTTPPublicPort, s.options.WebListen)
	httpInternalPort := k8sutil.GetPortOrDefault(defaultHTTPInternalPort, s.options.WebInternalListen)
	grpcPort := k8sutil.GetPortOrDefault(defaultGRPCPort, s.options.GrpcListen)

	k8sutil.CheckProbePort(httpInternalPort, s.LivenessProbe)
	k8sutil.CheckProbePort(httpInternalPort, s.ReadinessProbe)

	ret := s.ToContainer()
	ret.Name = "observatorium-api"
	ret.Args = cmdopt.GetOpts(s.options)
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
		k8sutil.NewServicePort("http-public", httpPublicPort, httpPublicPort),
		k8sutil.NewServicePort("http-internal", httpInternalPort, httpInternalPort),
		k8sutil.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http-internal",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if s.options.RbacConfig != nil {
		s.options.RbacConfig.AddToContainer(ret)
	}

	if s.options.TenantsConfig != nil {
		s.options.TenantsConfig.AddToContainer(ret)
	}

	return ret
}
