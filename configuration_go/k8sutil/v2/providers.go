package v2

import (
	"fmt"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MetaConfig represents the configuration required to create a Kubernetes ObjectMeta.
type MetaConfig struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

// MakeMeta returns a Kubernetes ObjectMeta.
func (m *MetaConfig) MakeMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      m.Name,
		Namespace: m.Namespace,
		Labels:    m.Labels,
	}
}

// Clone returns a copy of the MetaConfig.
func (m *MetaConfig) Clone() MetaConfig {
	return MetaConfig{
		Name:      m.Name,
		Namespace: m.Namespace,
		Labels:    maps.Clone(m.Labels),
	}
}

type ConfigMap struct {
	MetaConfig
	Data map[string]string
}

func NewConfigMap(metaConfig MetaConfig, data map[string]string) *ConfigMap {
	return &ConfigMap{
		MetaConfig: metaConfig,
		Data:       data,
	}
}

func (c *ConfigMap) MakeManifest() runtime.Object {
	return &corev1.ConfigMap{
		TypeMeta:   k8sutil.ConfigMapMeta,
		ObjectMeta: c.MetaConfig.MakeMeta(),
		Data:       c.Data,
	}
}

type Secret struct {
	MetaConfig
	Data map[string][]byte
}

func NewSecret(metaConfig MetaConfig, data map[string][]byte) *Secret {
	return &Secret{
		MetaConfig: metaConfig,
		Data:       data,
	}
}

func (s *Secret) MakeManifest() runtime.Object {
	return &corev1.Secret{
		TypeMeta:   k8sutil.SecretMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Data:       s.Data,
	}
}

// ContainerProvider is the interface that containers must implement to be added to a pod.
type ContainerProvider interface {
	GetContainer() corev1.Container
	GetVolumes() []corev1.Volume

	ServiceProvider
	ServiceMonitorProvider
	PersistentVolumeClaimProvider
	ManifestsProvider
}

// Container represents a container in a pod.
// It implements the ContainerProvider interface.
type Container struct {
	Name            string
	Image           string
	ImageTag        string
	ImagePullPolicy corev1.PullPolicy
	Resources       corev1.ResourceRequirements
	Env             []corev1.EnvVar
	LivenessProbe   *corev1.Probe
	ReadinessProbe  *corev1.Probe
	Args            []string
	Ports           []corev1.ContainerPort
	VolumeMounts    []corev1.VolumeMount

	Volumes      []corev1.Volume
	VolumeClaims []VolumeClaim
	ServicePorts []corev1.ServicePort
	MonitorPorts []monv1.Endpoint
	Manifests    map[string]ManifestProvider
}

// GetContainer returns a Kubernetes Container.
func (c *Container) GetContainer() corev1.Container {
	return corev1.Container{
		Name:                     c.Name,
		Image:                    fmt.Sprintf("%s:%s", c.Image, c.ImageTag),
		ImagePullPolicy:          c.ImagePullPolicy,
		Resources:                c.Resources,
		Env:                      c.Env,
		LivenessProbe:            c.LivenessProbe,
		ReadinessProbe:           c.ReadinessProbe,
		Args:                     c.Args,
		Ports:                    c.Ports,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		VolumeMounts:             c.VolumeMounts,
	}
}

// GetVolumes returns the volumes that the container requires.
func (c *Container) GetVolumes() []corev1.Volume {
	return c.Volumes
}

// GetServicePorts returns the ports that the container exposes.
func (c *Container) GetServicePorts() []corev1.ServicePort {
	return c.ServicePorts
}

// GetServiceMonitorEndpoints returns the endpoints to be monitored by Prometheus.
func (c *Container) GetServiceMonitorEndpoints() []monv1.Endpoint {
	return c.MonitorPorts
}

// GetVolumeClaims returns the volume claims that the container requires.
func (c *Container) GetVolumeClaims() []VolumeClaim {
	return c.VolumeClaims
}

// AddConfigMap adds a config map to the container.
func (c *Container) AddManifest(filename string, mf ManifestProvider) {
	if c.Manifests == nil {
		c.Manifests = map[string]ManifestProvider{}
	}

	c.Manifests[filename] = mf
}

