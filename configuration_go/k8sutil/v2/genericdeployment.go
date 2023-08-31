package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// DeploymentGenericConfig represents certain config fields
// that might be useful to add/override in a Deployment/StatefulSet. It contains
// fields of both DeploymentSpec and PodSpec.
type DeploymentGenericConfig struct {
	Image                string
	ImageTag             string
	ImagePullPolicy      corev1.PullPolicy
	Name                 string
	Namespace            string
	CommonLabels         map[string]string
	Replicas             int32
	DeploymentStrategy   appsv1.DeploymentStrategy // Only applies to Deployment kind
	PodResources         corev1.ResourceRequirements
	Affinity             corev1.Affinity
	SecurityContext      corev1.PodSecurityContext
	EnableServiceMonitor bool
	Env                  []corev1.EnvVar

	LivenessProbe  corev1.Probe
	ReadinessProbe corev1.Probe

	TerminationMessagePolicy      corev1.TerminationMessagePolicy
	TerminationGracePeriodSeconds int64

	Sidecars   []ContainerProvider
	ConfigMaps map[string]map[string]string // Except the ones defined in sidecars
}

func (d DeploymentGenericConfig) ToContainer() *Container {
	return &Container{
		Name:            d.Name,
		Image:           d.Image,
		ImageTag:        d.ImageTag,
		ImagePullPolicy: d.ImagePullPolicy,
		Env:             d.Env,
		Resources:       d.PodResources,
		LivenessProbe:   &d.LivenessProbe,
		ReadinessProbe:  &d.ReadinessProbe,
		ConfigMaps:      d.ConfigMaps,
	}
}
