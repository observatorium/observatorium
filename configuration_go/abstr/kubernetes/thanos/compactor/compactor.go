package compactor

import (
	"fmt"
	"maps"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	k8sutilv2 "github.com/observatorium/observatorium/configuration_go/k8sutil/v2"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	dataVolumeName      string = "data"
	defaultHTTPPort     int    = 10902
	defaultNamespace    string = "observatorium"
	servicePortName     string = "http"
	thanosContainerName string = "thanos"
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

// K8sConfig represents the Kubernetes configuration for the compactor.
type K8sConfig struct {
	DataVolumeClaim *corev1.PersistentVolumeClaim
	k8sutilv2.CommonConfig
}

// DefaultK8sConfig returns the default configuration for the compactor.
// It can be used as a base for further customization.
func DefaultK8sConfig() *K8sConfig {
	labels := map[string]string{
		k8sutil.ComponentLabel: "database-compactor",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.NameLabel:      "thanos-compact",
		k8sutil.PartOfLabel:    "observatorium",
	}

	ret := &K8sConfig{
		DataVolumeClaim: &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dataVolumeName,
				Namespace: defaultNamespace,
				Labels:    labels,
			},
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
		CommonConfig: k8sutilv2.CommonConfig{
			Image:                         "quay.io/thanos/thanos",
			ImageTag:                      "v0.31.0",
			Name:                          "thanos-compact",
			Namespace:                     defaultNamespace,
			Labels:                        labels,
			Replicas:                      1,
			TerminationGracePeriodSeconds: 120,
			ImagePullPolicy:               corev1.PullIfNotPresent,
			SecurityContext:               k8sutilv2.GetDefaultSecurityContext(),
			ServiceMonitor:                k8sutilv2.ServiceMonitorConfig{Enabled: true, Namespace: defaultNamespace, Labels: maps.Clone(labels)},
			LivenessProbe: k8sutilv2.ProbeConfig{
				FailureThreshold: 4,
				PeriodSeconds:    30,
			},
			ReadinessProbe: k8sutilv2.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			},
			Resources: k8sutilv2.NewResourcesRequirements("2", "3", "2000Mi", "3000Mi"),
		},
	}

	ret.Affinity = k8sutilv2.NewAntiAffinity(
		[]string{ret.Namespace},
		map[string]string{
			k8sutil.NameLabel:     ret.Labels[k8sutil.NameLabel],
			k8sutil.InstanceLabel: ret.Labels[k8sutil.InstanceLabel],
		},
	)

	return ret
}

// manifestKeys represents the the filenames used to generate the kubernetes manifests.
type manifestKeys struct {
	Deployment     string
	Service        string
	ServiceAccount string
	ServiceMonitor string
}

// Compactor represents the compactor. It contains both the compactor options and the Kubernetes configuration.
type compactor struct {
	ManifestKeys manifestKeys

	kubeCfg   *K8sConfig
	options   *CompactorOptions
	manifests k8sutil.ObjectMap
}

// NewCompactor returns a new compactor.
// It is used to create the kubernetes manifests for deploying the compactor.
func NewCompactor(options *CompactorOptions, k8sCfg *K8sConfig) *compactor {
	if k8sCfg == nil {
		k8sCfg = DefaultK8sConfig()
	}

	ret := &compactor{
		ManifestKeys: manifestKeys{
			Deployment:     "compactor-statefulset",
			Service:        "compactor-service",
			ServiceAccount: "compactor-serviceAccount",
			ServiceMonitor: "compactor-service-monitor",
		},
		options:   options,
		kubeCfg:   k8sCfg,
		manifests: k8sutil.ObjectMap{},
	}

	return ret
}

// Manifests returns the kubernetes manifests for deploying the compactor.
func (c *compactor) Manifests() k8sutil.ObjectMap {
	c.makeStatefulSet()
	c.makeServiceAccount()
	c.makeService()

	if c.kubeCfg.ServiceMonitor.Enabled {
		c.makeServiceMonitor()
	}
	return c.manifests
}

func (c *compactor) makeServiceAccount() {
	queryServiceAccount := corev1.ServiceAccount{
		TypeMeta: k8sutil.ServiceAccountMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.kubeCfg.Name,
			Labels:    c.kubeCfg.Labels,
			Namespace: c.kubeCfg.Namespace,
		},
	}

	c.manifests[c.ManifestKeys.ServiceAccount] = &queryServiceAccount
}

