package workload

import (
	"maps"
	"unicode/utf8"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeploymentWorkload represents a generic deployment workload with most commonly used options.
type DeploymentWorkload struct {
	DeploymentStrategy appsv1.DeploymentStrategy
	Replicas           int32

	PodConfig
}

// Objects returns the list of runtime objects for the given workload.
func (d DeploymentWorkload) Objects(container *Container) []runtime.Object {
	pod := d.Pod(container)
	ret := d.generateCommonObjects(pod)
	ret = append(ret, d.deployment(pod))

	return ret
}

func (d DeploymentWorkload) deployment(pod *Pod) runtime.Object {
	dep := &Deployment{
		MetaConfig: *d.ObjectMeta(),
		Replicas:   int32(d.Replicas),
		Strategy:   d.DeploymentStrategy,
		Pod:        pod,
	}

	return dep.Object()
}

// StatefulSetWorkload represents a generic statefulset workload with most commonly used options.
type StatefulSetWorkload struct {
	Replicas   int32
	VolumeType string
	VolumeSize string

	PodConfig
}

// Objects returns the list of runtime objects for the given workload.
func (s StatefulSetWorkload) Objects(container *Container) []runtime.Object {
	pod := s.Pod(container)
	ret := s.generateCommonObjects(pod)
	ret = append(ret, s.statefulSet(pod))

	return ret
}

func (s StatefulSetWorkload) statefulSet(pod *Pod) runtime.Object {
	statefulset := &StatefulSet{
		MetaConfig: *s.ObjectMeta(),
		Replicas:   int32(s.Replicas),
		Pod:        pod,
	}

	return statefulset.Object()
}

// PodConfig represents a generic pod configuration with most commonly used options.
type PodConfig struct {
	// Container fields
	ContainerResources corev1.ResourceRequirements
	Env                []corev1.EnvVar
	Image              string
	ImageTag           string
	ImagePullPolicy    corev1.PullPolicy
	LivenessProbe      *corev1.Probe
	ReadinessProbe     *corev1.Probe

	// Pod fields
	Affinity                      *corev1.Affinity
	SecurityContext               *corev1.PodSecurityContext
	TerminationGracePeriodSeconds int64

	// Workload fields
	CommonLabels map[string]string
	Name         string
	Namespace    string

	EnableServiceMonitor bool

	// Container dependencies
	// ConfigMaps and Secrets are the ones required by the main container, others are directly defined in Sidecars
	ConfigMaps     map[string]map[string]string // maps a configmap name to its data of type map[string]string
	Secrets        map[string]map[string][]byte // maps a secret name to its data of type map[string][]byte
	Sidecars       []ContainerProvider
	InitContainers []ContainerProvider
}

// ToContainer returns the main Container object of the pod with the given pod configuration.
func (d PodConfig) ToContainer() *Container {
	return &Container{
		Name:            d.Name,
		Image:           d.Image,
		ImageTag:        d.ImageTag,
		ImagePullPolicy: d.ImagePullPolicy,
		Env:             d.Env,
		Resources:       d.ContainerResources,
		LivenessProbe:   d.LivenessProbe,
		ReadinessProbe:  d.ReadinessProbe,
		ConfigMaps:      d.ConfigMaps,
		Secrets:         d.Secrets,
	}
}

// ObjectMeta returns the ObjectMeta object of the pod with the given pod configuration.
func (d PodConfig) ObjectMeta() *MetaConfig {
	labels := maps.Clone(d.CommonLabels)
	if d.ImageTag != "" {
		labels[VersionLabel] = d.ImageTag
	}

	return &MetaConfig{
		Name:      d.Name,
		Namespace: d.Namespace,
		Labels:    labels,
	}
}

// Pod returns a Pod object with the given container and sidecars.
func (d PodConfig) Pod(container *Container) *Pod {
	return &Pod{
		TerminationGracePeriodSeconds: &d.TerminationGracePeriodSeconds,
		Affinity:                      d.Affinity,
		SecurityContext:               d.SecurityContext,
		ServiceAccountName:            d.Name,
		ContainerProviders:            append([]ContainerProvider{container}, d.Sidecars...),
		InitContainersProviders:       d.InitContainers,
	}
}

// Service returns a Service object for the given pod.
func (d PodConfig) Service(pod *Pod) runtime.Object {
	service := &Service{
		MetaConfig:   *d.ObjectMeta(),
		ServicePorts: pod,
	}

	return service.Object()
}

// ServiceMonitor returns a ServiceMonitor object for the given pod.
func (d PodConfig) ServiceMonitor(pod *Pod) runtime.Object {
	serviceMonitor := &ServiceMonitor{
		MetaConfig:              *d.ObjectMeta(),
		ServiceMonitorEndpoints: pod,
	}

	return serviceMonitor.Object()
}

// ServiceAccount returns a ServiceAccount object.
func (d PodConfig) ServiceAccount() runtime.Object {
	metaCfg := d.ObjectMeta().MakeMeta()
	delete(metaCfg.Labels, VersionLabel)
	return &corev1.ServiceAccount{
		TypeMeta:   ServiceAccountMeta,
		ObjectMeta: metaCfg,
	}
}

// ConfigMapsAndSecrets returns the list of ConfigMap and Secret objects for the given pod.
func (d PodConfig) ConfigMapsAndSecrets(pod *Pod) []runtime.Object {
	ret := []runtime.Object{}
	for name, data := range pod.GetConfigMaps() {
		metaCfg := d.ObjectMeta().MakeMeta()
		delete(metaCfg.Labels, VersionLabel)
		cm := &corev1.ConfigMap{
			TypeMeta:   ConfigMapMeta,
			ObjectMeta: metaCfg,
			Data:       data,
		}
		cm.Name = name
		ret = append(ret, cm)
	}

	for name, data := range pod.GetSecrets() {
		metaCfg := d.ObjectMeta().MakeMeta()
		delete(metaCfg.Labels, VersionLabel)
		secret := &corev1.Secret{
			TypeMeta:   SecretMeta,
			ObjectMeta: metaCfg,
		}
		secret.Name = name

		// check if data is a string and store it as a stringData if possible for better readability
		stringData := map[string]string{}
		isStringData := true
		for k, v := range data {
			if utf8.Valid(v) {
				stringData[k] = string(v)
			} else {
				isStringData = false
				break
			}
		}

		if isStringData {
			secret.StringData = stringData
		} else {
			secret.Data = data
		}
		ret = append(ret, secret)
	}

	return ret
}

func (d PodConfig) generateCommonObjects(pod *Pod) []runtime.Object {
	ret := []runtime.Object{
		d.ServiceAccount(),
	}

	if len(pod.GetServicePorts()) > 0 {
		ret = append(ret, d.Service(pod))
	}

	if d.EnableServiceMonitor && len(pod.GetServiceMonitorEndpoints()) > 0 {
		ret = append(ret, d.ServiceMonitor(pod))
	}

	ret = append(ret, d.ConfigMapsAndSecrets(pod)...)

	return ret
}
