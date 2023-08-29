package compactor

import (
	"log"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	k8sutilv2 "github.com/observatorium/observatorium/configuration_go/k8sutil/v2"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	dataVolumeName   string = "data"
	defaultHTTPPort  int    = 10902
	defaultNamespace string = "observatorium"
	servicePortName  string = "http"
	defaultImage     string = "quay.io/thanos/thanos"
	defaultImageTag  string = "v0.31.0"
)

// CompactorOptions represents the options/flags for the compactor.
// See https://thanos.io/tip/components/compact.md/#flags for details.
type CompactorOptions struct {
	BlockFilesConcurrency              int           `opt:"block-files-concurrency"`
	BlockMetaFetchConcurrency          int           `opt:"block-meta-fetch-concurrency"`
	BlockViewerGlobalSyncBlockInterval time.Duration `opt:"block-viewer.global.sync-block-interval"`
	BlockViewerGlobalSyncBlockTimeout  time.Duration `opt:"block-viewer.global.sync-block-timeout"`
	BucketWebLabel                     string        `opt:"bucket-web-label"`
	CompactBlocksFetchConcurrency      int           `opt:"compact.blocks-fetch-concurrency"`
	CompactCleanupInterval             time.Duration `opt:"compact.cleanup-interval"`
	CompactConcurrency                 int           `opt:"compact.concurrency"`
	CompactProgressInterval            time.Duration `opt:"compact.progress-interval"`
	ConsistencyDelay                   time.Duration `opt:"consistency-delay"`
	DataDir                            string        `opt:"data-dir"`
	DeduplicationFunc                  string        `opt:"deduplication.func"`
	DeduplicationReplicaLabel          string        `opt:"deduplication.replica-label"`
	DeleteDelay                        time.Duration `opt:"delete-delay"`
	DownsampleConcurrency              int           `opt:"downsample.concurrency"`
	DownsamplingDisable                bool          `opt:"downsampling.disable"`
	HashFunc                           string        `opt:"hash-func"`
	HttpAddress                        net.TCPAddr   `opt:"http-address"`
	HttpGracePeriod                    time.Duration `opt:"http-grace-period"`
	HttpConfig                         string        `opt:"http.config"`
	LogFormat                          string        `opt:"log.format"`
	LogLevel                           string        `opt:"log.level"`
	MaxTime                            string        `opt:"max-time"`
	MinTime                            string        `opt:"min-time"`
	ObjstoreConfig                     string        `opt:"objstore.config"`
	ObjstoreConfigFile                 string        `opt:"objstore.config-file"`
	RetentionResolution1h              time.Duration `opt:"retention.resolution-1h"`
	RetentionResolution5m              time.Duration `opt:"retention.resolution-5m"`
	RetentionResolutionRaw             time.Duration `opt:"retention.resolution-raw"`
	SelectorRelabelConfig              string        `opt:"selector.relabel-config"`
	SelectorRelabelConfigFile          string        `opt:"selector.relabel-config-file"`
	TracingConfig                      string        `opt:"tracing.config"`
	TracingConfigFile                  string        `opt:"tracing.config-file"`
	Version                            bool          `opt:"version,noval"`
	Wait                               bool          `opt:"wait,noval"`
	WaitInterval                       time.Duration `opt:"wait-interval"`
	WebDisable                         bool          `opt:"web.disable"`
	WebDisableCors                     bool          `opt:"web.disable-cors"`
	WebExternalPrefix                  string        `opt:"web.external-prefix"`
	WebPrefixHeader                    string        `opt:"web.prefix-header"`
	WebRoutePrefix                     string        `opt:"web.route-prefix"`

	// Extra options not officially supported by the compactor.
	cmdopt.ExtraOpts
}