func (c *Container) MakeManifests() k8sutil.ObjectMap {
	ret := k8sutil.ObjectMap{}

	for filename, mf := range c.Manifests {
		ret[filename] = mf.MakeManifest()
	}

	return ret
}

// VolumeClaim represents a volume claim.
type VolumeClaim struct {
	Name string
	Spec corev1.PersistentVolumeClaimSpec
}

// PersistentVolumeClaimProvider is the interface to be implemented by pods that require volume claims.
type PersistentVolumeClaimProvider interface {
	GetVolumeClaims() []VolumeClaim
}

// PodProvider is the interface to be implemented by pods to be used in a StatefulSet or Deployment.
type PodProvider interface {
	MakePodSpec() corev1.PodSpec

	ServiceProvider
	ServiceMonitorProvider
	PersistentVolumeClaimProvider
}

// Pod represents a pod.
// It implements the PodProvider interface.
type Pod struct {
	TerminationGracePeriodSeconds int64
	Affinity                      *corev1.Affinity
	SecurityContext               corev1.PodSecurityContext
	ServiceAccountName            string

	ContainerProviders []ContainerProvider
}

// MakePodSpec returns a Kubernetes PodSpec.
func (s *Pod) MakePodSpec() corev1.PodSpec {
	containers := []corev1.Container{}
	volumes := []corev1.Volume{}

	for _, cp := range s.ContainerProviders {
		containers = append(containers, cp.GetContainer())
		volumes = append(volumes, cp.GetVolumes()...)
	}

	return corev1.PodSpec{
		TerminationGracePeriodSeconds: int64Ptr(s.TerminationGracePeriodSeconds),
		Affinity:                      s.Affinity,
		Containers:                    containers,
		ServiceAccountName:            s.ServiceAccountName,
		SecurityContext:               &s.SecurityContext,
		NodeSelector: map[string]string{
			k8sutil.OsLabel: k8sutil.LinuxOs,
		},
		Volumes: volumes,
	}
}

// GetServicePorts returns the ports that the pod exposes.
func (s *Pod) GetServicePorts() []corev1.ServicePort {
	ret := []corev1.ServicePort{}

	for _, cp := range s.ContainerProviders {
		ret = append(ret, cp.GetServicePorts()...)
	}

	return ret
}

// GetServiceMonitorEndpoints returns the endpoints to be monitored by Prometheus.
func (s *Pod) GetServiceMonitorEndpoints() []monv1.Endpoint {
	ret := []monv1.Endpoint{}

	for _, cp := range s.ContainerProviders {
		ret = append(ret, cp.GetServiceMonitorEndpoints()...)
	}

	return ret
}

// GetVolumeClaims returns the volume claims that the pod requires.
func (s *Pod) GetVolumeClaims() []VolumeClaim {
	ret := []VolumeClaim{}

	for _, cp := range s.ContainerProviders {
		ret = append(ret, cp.GetVolumeClaims()...)
	}

	return ret
}

// StatefulSet represents a Kubernetes StatefulSet.
type StatefulSet struct {
	Replicas   int32
	MetaConfig MetaConfig
	Pod        PodProvider
}

// MakeManifest returns a Kubernetes StatefulSet.
func (s *StatefulSet) MakeManifest() runtime.Object {
	vcs := []corev1.PersistentVolumeClaim{}
	for _, vc := range s.Pod.GetVolumeClaims() {
		vcs = append(vcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:   vc.Name,
				Labels: s.MetaConfig.Labels,
			},
			Spec: vc.Spec,
		})
	}

	selectorMatcheLabels := maps.Clone(s.MetaConfig.Labels)
	delete(selectorMatcheLabels, k8sutil.VersionLabel)

	return &appsv1.StatefulSet{
		TypeMeta:   k8sutil.StatefulSetMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(s.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorMatcheLabels,
			},
			ServiceName: s.MetaConfig.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    s.MetaConfig.Labels,
					Namespace: s.MetaConfig.Namespace,
				},
				Spec: s.Pod.MakePodSpec(),
			},
			VolumeClaimTemplates: vcs,
		},
	}
}