func (c *compactor) makeServiceMonitor() {
	endpoints := []monv1.Endpoint{
		{
			Port:           servicePortName,
			RelabelConfigs: k8sutilv2.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	endpoints = append(endpoints, c.kubeCfg.ServiceMonitorEndpoints...)

	serviceMonitor := monv1.ServiceMonitor{
		TypeMeta: k8sutil.ServiceMonitorMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.kubeCfg.Name,
			Namespace: c.kubeCfg.Namespace,
			Labels:    c.kubeCfg.Labels,
		},

		Spec: monv1.ServiceMonitorSpec{
			Endpoints: endpoints,
			Selector: metav1.LabelSelector{
				MatchLabels: c.kubeCfg.Labels,
			},
		},
	}

	c.manifests[c.ManifestKeys.ServiceMonitor] = &serviceMonitor
}

func (c *compactor) makeService() {
	httpPort := c.getHTTPPort()
	ports := []corev1.ServicePort{
		{
			Name:       servicePortName,
			Port:       int32(httpPort),
			TargetPort: intstr.FromInt(httpPort),
			Protocol:   corev1.ProtocolTCP,
		},
	}
	ports = append(ports, c.kubeCfg.ServicePorts...)

	labelsWithVersion := maps.Clone(c.kubeCfg.Labels)
	labelsWithVersion[k8sutil.VersionLabel] = c.kubeCfg.ImageTag

	service := corev1.Service{
		TypeMeta: k8sutil.ServiceMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.kubeCfg.Name,
			Labels:    labelsWithVersion,
			Namespace: c.kubeCfg.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: c.kubeCfg.Labels,
		},
	}

	c.manifests[c.ManifestKeys.Service] = &service
}

func (c *compactor) makeStatefulSet() {
	podTemplateSpec := c.makePodTemplateSpec()

	labelsWithVersion := maps.Clone(c.kubeCfg.Labels)
	labelsWithVersion[k8sutil.VersionLabel] = c.kubeCfg.ImageTag

	commonObjectMeta := metav1.ObjectMeta{
		Name:      c.kubeCfg.Name,
		Labels:    labelsWithVersion,
		Namespace: c.kubeCfg.Namespace,
	}

	var volumeClaims []corev1.PersistentVolumeClaim
	if c.kubeCfg.DataVolumeClaim != nil {
		volumeClaims = append(volumeClaims, *c.kubeCfg.DataVolumeClaim)
	}

	statefulSet := appsv1.StatefulSet{
		TypeMeta:   k8sutil.StatefulSetMeta,
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(c.kubeCfg.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: c.kubeCfg.Labels,
			},
			ServiceName:          c.kubeCfg.Name,
			VolumeClaimTemplates: volumeClaims,
			Template:             podTemplateSpec,
		},
	}

	c.manifests[c.ManifestKeys.Deployment] = &statefulSet
}

func (c *compactor) makePodTemplateSpec() corev1.PodTemplateSpec {
	containers := []corev1.Container{c.makeCompactorContainer()}
	containers = append(containers, c.kubeCfg.SideCars...)

	labelsWithVersion := maps.Clone(c.kubeCfg.Labels)
	labelsWithVersion[k8sutil.VersionLabel] = c.kubeCfg.ImageTag

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labelsWithVersion,
			Namespace: c.kubeCfg.Namespace,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: int64Ptr(c.kubeCfg.TerminationGracePeriodSeconds),
			Affinity:                      c.kubeCfg.Affinity,
			Containers:                    containers,
			ServiceAccountName:            c.kubeCfg.Name,
			SecurityContext:               &c.kubeCfg.SecurityContext,
			NodeSelector: map[string]string{
				k8sutil.OsLabel: k8sutil.LinuxOs,
			},
			Volumes: c.kubeCfg.PodVolumes,
		},
	}
}

func (c *compactor) makeCompactorContainer() corev1.Container {
	httpPort := c.getHTTPPort()

	args := []string{"compact"}
	args = append(args, cmdopt.GetOpts(c.options)...)

	ret := corev1.Container{
		Name:                     thanosContainerName,
		Image:                    fmt.Sprintf("%s:%s", c.kubeCfg.Image, c.kubeCfg.ImageTag),
		ImagePullPolicy:          c.kubeCfg.ImagePullPolicy,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		Args:                     args,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: int32(httpPort),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe:  k8sutilv2.NewProbe("/-/healthy", httpPort, c.kubeCfg.LivenessProbe),
		ReadinessProbe: k8sutilv2.NewProbe("/-/ready", httpPort, c.kubeCfg.ReadinessProbe),
		Resources:      c.kubeCfg.Resources,
		Env:            c.kubeCfg.Env,
	}

	// If a data directory is specified, mount the volume.
	if c.options.DataDir != "" {
		ret.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      c.kubeCfg.DataVolumeClaim.Name,
				MountPath: c.options.DataDir,
				ReadOnly:  false,
			},
		}
	}

	return ret
}

func (c *compactor) getHTTPPort() int {
	if c.options.HttpAddress.Port != 0 {
		return c.options.HttpAddress.Port
	}
	return defaultHTTPPort
}

func int32Ptr(i int32) *int32    { return &i }
func int64Ptr(i int64) *int64    { return &i }
func stringPtr(s string) *string { return &s }
