/*
This file contains the Kubernetes manifest providers.
These providers are simplified versions of some Kubernetes objects used to generate the manifests.
They implement the ManifestProvider interface, which allows generating the Kubernetes manifests.

Subcomponents provide methods required by higher-level components to generate their manifests.
For example, a pod implements the PodProvider interface, which provides methods to retrieve required volumes,
volume claims, config maps and secrets. These are, in turn, provided by the containers that are part of the pod.
This design groups the dependencies of each container in a single place for better code organization.
*/

package k8sutil

import (
	"fmt"
	"maps"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
		Name:        m.Name,
		Namespace:   m.Namespace,
		Labels:      m.Labels,
		Annotations: map[string]string{},
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

// ConfigMapsAndSecretsProvider is the interface that containers must implement to be added to a pod.
// It returns the list of config maps that the container requires.
type ConfigMapsAndSecretsProvider interface {
	GetConfigMaps() map[string]map[string]string
	GetSecrets() map[string]map[string][]byte
}

// ContainerProvider is the interface that containers must implement to be added to a pod.
type ContainerProvider interface {
	GetContainer() corev1.Container
	GetVolumes() []corev1.Volume

	ServiceProvider
	ServiceMonitorProvider
	PersistentVolumeClaimProvider
	ConfigMapsAndSecretsProvider
}

// Container represents a container in a pod.
// It encapsulates all the container's dependencies.
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

	// Dependencies
	Volumes      []corev1.Volume
	VolumeClaims []VolumeClaim
	ServicePorts []corev1.ServicePort
	MonitorPorts []monv1.Endpoint
	ConfigMaps   map[string]map[string]string
	Secrets      map[string]map[string][]byte
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

// GetConfigMaps returns the config maps that the container requires.
func (c *Container) GetConfigMaps() map[string]map[string]string {
	if c.ConfigMaps == nil {
		c.ConfigMaps = map[string]map[string]string{}
	}

	return c.ConfigMaps
}

// GetSecrets returns the secrets that the container requires.
func (c *Container) GetSecrets() map[string]map[string][]byte {
	if c.Secrets == nil {
		c.Secrets = map[string]map[string][]byte{}
	}

	return c.Secrets
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
	ConfigMapsAndSecretsProvider
}

// Pod represents a pod.
// It implements the PodProvider interface.
type Pod struct {
	TerminationGracePeriodSeconds *int64
	Affinity                      *corev1.Affinity
	SecurityContext               *corev1.PodSecurityContext
	ServiceAccountName            string

	ContainerProviders      []ContainerProvider
	InitContainersProviders []ContainerProvider
}

// MakePodSpec returns a Kubernetes PodSpec.
func (p *Pod) MakePodSpec() corev1.PodSpec {
	containers := []corev1.Container{}
	initContainers := []corev1.Container{}
	volumes := []corev1.Volume{}

	for _, cp := range p.ContainerProviders {
		containers = append(containers, cp.GetContainer())
		volumes = append(volumes, cp.GetVolumes()...)
	}

	for _, cp := range p.InitContainersProviders {
		initContainers = append(initContainers, cp.GetContainer())
		volumes = append(volumes, cp.GetVolumes()...)
	}

	return corev1.PodSpec{
		TerminationGracePeriodSeconds: p.TerminationGracePeriodSeconds,
		Affinity:                      p.Affinity,
		Containers:                    containers,
		InitContainers:                initContainers,
		ServiceAccountName:            p.ServiceAccountName,
		SecurityContext:               p.SecurityContext,
		NodeSelector: map[string]string{
			OsLabel: LinuxOs,
		},
		Volumes: volumes,
	}
}

// GetServicePorts returns the ports that the pod exposes.
func (p *Pod) GetServicePorts() []corev1.ServicePort {
	ret := []corev1.ServicePort{}

	for _, cp := range p.ContainerProviders {
		ret = append(ret, cp.GetServicePorts()...)
	}

	return ret
}

// GetServiceMonitorEndpoints returns the endpoints to be monitored by Prometheus.
func (p *Pod) GetServiceMonitorEndpoints() []monv1.Endpoint {
	ret := []monv1.Endpoint{}

	for _, cp := range p.ContainerProviders {
		ret = append(ret, cp.GetServiceMonitorEndpoints()...)
	}

	return ret
}

// GetVolumeClaims returns the volume claims that the pod requires.
func (p *Pod) GetVolumeClaims() []VolumeClaim {
	ret := []VolumeClaim{}

	for _, cp := range p.ContainerProviders {
		ret = append(ret, cp.GetVolumeClaims()...)
	}

	return ret
}

// GetConfigMaps returns the config maps that the pod requires.
func (p *Pod) GetConfigMaps() map[string]map[string]string {
	ret := map[string]map[string]string{}

	for _, cp := range p.ContainerProviders {
		for k, v := range cp.GetConfigMaps() {
			ret[k] = v
		}
	}

	return ret
}

// GetSecrets returns the secrets that the pod requires.
func (p *Pod) GetSecrets() map[string]map[string][]byte {
	ret := map[string]map[string][]byte{}

	for _, cp := range p.ContainerProviders {
		for k, v := range cp.GetSecrets() {
			ret[k] = v
		}
	}

	return ret
}

// Deployment represents a Kubernetes Deployment.
type Deployment struct {
	Replicas   int32
	MetaConfig MetaConfig
	Pod        PodProvider
}

// MakeManifest returns a Kubernetes Deployment.
func (d *Deployment) MakeManifest() runtime.Object {
	selectorMatcheLabels := maps.Clone(d.MetaConfig.Labels)
	delete(selectorMatcheLabels, VersionLabel)

	return &appsv1.Deployment{
		TypeMeta:   DeploymentMeta,
		ObjectMeta: d.MetaConfig.MakeMeta(),
		Spec: appsv1.DeploymentSpec{
			Replicas: &d.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorMatcheLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    maps.Clone(d.MetaConfig.Labels),
					Namespace: d.MetaConfig.Namespace,
				},
				Spec: d.Pod.MakePodSpec(),
			},
		},
	}
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
				Labels: maps.Clone(s.MetaConfig.Labels),
			},
			Spec: vc.Spec,
		})
	}

	selectorMatcheLabels := maps.Clone(s.MetaConfig.Labels)
	delete(selectorMatcheLabels, VersionLabel)

	return &appsv1.StatefulSet{
		TypeMeta:   StatefulSetMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Spec: appsv1.StatefulSetSpec{
			Replicas: &s.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorMatcheLabels,
			},
			ServiceName: s.MetaConfig.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    maps.Clone(s.MetaConfig.Labels),
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

type ServiceProviderFunc func() []corev1.ServicePort

func (f ServiceProviderFunc) GetServicePorts() []corev1.ServicePort {
	return f()
}

// Service represents a Kubernetes Service.
type Service struct {
	MetaConfig
	ServicePorts ServiceProvider
	ClusterIP    string
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
	delete(selector, VersionLabel)

	return &corev1.Service{
		TypeMeta:   ServiceMeta,
		ObjectMeta: s.MetaConfig.MakeMeta(),
		Spec: corev1.ServiceSpec{
			Ports:     s.ServicePorts.GetServicePorts(),
			Selector:  selector,
			ClusterIP: s.ClusterIP,
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
	delete(selector, VersionLabel)
	metaCfg := s.MetaConfig.MakeMeta()

	return &monv1.ServiceMonitor{
		TypeMeta:   ServiceMonitorMeta,
		ObjectMeta: metaCfg,
		Spec: monv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: selector,
			},
			Endpoints: s.ServiceMonitorEndpoints.GetServiceMonitorEndpoints(),
		},
	}
}

// ManifestProvider is the interface to be implemented to generate a Kubernetes manifest for a resource.
type ManifestProvider interface {
	MakeManifest() runtime.Object
}
