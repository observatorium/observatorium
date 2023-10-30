package k8sutil

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeploymentGenericConfig represents certain config fields
// that might be useful to add/override in a Deployment/StatefulSet. It contains
// fields of both DeploymentSpec and PodSpec.
type DeploymentGenericConfig struct {
	// Container fields
	Image                string
	ImageTag             string
	ImagePullPolicy      corev1.PullPolicy
	Name                 string
	Namespace            string
	CommonLabels         map[string]string
	Replicas             int32
	DeploymentStrategy   appsv1.DeploymentStrategy // Only applies to Deployment kind
	PodResources         corev1.ResourceRequirements
	Affinity             *corev1.Affinity
	SecurityContext      *corev1.PodSecurityContext
	EnableServiceMonitor bool
	Env                  []corev1.EnvVar
	LivenessProbe        *corev1.Probe
	ReadinessProbe       *corev1.Probe

	// Pod fields
	TerminationMessagePolicy      corev1.TerminationMessagePolicy
	TerminationGracePeriodSeconds int64

	// Container dependencies
	// ConfigMaps and Secrets are the ones required by the main container, others are directly defined in Sidecars
	ConfigMaps map[string]map[string]string // maps a configmap name to its data of type map[string]string
	Secrets    map[string]map[string][]byte // maps a secret name to its data of type map[string][]byte
	Sidecars   []ContainerProvider
}

func (d DeploymentGenericConfig) ToContainer() *Container {
	return &Container{
		Name:            d.Name,
		Image:           d.Image,
		ImageTag:        d.ImageTag,
		ImagePullPolicy: d.ImagePullPolicy,
		Env:             d.Env,
		Resources:       d.PodResources,
		LivenessProbe:   d.LivenessProbe,
		ReadinessProbe:  d.ReadinessProbe,
		ConfigMaps:      d.ConfigMaps,
		Secrets:         d.Secrets,
	}
}

type ObjectProcessor func(obj runtime.Object)
type DeploymentOption func(d *DeploymentGenericConfig)

// WithImage overrides the default image.
func WithImage(image, imageTag string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Image = image
		d.ImageTag = imageTag
	}
}

// WithImagePullPolicy overrides default image pull policy.
func WithImagePullPolicy(imagePullPolicy corev1.PullPolicy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.ImagePullPolicy = imagePullPolicy
	}
}

// WithName overrides the default name of all the individual objects.
func WithName(name string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Name = name
	}
}

// WithNamespace overrides the default namespace of all the individual objects.
func WithNamespace(namespace string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Namespace = namespace
	}
}

// WithCommonLabels overrides the default K8s metadata labels and selectors.
func WithCommonLabels(commonLabels map[string]string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.CommonLabels = commonLabels
	}
}

// WithReplicas overrides the default number of replicas to run.
func WithReplicas(replicas int32) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Replicas = replicas
	}
}

// WithDeploymentStrategy overrides the default deployment strategy of pods.
func WithDeploymentStrategy(ds appsv1.DeploymentStrategy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.DeploymentStrategy = ds
	}
}

// WithResources overrides the default Pod resource config.
func WithResources(resource corev1.ResourceRequirements) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.PodResources = resource
	}
}

// WithAffinity overrides the default Pod scheduling affinity rules.
func WithAffinity(affinity *corev1.Affinity) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Affinity = affinity
	}
}

// WithSecurityContext overrides the default Pod security context.
func WithSecurityContext(sc *corev1.PodSecurityContext) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.SecurityContext = sc
	}
}

// WithServiceMonitor enables generation of a ServiceMonitor to scrape the deployment.
func WithServiceMonitor() DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.EnableServiceMonitor = true
	}
}

// WithProbe overrides the default K8s liveness and readiness probes of main deployment container.
func WithProbes(livenessProbe, readinessProbe *corev1.Probe) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.LivenessProbe = livenessProbe
		d.ReadinessProbe = readinessProbe
	}
}

// WithTerminationMessagePolicy overrides the default termination message policy of main deployment container.
func WithTerminationMessagePolicy(terminationMessagePolicy corev1.TerminationMessagePolicy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.TerminationMessagePolicy = terminationMessagePolicy
	}
}

// WithTerminationGracePeriod overrides the default termination grace period of main deployment container.
func WithTerminationGracePeriod(duration int64) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.TerminationGracePeriodSeconds = duration
	}
}

// WithSideCars overrides the default pod sidecars.
func WithSidecars(sidecars ...ContainerProvider) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Sidecars = sidecars
	}
}
