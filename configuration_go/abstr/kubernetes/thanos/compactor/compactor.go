package compactor

import (
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	thanostime "github.com/observatorium/observatorium/configuration_go/schemas/thanos/time"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	dataVolumeName  string = "data"
	defaultHTTPPort int    = 10902
)

// CompactorOptions represents the options/flags for the compactor.
// See https://thanos.io/tip/components/compact.md/#flags for details.
type CompactorOptions struct {
	BlockFilesConcurrency              int                             `opt:"block-files-concurrency"`
	BlockMetaFetchConcurrency          int                             `opt:"block-meta-fetch-concurrency"`
	BlockViewerGlobalSyncBlockInterval time.Duration                   `opt:"block-viewer.global.sync-block-interval"`
	BlockViewerGlobalSyncBlockTimeout  time.Duration                   `opt:"block-viewer.global.sync-block-timeout"`
	BucketWebLabel                     string                          `opt:"bucket-web-label"`
	CompactBlocksFetchConcurrency      int                             `opt:"compact.blocks-fetch-concurrency"`
	CompactCleanupInterval             time.Duration                   `opt:"compact.cleanup-interval"`
	CompactConcurrency                 int                             `opt:"compact.concurrency"`
	CompactProgressInterval            time.Duration                   `opt:"compact.progress-interval"`
	ConsistencyDelay                   time.Duration                   `opt:"consistency-delay"`
	DataDir                            string                          `opt:"data-dir"`
	DeduplicationFunc                  string                          `opt:"deduplication.func"`
	DeduplicationReplicaLabel          string                          `opt:"deduplication.replica-label"`
	DeleteDelay                        time.Duration                   `opt:"delete-delay"`
	DownsampleConcurrency              int                             `opt:"downsample.concurrency"`
	DownsamplingDisable                bool                            `opt:"downsampling.disable"`
	HashFunc                           string                          `opt:"hash-func"`
	HttpAddress                        *net.TCPAddr                    `opt:"http-address"`
	HttpGracePeriod                    time.Duration                   `opt:"http-grace-period"`
	HttpConfig                         string                          `opt:"http.config"`
	LogFormat                          log.Format                      `opt:"log.format"`
	LogLevel                           log.Level                       `opt:"log.level"`
	MaxTime                            *thanostime.TimeOrDurationValue `opt:"max-time"`
	MinTime                            *thanostime.TimeOrDurationValue `opt:"min-time"`
	ObjstoreConfig                     string                          `opt:"objstore.config"`
	ObjstoreConfigFile                 string                          `opt:"objstore.config-file"`
	RetentionResolution1h              time.Duration                   `opt:"retention.resolution-1h"`
	RetentionResolution5m              time.Duration                   `opt:"retention.resolution-5m"`
	RetentionResolutionRaw             time.Duration                   `opt:"retention.resolution-raw"`
	SelectorRelabelConfig              string                          `opt:"selector.relabel-config"`
	SelectorRelabelConfigFile          string                          `opt:"selector.relabel-config-file"`
	TracingConfig                      string                          `opt:"tracing.config"`
	TracingConfigFile                  string                          `opt:"tracing.config-file"`
	Version                            bool                            `opt:"version,noval"`
	Wait                               bool                            `opt:"wait,noval"`
	WaitInterval                       time.Duration                   `opt:"wait-interval"`
	WebDisable                         bool                            `opt:"web.disable"`
	WebDisableCors                     bool                            `opt:"web.disable-cors"`
	WebExternalPrefix                  string                          `opt:"web.external-prefix"`
	WebPrefixHeader                    string                          `opt:"web.prefix-header"`
	WebRoutePrefix                     string                          `opt:"web.route-prefix"`

	// Extra options not officially supported by the compactor.
	cmdopt.ExtraOpts
}

type CompactorStatefulSet struct {
	options *CompactorOptions
	workload.StatefulSetWorkload
}

func NewDefaultOptions() *CompactorOptions {
	return &CompactorOptions{
		ObjstoreConfig:            "$(OBJSTORE_CONFIG)",
		Wait:                      true,
		LogLevel:                  "warn",
		LogFormat:                 "logfmt",
		DataDir:                   "/var/thanos/compactor",
		RetentionResolutionRaw:    time.Hour * 24 * 365,
		DeleteDelay:               time.Hour * 24 * 2,
		CompactConcurrency:        1,
		DownsampleConcurrency:     1,
		DeduplicationReplicaLabel: "replica",
	}
}

// NewCompactor returns a new compactor statefulset with default values.
// It allows generating the all the manifests for the compactor.
func NewCompactor(opts *CompactorOptions, namespace, imageTag string) *CompactorStatefulSet {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-compact",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "database-compactor",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	ssWorkload := workload.StatefulSetWorkload{
		Replicas:   1,
		VolumeSize: "50Gi",
		PodConfig: workload.PodConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-compact",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("2", "3", "2000Mi", "3000Mi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,
			LivenessProbe: kghelpers.NewProbe("/-/healthy", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 4,
				PeriodSeconds:    30,
			}),
			ReadinessProbe: kghelpers.NewProbe("/-/ready", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				kghelpers.NewEnvFromSecret("OBJSTORE_CONFIG", "objectStore-secret", "thanos.yaml"),
				kghelpers.NewEnvFromField("HOST_IP_ADDRESS", "status.hostIP"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}

	return &CompactorStatefulSet{
		options:             opts,
		StatefulSetWorkload: ssWorkload,
	}
}

// Manifests returns the manifests for the compactor.
// It includes the statefulset, the service, the service monitor, the service account and the config maps required by the containers.
func (c *CompactorStatefulSet) Objects() []runtime.Object {
	container := c.makeContainer()
	return c.StatefulSetWorkload.Objects(container)
}

func (c *CompactorStatefulSet) makeContainer() *workload.Container {
	httpPort := kghelpers.GetPortOrDefault(defaultHTTPPort, c.options.HttpAddress)
	kghelpers.CheckProbePort(httpPort, c.LivenessProbe)
	kghelpers.CheckProbePort(httpPort, c.ReadinessProbe)

	// Print warning if data directory is not specified.
	if c.options.DataDir == "" {
		panic("data directory is not specified for the statefulset.")
	}

	ret := c.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"compact"}, cmdopt.GetOpts(c.options)...)
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
	ret.VolumeClaims = append(ret.VolumeClaims, workload.PersistentVolumeClaim{
		Name:  dataVolumeName,
		Size:  c.VolumeSize,
		Class: c.VolumeType,
	})
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: c.options.DataDir,
		},
	}

	return ret
}