// DefaultMetaConfig returns the default meta configuration for the compactor.
func DefaultMetaConfig() k8sutilv2.MetaConfig {
	return k8sutilv2.MetaConfig{
		Name:      "observatorium-thanos-compact",
		Namespace: defaultNamespace,
		Labels: map[string]string{
			k8sutil.NameLabel:      "thanos-compact",
			k8sutil.InstanceLabel:  "observatorium",
			k8sutil.PartOfLabel:    "observatorium",
			k8sutil.ComponentLabel: "database-compactor",
			k8sutil.VersionLabel:   defaultImageTag,
		},
	}
}

// NewBaseContainerProvider returns a new container provider for the compactor.
func NewBaseContainerProvider(c *CompactorOptions) *k8sutilv2.Container {
	if c == nil {
		c = &CompactorOptions{}
	}

	httpPort := defaultHTTPPort
	if c.HttpAddress.Port != 0 {
		httpPort = c.HttpAddress.Port
	}

	livenessProbeCfg := k8sutilv2.ProbeConfig{
		FailureThreshold: 4,
		PeriodSeconds:    30,
	}
	readinessProbeCfg := k8sutilv2.ProbeConfig{
		FailureThreshold: 20,
		PeriodSeconds:    5,
	}

	return &k8sutilv2.Container{
		Image:           defaultImage,
		ImageTag:        defaultImageTag,
		Name:            "thanos",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args:            append([]string{"compact"}, cmdopt.GetOpts(c)...),
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: int32(httpPort),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe:  k8sutilv2.NewProbe("/-/healthy", httpPort, livenessProbeCfg),
		ReadinessProbe: k8sutilv2.NewProbe("/-/ready", httpPort, readinessProbeCfg),
		Resources:      k8sutilv2.NewResourcesRequirements("2", "3", "2000Mi", "3000Mi"),

		ServicePorts: []corev1.ServicePort{
			k8sutilv2.NewServicePort("http", httpPort, httpPort),
		},
		MonitorPorts: []monv1.Endpoint{
			{
				Port:           "http",
				RelabelConfigs: k8sutilv2.GetDefaultServiceMonitorRelabelConfig(),
			},
		},
	}
}

// NewStatefulSet returns a new statefulset container provider for the compactor.
// It includes the compactor container and the volume claims.
func NewSSContainerProvider(c *CompactorOptions) *k8sutilv2.Container {
	container := NewBaseContainerProvider(c)

	// Print warning if data directory is not specified.
	if c.DataDir == "" {
		log.Println("Warning: data directory is not specified for the statefulset.")
	}

	container.VolumeClaims = []k8sutilv2.VolumeClaim{
		{
			Name: dataVolumeName,
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
				StorageClassName: stringPtr("gp2-csi"),
			},
		},
	}

	return container
}

// NewPodSpecProvider returns a new pod spec provider for the compactor.
func NewPodSpecProvider(metaCfg k8sutilv2.MetaConfig, containers []k8sutilv2.ContainerProvider) *k8sutilv2.Pod {
	// Check if all the required labels are present in the meta config.
	for _, k := range []string{k8sutil.NameLabel, k8sutil.InstanceLabel} {
		if _, ok := metaCfg.Labels[k]; !ok {
			log.Printf("Warning: key %s not found in compactor meta labels", k)
		}
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     metaCfg.Labels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: metaCfg.Labels[k8sutil.InstanceLabel],
	}

	namespaces := []string{metaCfg.Namespace}

	return &k8sutilv2.Pod{
		TerminationGracePeriodSeconds: 120,
		Affinity:                      k8sutilv2.NewAntiAffinity(namespaces, labelSelectors),
		SecurityContext:               k8sutilv2.GetDefaultSecurityContext(),
		ContainerProviders:            containers,
	}
}

func NewStatefulSet(metaCfg k8sutilv2.MetaConfig, pod k8sutilv2.PodProvider) *k8sutilv2.StatefulSet {
	ss := k8sutilv2.StatefulSet{
		MetaConfig: metaCfg,
		Replicas:   1,
		Pod:        pod,
	}

	return &ss
}

func stringPtr(s string) *string { return &s }
