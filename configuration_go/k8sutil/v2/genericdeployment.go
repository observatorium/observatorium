package k8sutil

import (
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Resource represents the resource requirements for a container.
type Resource struct {
	Requests, Limits string
}

// ProbeConfig represents the configuration of a container probe (liveness or readiness).
type ProbeConfig struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
}

// ServiceMonitorConfig represents the configuration of a ServiceMonitor.
type ServiceMonitorConfig struct {
	Enabled   bool
	Namespace string
	Labels    map[string]string
}

// CommonConfig represents the common configuration for a container.
// It is applicable to both Deployment and StatefulSet.
type CommonConfig struct {
	Name                          string
	Namespace                     string
	Labels                        map[string]string
	Image                         string
	ImageTag                      string
	ImagePullPolicy               corev1.PullPolicy
	Replicas                      int32
	Affinity                      *corev1.Affinity
	TerminationMessagePolicy      corev1.TerminationMessagePolicy
	TerminationGracePeriodSeconds int64
	SecurityContext               corev1.PodSecurityContext
	ServiceAccountName            string
	Resources                     corev1.ResourceRequirements
	ServiceMonitor                ServiceMonitorConfig
	LivenessProbe                 ProbeConfig
	ReadinessProbe                ProbeConfig
	Env                           []corev1.EnvVar

	// Configuration to add containers, volumes, service and servicemonitor ports in addition to the default ones.
	SideCars                []corev1.Container
	PodVolumes              []corev1.Volume
	ServicePorts            []corev1.ServicePort
	ServiceMonitorEndpoints []monv1.Endpoint
}

// GetDefaultServiceMonitorRelabelConfig returns the default relabel config for a ServiceMonitor.
func GetDefaultServiceMonitorRelabelConfig() []*monv1.RelabelConfig {
	return []*monv1.RelabelConfig{
		{
			Action:       "replace",
			Separator:    "/",
			SourceLabels: []monv1.LabelName{"namespace", "pod"},
			TargetLabel:  "instance",
		},
	}
}

// GetDefaultSecurityContext returns the default security context for a container.
func GetDefaultSecurityContext() corev1.PodSecurityContext {
	return corev1.PodSecurityContext{
		RunAsUser: int64Ptr(65534),
		FSGroup:   int64Ptr(65534),
	}
}

// NewServiceMonitor creates a new ServiceMonitor object.
func NewServiceMonitor(name, servicePortName, namespace string, labels, matchLabels map[string]string) *monv1.ServiceMonitor {
	endpoints := []monv1.Endpoint{
		{
			Port: servicePortName,
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
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
}

func NewResourcesRequirements(cpuRequest, cpuLimit, memoryRequest, memoryLimit string) corev1.ResourceRequirements {
	ret := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}

	setResourcesRequirements(ret.Requests, corev1.ResourceCPU, cpuRequest)
	setResourcesRequirements(ret.Limits, corev1.ResourceCPU, cpuLimit)
	setResourcesRequirements(ret.Requests, corev1.ResourceMemory, memoryRequest)
	setResourcesRequirements(ret.Limits, corev1.ResourceMemory, memoryLimit)

	return ret
}

func setResourcesRequirements(resList corev1.ResourceList, reourceName corev1.ResourceName, value string) {
	if value == "" {
		return
	}

	resList[reourceName] = resource.MustParse(value)
}

func NewProbe(path string, port int, cfg ProbeConfig) *corev1.Probe {
	ret := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/-/healthy",
				Port: intstr.FromInt(port),
			},
		},
		FailureThreshold:    cfg.FailureThreshold,
		InitialDelaySeconds: cfg.InitialDelaySeconds,
		PeriodSeconds:       cfg.PeriodSeconds,
		SuccessThreshold:    cfg.SuccessThreshold,
		TimeoutSeconds:      cfg.TimeoutSeconds,
	}

	return ret
}

func NewAntiAffinity(namespaces []string, labelSelectors map[string]string) *corev1.Affinity {
	matchExpressions := []metav1.LabelSelectorRequirement{}

	for k, v := range labelSelectors {
		matchExpressions = append(matchExpressions, metav1.LabelSelectorRequirement{
			Key:      k,
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{v},
		})
	}

	ret := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: k8sutil.HostnameLabel,
						Namespaces:  namespaces,
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: matchExpressions,
						},
					},
				},
			},
		},
	}

	return ret
}

func NewEnvFromSecret(envName, secretKey, secretName string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: secretKey,
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
			},
		},
	}
}

func NewEnvFromField(envName, fieldPath string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

func NewPodVolumeFromSecret(name, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func int64Ptr(i int64) *int64 { return &i }
