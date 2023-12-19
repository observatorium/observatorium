package api

import (
	"fmt"
	"net"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/option"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultNamespace        string = "observatorium"
	defaultHTTPPublicPort   int    = 8080
	defaultHTTPInternalPort int    = 8081
	defaultGRPCPort         int    = 8090
)

type rbacConfig = option.ConfigFile[string]

func NewRbacConfig(name string, value string) *rbacConfig {
	return option.NewConfigFile("/etc/observatorium/rbac", "config.yaml", name, value)
}

type tenantsConfig = k8sutil.ConfigFileWithType[Tenants]

func NewTenantsConfig(name string, value Tenants) *tenantsConfig {
	return k8sutil.NewConfigFileWithType[Tenants]("/etc/observatorium/tenants", "config.yaml", "tenants", "observatorium-tenants")
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
	LogsWriteTimeout                            model.Duration `opt:"logs.write-timeout"`
	LogsWriteEndpoint                           string         `opt:"logs.write.endpoint"`
	MetricsAlertmanagerEndpoint                 string         `opt:"metrics.alertmanager.endpoint"`
	MetricsReadEndpoint                         string         `opt:"metrics.read.endpoint"`
	MetricsRulesEndpoint                        string         `opt:"metrics.rules.endpoint"`
	MetricsTenantHeader                         string         `opt:"metrics.tenant-header"`
	MetricsTenantLabel                          string         `opt:"metrics.tenant-label"`
	MetricsTlsCaFile                            string         `opt:"metrics.tls.ca-file"`
	MetricsTlsCertFile                          string         `opt:"metrics.tls.cert-file"`
	MetricsTlsKeyFile                           string         `opt:"metrics.tls.key-file"`
	MetricsWriteTimeout                         model.Duration `opt:"metrics.write-timeout"`
	MetricsWriteEndpoint                        string         `opt:"metrics.write.endpoint"`
	MiddlewareBacklogDurationConcurrentRequests model.Duration `opt:"middleware.backlog-duration-concurrent-requests"`
	MiddlewareBacklogLimitConcurrentRequests    int            `opt:"middleware.backlog-limit-concurrent-requests"`
	MiddlewareConcurrentRequestLimit            int            `opt:"middleware.concurrent-request-limit"`
	MiddlewareRateLimiterGrpcAddress            string         `opt:"middleware.rate-limiter.grpc-address"`
	RbacConfig                                  *rbacConfig    `opt:"rbac.config"`
	ServerReadHeaderTimeout                     model.Duration `opt:"server.read-header-timeout"`
	ServerReadTimeout                           model.Duration `opt:"server.read-timeout"`
	ServerWriteTimeout                          model.Duration `opt:"server.write-timeout"`
	TenantsConfig                               *tenantsConfig `opt:"tenants.config"`
	TlsCipherSuites                             string         `opt:"tls.cipher-suites"`
	TlsClientAuthType                           string         `opt:"tls.client-auth-type"`
	TlsHealthchecksServerCaFile                 string         `opt:"tls.healthchecks.server-ca-file"`
	TlsHealthchecksServerName                   string         `opt:"tls.healthchecks.server-name"`
	TlsInternalServerCertFile                   string         `opt:"tls.internal.server.cert-file"`
	TlsInternalServerKeyFile                    string         `opt:"tls.internal.server.key-file"`
	TlsMaxVersion                               string         `opt:"tls.max-version"`
	TlsMinVersion                               string         `opt:"tls.min-version"`
	TlsReloadInterval                           model.Duration `opt:"tls.reload-interval"`
	TlsServerCertFile                           string         `opt:"tls.server.cert-file"`
	TlsServerKeyFile                            string         `opt:"tls.server.key-file"`
	TracesReadEndpoint                          string         `opt:"traces.read.endpoint"`
	TracesTempoEndpoint                         string         `opt:"traces.tempo.endpoint"`
	TracesTenantHeader                          string         `opt:"traces.tenant-header"`
	TracesTlsCaFile                             string         `opt:"traces.tls.ca-file"`
	TracesTlsCertFile                           string         `opt:"traces.tls.cert-file"`
	TracesTlsKeyFile                            string         `opt:"traces.tls.key-file"`
	TracesWriteTimeout                          model.Duration `opt:"traces.write-timeout"`
	TracesWriteEndpoint                         string         `opt:"traces.write.endpoint"`
	WebHealthchecksURL                          string         `opt:"web.healthchecks.url"`
	WebInternalListen                           *net.TCPAddr   `opt:"web.internal.listen"`
	WebListen                                   *net.TCPAddr   `opt:"web.listen"`
}

type ObservatoriumAPIDeployment struct {
	Options *ObservatoriumAPIOptions
	k8sutil.DeploymentGenericConfig
}

func NewObservatoriumAPI() *ObservatoriumAPIDeployment {
	opts := &ObservatoriumAPIOptions{}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "observatorium-api",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "api",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &ObservatoriumAPIDeployment{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/observatorium/api",
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-api",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("100m", "1", "1Gi", "4Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/live", defaultHTTPInternalPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/ready", defaultHTTPInternalPort, k8sutil.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				k8sutil.NewEnvFromField("POD_NAME", "metadata.name"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}
}

func (s *ObservatoriumAPIDeployment) Manifests() k8sutil.ObjectMap {
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

	deployment := &k8sutil.Deployment{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   s.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"obs-api-statefulSet": deployment.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["obs-api-service"] = service.MakeManifest()

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["obs-api-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["obs-api-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["obs-api-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["obs-api-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *ObservatoriumAPIDeployment) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &ObservatoriumAPIOptions{}
	}

	httpPublicPort := defaultHTTPPublicPort
	if s.Options.WebListen != nil && s.Options.WebListen.Port != 0 {
		httpPublicPort = s.Options.WebListen.Port
	}

	httpInternalPort := defaultHTTPInternalPort
	if s.Options.WebInternalListen != nil && s.Options.WebInternalListen.Port != 0 {
		httpInternalPort = s.Options.WebInternalListen.Port
	}

	grpcPort := defaultGRPCPort
	if s.Options.GrpcListen != nil && s.Options.GrpcListen.Port != 0 {
		grpcPort = s.Options.GrpcListen.Port
	}

	livenessPort := s.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(httpInternalPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, httpInternalPort))
	}

	readinessPort := s.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(httpInternalPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, httpInternalPort))
	}

	ret := s.ToContainer()
	ret.Name = "observatorium-api"
	ret.Args = cmdopt.GetOpts(s.Options)
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

	if s.Options.RbacConfig != nil {
		ret.ConfigMaps[s.Options.RbacConfig.Name] = map[string]string{
			s.Options.RbacConfig.FileName(): s.Options.RbacConfig.Value,
		}

		ret.Volumes = append(ret.Volumes, k8sutil.NewPodVolumeFromConfigMap("rbac-config", s.Options.RbacConfig.FileName()))

		ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
			Name:      "rbac-config",
			MountPath: s.Options.RbacConfig.MountPath(),
			ReadOnly:  true,
		})
	}

	if s.Options.TenantsConfig != nil {
		s.Options.TenantsConfig.AddToContainer(ret)
	}

	return ret
}
