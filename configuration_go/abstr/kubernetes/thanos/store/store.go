package store

import (
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/cache"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	thanostime "github.com/observatorium/observatorium/configuration_go/schemas/thanos/time"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/units"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"
	corev1 "k8s.io/api/core/v1"
)

const (
	dataVolumeName  string = "data"
	defaultHTTPPort int    = 10902
	defaultGRPCPort int    = 10901
)

// StoreOptions represents the options/flags for the store.
// See https://thanos.io/tip/components/store.md/#flags for details.
type StoreOptions struct {
	BlockMetaFetchConcurrency        int                             `opt:"block-meta-fetch-concurrency"`
	BlockSyncConcurrency             int                             `opt:"block-sync-concurrency"`
	BucketWebLabel                   string                          `opt:"bucket-web-label"`
	CacheIndexHeader                 bool                            `opt:"cache-index-header,noval"`
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
	LogFormat                        log.LogFormat                   `opt:"log.format"`
	LogLevel                         log.LogLevel                    `opt:"log.level"`
	MaxTime                          *thanostime.TimeOrDurationValue `opt:"max-time"`
	MinTime                          *thanostime.TimeOrDurationValue `opt:"min-time"`
	ObjstoreConfig                   string                          `opt:"objstore.config"`
	ObjstoreConfigFile               string                          `opt:"objstore.config-file"`
	RequestLoggingConfig             *reqlogging.RequestConfig       `opt:"request.logging-config"`
	RequestLoggingConfigFile         string                          `opt:"request.logging-config-file"`
	SelectorRelabelConfig            *relabel.Config                 `opt:"selector.relabel-config"`
	SelectorRelabelConfigFile        string                          `opt:"selector.relabel-config-file"`
	StoreEnableIndexHeaderLazyReader bool                            `opt:"store.enable-index-header-lazy-reader,noval"`
	StoreEnableLazyExpandedPostings  bool                            `opt:"store.enable-lazy-expanded-postings,noval"`
	StoreGrpcDownloadedBytesLimit    units.Bytes                     `opt:"store.grps.downloaded-bytes-limit"`
	StoreGrpcSeriesMaxConcurrency    int                             `opt:"store.grps.series-max-concurrency"`
	StoreLimitsRequestSamples        int                             `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries         int                             `opt:"store.limits.request-series"`
	SyncBlockDuration                time.Duration                   `opt:"sync-block-duration"`
	TracingConfig                    *trclient.TracingConfig         `opt:"tracing.config"`
	TracingConfigFile                string                          `opt:"tracing.config-file"`
	WebDisable                       bool                            `opt:"web.disable,noval"`
	WebDisableCors                   bool                            `opt:"web.disable-cors,noval"`
	WebExternalPrefix                string                          `opt:"web.external-prefix"`
	WebPrefixHeader                  string                          `opt:"web.prefix-header"`

	// Extra options not officially supported by the store.
	cmdopt.ExtraOpts
}

type StoreStatefulSet struct {
	options    *StoreOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewDefaultOptions() *StoreOptions {
	return &StoreOptions{
		LogLevel:                 "warn",
		LogFormat:                "logfmt",
		DataDir:                  "/var/thanos/store",
		ObjstoreConfig:           "$(OBJSTORE_CONFIG)",
		IgnoreDeletionMarksDelay: 24 * time.Hour,
	}
}

func NewStore(opts *StoreOptions, namespace, imageTag string) *StoreStatefulSet {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-store",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "object-store-gateway",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	return &StoreStatefulSet{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-store",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("500m", "1", "200Mi", "400Mi"),
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

	ret := k8sutil.ObjectMap{}
	ret.AddAll(s.GenerateObjectsStatefulSet(container))

	return ret
}

func (s *StoreStatefulSet) makeContainer() *k8sutil.Container {
	httpPort := k8sutil.GetPortOrDefault(defaultHTTPPort, s.options.HttpAddress)
	k8sutil.CheckProbePort(httpPort, s.LivenessProbe)
	k8sutil.CheckProbePort(httpPort, s.ReadinessProbe)

	grpcPort := k8sutil.GetPortOrDefault(defaultGRPCPort, s.options.GrpcAddress)

	if s.options.DataDir == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"store"}, cmdopt.GetOpts(s.options)...)
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
			MountPath: s.options.DataDir,
		},
	}

	return ret
}
