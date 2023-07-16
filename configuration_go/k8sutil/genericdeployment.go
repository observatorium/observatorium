package k8sutil

import (
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// SidecarConfig represents the configuration required to add extra containers
// to a particular Deployment/StatefulSet.
type SidecarConfig struct {
	Sidecars                      []corev1.Container
	AdditionalPodVolumes          []corev1.Volume
	AdditionalServicePorts        []corev1.ServicePort
	AdditionalServiceMonitorPorts []monv1.Endpoint
}

// GenericDeploymentConfig represents certain config fields
// that might be useful to add/override in a Deployment. It contains
// fields of both DeploymentSpec and PodSpec.
// It also has method defined for overriding any default values.
type GenericDeploymentConfig struct {
	Image                string
	ImageTag             string
	ImagePullPolicy      corev1.PullPolicy
	Name                 string
	Namespace            string
	CommonLabels         map[string]string
	Replicas             int32
	DeploymentStrategy   appsv1.DeploymentStrategy
	PodResources         corev1.ResourceRequirements
	Affinity             corev1.Affinity
	SecurityContext      corev1.PodSecurityContext
	EnableServiceMonitor bool

	LivenessProbe  corev1.Probe
	ReadinessProbe corev1.Probe

	TerminationMessagePolicy      corev1.TerminationMessagePolicy
	TerminationGracePeriodSeconds int64

	Sidecars SidecarConfig
}

type DeploymentOption func(d *GenericDeploymentConfig)

// WithImage overrides the default image.
func WithImage(image, imageTag string) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Image = image
		d.ImageTag = imageTag
	}
}

// WithImagePullPolicy overrides default image pull policy.
func WithImagePullPolicy(imagePullPolicy corev1.PullPolicy) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.ImagePullPolicy = imagePullPolicy
	}
}

// WithName overrides the default name of all the individual objects.
func WithName(name string) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Name = name
	}
}

// WithNamespace overrides the default namespace of all the individual objects.
func WithNamespace(namespace string) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Namespace = namespace
	}
}

// WithCommonLabels overrides the default K8s metadata labels and selectors.
func WithCommonLabels(commonLabels map[string]string) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.CommonLabels = commonLabels
	}
}

// WithReplicas overrides the default number of replicas to run.
func WithReplicas(replicas int32) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Replicas = replicas
	}
}

// WithDeploymentStrategy overrides the default deployment strategy of pods.
func WithDeploymentStrategy(ds appsv1.DeploymentStrategy) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.DeploymentStrategy = ds
	}
}

// WithResources overrides the default Pod resource config.
func WithResources(resource corev1.ResourceRequirements) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.PodResources = resource
	}
}

// WithAffinity overrides the default Pod scheduling affinity rules.
func WithAffinity(affinity corev1.Affinity) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Affinity = affinity
	}
}

// WithSecurityContext overrides the default Pod security context.
func WithSecurityContext(sc corev1.PodSecurityContext) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.SecurityContext = sc
	}
}

// WithServiceMonitor enables generation of a ServiceMonitor to scrape the deployment.
func WithServiceMonitor() DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.EnableServiceMonitor = true
	}
}

// WithProbe overrides the default K8s liveness and readiness probes of main deployment container.
func WithProbes(livenessProbe, readinessProbe corev1.Probe) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.LivenessProbe = livenessProbe
		d.ReadinessProbe = readinessProbe
	}
}

// WithTerminationMessagePolicy overrides the default termination message policy of main deployment container.
func WithTerminationMessagePolicy(terminationMessagePolicy corev1.TerminationMessagePolicy) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.TerminationMessagePolicy = terminationMessagePolicy
	}
}

// WithTerminationGracePeriod overrides the default termination grace period of main deployment container.
func WithTerminationGracePeriod(duration int64) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.TerminationGracePeriodSeconds = duration
	}
}

// WithSidecars allows overriding default K8s Deployment and adding additional containers,
// alongside the main deployment container.
// Use SidecarConfig struct to attach additonal volumes, service ports and service monitor
// scrape endpoints for these sidecars.
func WithSidecars(s SidecarConfig) DeploymentOption {
	return func(d *GenericDeploymentConfig) {
		d.Sidecars = s
	}
}
