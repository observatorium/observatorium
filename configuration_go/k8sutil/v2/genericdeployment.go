package k8sutil

import (
	"fmt"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ExtraConfig represents the configuration required to add extra containers, volumes
// service and servicemonitor ports to a particular Deployment/StatefulSet.
type ExtraConfig struct {
	Sidecars                      []corev1.Container
	AdditionalPodVolumes          []corev1.Volume
	AdditionalServicePorts        []corev1.ServicePort
	AdditionalServiceMonitorPorts []monv1.Endpoint
}

// DeploymentGenericConfig represents certain config fields
// that might be useful to add/override in a Deployment. It contains
// fields of both DeploymentSpec and PodSpec.
// It also has method defined for overriding any default values.
// type DeploymentGenericConfig struct {
// 	appsv1.Deployment
// }

type DeploymentModifier func(d *appsv1.Deployment)

// WithImage overrides the default image.
func WithImage(image, imageTag string) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, imageTag)
	}
}

// WithImagePullPolicy overrides default image pull policy.
func WithImagePullPolicy(imagePullPolicy corev1.PullPolicy) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].ImagePullPolicy = imagePullPolicy
	}
}

// WithReplicas overrides the default number of replicas to run.
func WithReplicas(replicas int32) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Replicas = &replicas
	}
}

// WithDeploymentStrategy overrides the default deployment strategy of pods.
func WithDeploymentStrategy(ds appsv1.DeploymentStrategy) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Strategy = ds
	}
}

// WithResources overrides the default Pod resource config.
func WithCPUResources(requests, limits string) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		setResources(d, corev1.ResourceCPU, requests, limits)
	}
}

// WithResources overrides the default Pod resource config.
func WithMemoryResources(requests, limits string) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		setResources(d, corev1.ResourceMemory, requests, limits)
	}
}

func setResources(d *appsv1.Deployment, reourceName corev1.ResourceName, request, limits string) {
	if d.Spec.Template.Spec.Containers[0].Resources.Requests == nil {
		d.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{}
	}

	if d.Spec.Template.Spec.Containers[0].Resources.Limits == nil {
		d.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{}
	}

	d.Spec.Template.Spec.Containers[0].Resources.Requests[reourceName] = resource.MustParse(request)
	d.Spec.Template.Spec.Containers[0].Resources.Limits[reourceName] = resource.MustParse(limits)
}

// WithAffinity overrides the default Pod scheduling affinity rules.
func WithAffinity(affinity corev1.Affinity) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Affinity = &affinity
	}
}

// WithSecurityContext overrides the default Pod security context.
func WithSecurityContext(sc corev1.PodSecurityContext) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.SecurityContext = &sc
	}
}

// WithProbe overrides the default K8s liveness and readiness probes of main deployment container.
func WithProbes(livenessProbe, readinessProbe corev1.Probe) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].LivenessProbe = &livenessProbe
		d.Spec.Template.Spec.Containers[0].ReadinessProbe = &readinessProbe
	}
}

// WithTerminationMessagePolicy overrides the default termination message policy of main deployment container.
func WithTerminationMessagePolicy(terminationMessagePolicy corev1.TerminationMessagePolicy) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers[0].TerminationMessagePolicy = terminationMessagePolicy
	}
}

// WithTerminationGracePeriod overrides the default termination grace period of pod.
func WithTerminationGracePeriod(duration int64) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.TerminationGracePeriodSeconds = &duration
	}
}

// // WithExtra allows overriding default K8s Deployment and adding additional containers,
// // alongside the main deployment container, and allows attaching additonal volumes,
// // service ports and service monitor scrape endpoints to these deployments.
// func WithExtras(e ExtraConfig) DeploymentOption {
// 	return func(d *DeploymentGenericConfig) {
// 		d.Extras = e
// 	}
// }

// WithServiceMonitor enables generation of a ServiceMonitor to scrape the deployment.
// func WithServiceMonitor() DeploymentOption {
// 	return func(d *DeploymentGenericConfig) {
// 		d.EnableServiceMonitor = true
// 	}
// }

// // WithName overrides the default name of all the individual objects.
// func WithName(name string) DeploymentOption {
// 	return func(d *DeploymentGenericConfig) {
// 		d.Name = name
// 	}
// }

// // WithNamespace overrides the default namespace of all the individual objects.
// func WithNamespace(namespace string) DeploymentOption {
// 	return func(d *DeploymentGenericConfig) {
// 		d.Spec.Template.ObjectMeta.Namespace = namespace
// 	}
// }

// // WithCommonLabels overrides the default K8s metadata labels and selectors.
// func WithCommonLabels(commonLabels map[string]string) DeploymentOption {
// 	return func(d *DeploymentGenericConfig) {
// 		d.CommonLabels = commonLabels
// 	}
// }

func WithSideCars(sidecars ...corev1.Container) DeploymentModifier {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, sidecars...)
	}
}

type CommonConfig struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

func NewServiceMonitor(cfg *CommonConfig, matchLabels map[string]string) *monv1.ServiceMonitor {
	endpoints := []monv1.Endpoint{
		{
			Port: "http",
			RelabelConfigs: []*monv1.RelabelConfig{
				{
					Action:       "replace",
					Separator:    "/",
					SourceLabels: []monv1.LabelName{"namespace", "pod"},
					TargetLabel:  "instance",
				},
			},
		},
	}

	spec := monv1.ServiceMonitorSpec{
		Endpoints: endpoints,
		Selector: metav1.LabelSelector{
			MatchLabels: matchLabels,
		},
	}

	return &monv1.ServiceMonitor{
		TypeMeta: k8sutil.ServiceMonitorMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: spec,
	}
}

func NewProb(path string, port, failure, period int) *corev1.Probe {
	ret := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/-/healthy",
				Port: intstr.FromInt(port),
			},
		},
	}

	if failure != 0 {
		ret.FailureThreshold = int32(failure)
	}

	if period != 0 {
		ret.PeriodSeconds = int32(period)
	}

	return ret
}

func NewSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser: int64Ptr(65534),
		FSGroup:   int64Ptr(65534),
	}
}

func int64Ptr(i int64) *int64 { return &i }