// ServiceProvider is the interface to be implemented by pods that require a service.
type ServiceProvider interface {
	GetServicePorts() []corev1.ServicePort
}

// Service represents a Kubernetes Service.
type Service struct {
	MetaConfig
	ServicePorts ServiceProvider
}

// NewService returns a new Service.
func NewService(metaConfig MetaConfig, servicePorts ServiceProvider) *Service {
	return &Service{
		MetaConfig:   metaConfig,
		ServicePorts: servicePorts,
	}
}

// MakeManifest returns a Kubernetes Service.
func (s *Service) MakeManifest() runtime.Object {
	selector := maps.Clone(s.MetaConfig.Labels)
	delete(selector, k8sutil.VersionLabel)

	return &corev1.Service{
		TypeMeta:   k8sutil.ServiceMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Spec: corev1.ServiceSpec{
			Ports:    s.ServicePorts.GetServicePorts(),
			Selector: selector,
		},
	}
}

// ServiceMonitorProvider is the interface to be implemented by pods that require a service monitor.
type ServiceMonitorProvider interface {
	GetServiceMonitorEndpoints() []monv1.Endpoint
}

// ServiceMonitor represents a Kubernetes ServiceMonitor.
type ServiceMonitor struct {
	MetaConfig
	ServiceMonitorEndpoints ServiceMonitorProvider
}

// NewServiceMonitor returns a new ServiceMonitor.
func NewServiceMonitor(metaConfig MetaConfig, serviceMonitorEndpoints ServiceMonitorProvider) *ServiceMonitor {
	return &ServiceMonitor{
		MetaConfig:              metaConfig,
		ServiceMonitorEndpoints: serviceMonitorEndpoints,
	}
}

// MakeManifest returns a Kubernetes ServiceMonitor.
func (s *ServiceMonitor) MakeManifest() runtime.Object {
	selector := maps.Clone(s.MetaConfig.Labels)
	delete(selector, k8sutil.VersionLabel)

	return &monv1.ServiceMonitor{
		TypeMeta:   k8sutil.ServiceMonitorMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Spec: monv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: selector,
			},
			Endpoints: s.ServiceMonitorEndpoints.GetServiceMonitorEndpoints(),
		},
	}
}

// ServiceAccount represents a Kubernetes ServiceAccount.
type ServiceAccount struct {
	MetaConfig
	Name string
}

// NewServiceAccount returns a new ServiceAccount.
func NewServiceAccount(metaConfig MetaConfig, name string) *ServiceAccount {
	return &ServiceAccount{
		MetaConfig: metaConfig,
		Name:       name,
	}
}

// MakeManifest returns a Kubernetes ServiceAccount.
func (s *ServiceAccount) MakeManifest() runtime.Object {
	return &corev1.ServiceAccount{
		TypeMeta:   k8sutil.ServiceAccountMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
	}
}

// ManifestProvider is the interface to be implemented to generate a Kubernetes manifest for a resource.
type ManifestProvider interface {
	MakeManifest() runtime.Object
}

// Manifests represents a collection of manifests.
type Manifests struct {
	manifests map[string]ManifestProvider
}

type ManifestsProvider interface {
	MakeManifests() k8sutil.ObjectMap
}

// NewManifests returns a new Manifests.
func NewManifests() *Manifests {
	return &Manifests{
		manifests: map[string]ManifestProvider{},
	}
}

// Add adds a manifest to the collection.
func (m *Manifests) Add(filename string, mf ManifestProvider) {
	m.manifests[filename] = mf
}

// Make returns an ObjectMap that maps filenames to Kubernetes manifests.
func (m *Manifests) Make() k8sutil.ObjectMap {
	ret := k8sutil.ObjectMap{}

	for filename, mf := range m.manifests {
		ret[filename] = mf.MakeManifest()
	}

	return ret
}

func int64Ptr(i int64) *int64 { return &i }
func int32Ptr(i int32) *int32 { return &i }