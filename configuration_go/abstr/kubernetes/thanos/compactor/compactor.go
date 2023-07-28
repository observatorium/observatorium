package compactor

import (
	"fmt"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/clioption"
	k8sutil "github.com/observatorium/observatorium/configuration_go/k8sutil"
	k8sutilv2 "github.com/observatorium/observatorium/configuration_go/k8sutil/v2"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DeploymentKey     string = "compactor-deployment"
	ServiceKey        string = "compactor-service"
	ServiceMonitorKey string = "compactor-service-monitor"
)

// CompactorOptions represents the options/flags for the compactor.
// See https://thanos.io/tip/components/compact.md/#flags for details.
type CompactorOptions struct {
	BlockFilesConcurrency              int                    `opt:"block-files-concurrency"`
	BlockMetaFetchConcurrency          int                    `opt:"block-meta-fetch-concurrency"`
	BlockViewerGlobalSyncBlockInterval time.Duration          `opt:"block-viewer.global.sync-block-interval"`
	BlockViewerGlobalSyncBlockTimeout  time.Duration          `opt:"block-viewer.global.sync-block-timeout"`
	BucketWebLabel                     string                 `opt:"bucket-web-label"`
	CompactBlocksFetchConcurrency      int                    `opt:"compact.blocks-fetch-concurrency"`
	CompactCleanupInterval             time.Duration          `opt:"compact.cleanup-interval"`
	CompactConcurrency                 int                    `opt:"compact.concurrency"`
	CompactProgressInterval            time.Duration          `opt:"compact.progress-interval"`
	ConsistencyDelay                   time.Duration          `opt:"consistency-delay"`
	DataDir                            string                 `opt:"data-dir"`
	DeduplicationFunc                  string                 `opt:"deduplication.func"`
	DeduplicationReplicaLabel          string                 `opt:"deduplication.replica-label"`
	DeleteDelay                        time.Duration          `opt:"delete-delay"`
	DownsampleConcurrency              int                    `opt:"downsample.concurrency"`
	DownsamplingDisable                bool                   `opt:"downsampling.disable"`
	HashFunc                           string                 `opt:"hash-func"`
	HttpAddress                        net.TCPAddr            `opt:"http-address"`
	HttpGracePeriod                    time.Duration          `opt:"http-grace-period"`
	HttpConfig                         string                 `opt:"http.config"`
	LogFormat                          string                 `opt:"log.format"`
	LogLevel                           string                 `opt:"log.level"`
	MaxTime                            string                 `opt:"max-time"`
	MinTime                            string                 `opt:"min-time"`
	ObjstoreConfig                     string                 `opt:"objstore.config"`
	ObjstoreConfigFile                 string                 `opt:"objstore.config-file"`
	RetentionResolution1h              time.Duration          `opt:"retention.resolution-1h"`
	RetentionResolution5m              time.Duration          `opt:"retention.resolution-5m"`
	RetentionResolutionRaw             time.Duration          `opt:"retention.resolution-raw"`
	SelectorRelabelConfig              []*monv1.RelabelConfig `opt:"selector.relabel-config"`
	SelectorRelabelConfigFile          []*monv1.RelabelConfig `opt:"selector.relabel-config-file"`
	TracingConfig                      string                 `opt:"tracing.config"`
	TracingConfigFile                  string                 `opt:"tracing.config-file"`
	Version                            bool                   `opt:"version,noval"`
	Wait                               bool                   `opt:"wait,noval"`
	WaitInterval                       time.Duration          `opt:"wait-interval"`
	WebDisable                         bool                   `opt:"web.disable"`
	WebDisableCors                     bool                   `opt:"web.disable-cors"`
	WebExternalPrefix                  string                 `opt:"web.external-prefix"`
	WebPrefixHeader                    string                 `opt:"web.prefix-header"`
	WebRoutePrefix                     string                 `opt:"web.route-prefix"`
}

type Compactor struct {
	KubeCfg *k8sutilv2.CommonConfig
	// Deployment *appsv1.Deployment
	Options   *CompactorOptions
	manifests k8sutil.ObjectMap
	labels    map[string]string
}

func NewCompactor(options *CompactorOptions, kCfg *k8sutilv2.CommonConfig) *Compactor {

	ret := &Compactor{
		Options:   options,
		KubeCfg:   kCfg,
		manifests: k8sutil.ObjectMap{},
		labels: map[string]string{
			k8sutil.ComponentLabel: "database-compactor",
			k8sutil.InstanceLabel:  "observatorium",
			k8sutil.NameLabel:      "thanos-compact",
			k8sutil.PartOfLabel:    "observatorium",
		},
	}

	ret.makeDeployment()

	return ret
}

// K8sConfig overrides the default K8s options for Thanos Query.
func (c *Compactor) K8sConfig(opts ...k8sutilv2.DeploymentModifier) *Compactor {
	for _, o := range opts {
		o(c.getManifest(DeploymentKey).(*appsv1.Deployment))
	}

	return c
}

func (c *Compactor) AddServiceMonitor() {
	sm := k8sutilv2.NewServiceMonitor(c.KubeCfg, c.KubeCfg.Labels)
	c.manifests[ServiceMonitorKey] = sm
}

func (c *Compactor) Manifests() k8sutil.ObjectMap {
	return c.manifests
}

func (c *Compactor) getManifest(key string) runtime.Object {
	value, ok := c.manifests[key]
	if !ok {
		panic(fmt.Errorf("no manifest found for key %s", key))
	}

	return value
}

func (c *Compactor) makeDeployment() {
	// dep := k8sutilv2.DeploymentGenericConfig{}

	httpPort := 10902
	if c.Options.HttpAddress.Port != 0 {
		httpPort = c.Options.HttpAddress.Port
	}

	compactorContainer := corev1.Container{
		Name:  "thanos",
		Image: "quay.io/thanos/thanos:v0.31.0",
		Args:  []string{"compact"},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: int32(httpPort),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe:  k8sutilv2.NewProb("/-/healthy", httpPort, 4, 30),
		ReadinessProbe: k8sutilv2.NewProb("/-/ready", httpPort, 20, 5),
	}

	compactorContainer.Args = append(compactorContainer.Args, cmdopt.GetOpts(c.Options)...)

	// Attach any configured sidecars.
	containers := []corev1.Container{compactorContainer}

	deployment := appsv1.Deployment{
		TypeMeta: k8sutil.DeploymentMeta,
		// ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: c.labels,
			},
			// Strategy: q.DeploymentStrategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    c.KubeCfg.Labels,
					Namespace: c.KubeCfg.Namespace,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: int64Ptr(120),
					// Affinity:           &q.Affinity,
					Containers: containers,
					// ServiceAccountName: queryServiceAccount.Name,
					SecurityContext: k8sutilv2.NewSecurityContext(),
				},
			},
		},
	}

	c.manifests[DeploymentKey] = &deployment
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
