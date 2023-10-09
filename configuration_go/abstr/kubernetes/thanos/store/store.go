package store

import (
	"net"
	"time"

	"github.com/bwplotka/mimic"
	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache"
	thanoslog "github.com/observatorium/observatorium/configuration_go/schemas/thanos/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	thanostime "github.com/observatorium/observatorium/configuration_go/schemas/thanos/time"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/units"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"
	corev1 "k8s.io/api/core/v1"
)

const (
	dataVolumeName   string = "data"
	defaultHTTPPort  int    = 10902
	defaultGRPCPort  int    = 10901
	defaultNamespace string = "observatorium"
	defaultImage     string = "quay.io/thanos/thanos"
	defaultImageTag  string = "v0.32.2"
)

// StoreOptions represents the options/flags for the store.
// See https://thanos.io/tip/components/store.md/#flags for details.
type StoreOptions struct {
	BlockMetaFetchConcurrency        int                             `opt:"block-meta-fetch-concurrency"`
	BlockSyncConcurrency             int                             `opt:"block-sync-concurrency"`
	BucketWebLabel                   string                          `opt:"bucket-web-label"`
	CacheIndexHeader                 bool                            `opt:"cache-index-header"`
	ChunkPoolSize                    units.Bytes                     `opt:"chunk-pool-size"`
	ConsistencyDelay                 time.Duration                   `opt:"consistency-delay"`
	DataDir                          string                          `opt:"data-dir"`
	GrpcAddress                      *net.TCPAddr                    `opt:"grpc-address"`
	GrpcGracePeriod                  time.Duration                   `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge       time.Duration                   `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert                string                          `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa            string                          `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey                 string                          `opt:"grpc-server-tls-key"`
	HttpAddress                      *net.TCPAddr                    `opt:"http-address"`
	HttpGracePeriod                  time.Duration                   `opt:"http-grace-period"`
	HttpConfig                       string                          `opt:"http.config"`
	IgnoreDeletionMarksDelay         time.Duration                   `opt:"ignore-deletion-marks-delay"`
	IndexCacheSize                   units.Bytes                     `opt:"index-cache-size"`
	IndexCacheConfig                 *cache.IndexCacheConfig         `opt:"index-cache.config"`
	IndexCacheConfigFile             string                          `opt:"index-cache.config-file"`
	LogFormat                        thanoslog.LogFormat             `opt:"log.format"`
	LogLevel                         thanoslog.LogLevel              `opt:"log.level"`
	MaxTime                          *thanostime.TimeOrDurationValue `opt:"max-time"`
	MinTime                          *thanostime.TimeOrDurationValue `opt:"min-time"`
	ObjstoreConfig                   string                          `opt:"objstore.config"`
	ObjstoreConfigFile               string                          `opt:"objstore.config-file"`
	RequestLoggingConfig             *reqlogging.RequestConfig       `opt:"request.logging-config"`
	RequestLoggingConfigFile         string                          `opt:"request.logging-config-file"`
	SelectorRelabelConfig            *relabel.Config                 `opt:"selector.relabel-config"`
	SelectorRelabelConfigFile        string                          `opt:"selector.relabel-config-file"`
	StoreEnableIndexHeaderLazyReader bool                            `opt:"store.enable-index-header-lazy-reader"`
	StoreEnableLazyExpandedPostings  bool                            `opt:"store.enable-lazy-expanded-postings"`
	StoreGrpcDownloadedBytesLimit    units.Bytes                     `opt:"store.grps.downloaded-bytes-limit"`
	StoreGrpcSeriesMaxConcurrency    int                             `opt:"store.grps.series-max-concurrency"`
	StoreLimitsRequestSamples        int                             `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries         int                             `opt:"store.limits.request-series"`
	SyncBlockDuration                time.Duration                   `opt:"sync-block-duration"`
	TracingConfig                    *trclient.TracingConfig         `opt:"tracing.config"`
	TracingConfigFile                string                          `opt:"tracing.config-file"`
	WebDisable                       bool                            `opt:"web.disable"`
	WebDisableCors                   bool                            `opt:"web.disable-cors"`
	WebExternalPrefix                string                          `opt:"web.external-prefix"`
	WebPrefixHeader                  string                          `opt:"web.prefix-header"`

	// Extra options not officially supported by the store.
	cmdopt.ExtraOpts
}

type StoreStatefulSet struct {
	Options    *StoreOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewStore() *StoreStatefulSet {
	opts := &StoreOptions{
		LogLevel:                 "warn",
		LogFormat:                "logfmt",
		DataDir:                  "/var/thanos/store",
		ObjstoreConfig:           "$(OBJSTORE_CONFIG)",
		IgnoreDeletionMarksDelay: 24 * time.Hour,
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-store",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "object-store-gateway",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &StoreStatefulSet{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                defaultImage,
			ImageTag:             defaultImageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-store",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("500m", "1", "200Mi", "400Mi"),
			Affinity:             *k8sutil.NewAntiAffinity(nil, labelSelectors),
			SecurityContext:      k8sutil.GetDefaultSecurityContext(),
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
				k8sutil.NewEnvFromSecret("OBJSTORE_CONFIG", "objectStore-secret", "thanos.yaml"),
				k8sutil.NewEnvFromField("HOST_IP_ADDRESS", "status.hostIP"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
		VolumeSize: "50Gi",
	}
}

func (s *StoreStatefulSet) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      s.Name,
		Labels:    s.CommonLabels,
		Namespace: s.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &s.TerminationGracePeriodSeconds,
		Affinity:                      &s.Affinity,
		SecurityContext:               s.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container}, s.Sidecars...),
	}

	statefulset := &k8sutil.StatefulSet{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   s.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"store-statefulSet": statefulset.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["store-service"] = service.MakeManifest()

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["store-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["store-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["store-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["store-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *StoreStatefulSet) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &StoreOptions{}
	}

	httpPort := defaultHTTPPort
	if s.Options.HttpAddress != nil && s.Options.HttpAddress.Port != 0 {
		httpPort = s.Options.HttpAddress.Port
	}

	grpcPort := defaultGRPCPort
	if s.Options.GrpcAddress != nil && s.Options.GrpcAddress.Port != 0 {
		grpcPort = s.Options.GrpcAddress.Port
	}

	livenessPort := s.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(httpPort) {
		mimic.Panicf(`liveness probe port %d does not match http port %d`, livenessPort, httpPort)
	}

	readinessPort := s.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(httpPort) {
		mimic.Panicf(`readiness probe port %d does not match http port %d`, readinessPort, httpPort)
	}

	if s.Options.DataDir == "" {
		mimic.Panicf(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"store"}, cmdopt.GetOpts(s.Options)...)
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
		k8sutil.NewServicePort("http", httpPort, httpPort),
		k8sutil.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}
	ret.VolumeClaims = []k8sutil.VolumeClaim{
		k8sutil.NewVolumeClaimProvider(dataVolumeName, s.VolumeType, s.VolumeSize),
	}
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: s.Options.DataDir,
		},
	}

	return ret
}
