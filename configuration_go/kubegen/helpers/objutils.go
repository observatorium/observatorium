package helpers

import (
	"fmt"
	"sort"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const hostnameLabel string = "kubernetes.io/hostname"

// GetObject returns the object of type T from the given list of kubernetes objects.
// When specifying a name, it will return the object with the given name.
// This helper can be used for doing post processing on the objects.
func GetObject[T metav1.Object](objects []runtime.Object, name string) T {
	var ret T
	found := false

	for _, obj := range objects {
		if typedObject, ok := obj.(T); ok {
			if name != "" && typedObject.GetName() != name {
				continue
			}

			// Check if we already found an object of this type. If so, panic.
			if found {
				panic(fmt.Sprintf("found multiple objects of type %T", *new(T)))
			}

			ret = typedObject
			found = true
		}
	}

	if !found {
		panic(fmt.Sprintf("could not find object of type %T", *new(T)))
	}

	return ret
}

// NewResourcesRequirements returns a new resource requirements object for a container.
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

// ProbeConfig represents the configuration of a container probe (liveness or readiness).
type ProbeConfig struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
}

// NewProbe returns a new probe.
func NewProbe(path string, port int, cfg ProbeConfig) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt(port),
			},
		},
		FailureThreshold:    cfg.FailureThreshold,
		InitialDelaySeconds: cfg.InitialDelaySeconds,
		PeriodSeconds:       cfg.PeriodSeconds,
		SuccessThreshold:    cfg.SuccessThreshold,
		TimeoutSeconds:      cfg.TimeoutSeconds,
	}
}

// NewAntiAffinity returns a new anti-affinity rule.
func NewAntiAffinity(namespaces []string, labelSelectors map[string]string) *corev1.Affinity {
	matchExpressions := []metav1.LabelSelectorRequirement{}

	for k, v := range labelSelectors {
		matchExpressions = append(matchExpressions, metav1.LabelSelectorRequirement{
			Key:      k,
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{v},
		})
	}

	// Sort to avoid random order in generated YAML which generates noisy diffs.
	sort.Slice(matchExpressions, func(i, j int) bool {
		return matchExpressions[i].Key < matchExpressions[j].Key
	})

	ret := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: hostnameLabel,
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

// NewEnvFromSecret returns a new environment variable from a secret.
func NewEnvFromSecret(envName, secretName, secretKey string) corev1.EnvVar {
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

// NewEnvFromField returns a new environment variable from a field.
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

// NewPodVolumeFromSecret returns a new pod volume from a secret.
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

// NewPodVolumeFromConfigMap returns a new pod volume from a config map.
func NewPodVolumeFromConfigMap(name, configMapName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
}

// NewServicePort returns a new service port.
func NewServicePort(name string, port, targetPort int) corev1.ServicePort {
	return corev1.ServicePort{
		Name:       name,
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
		Protocol:   corev1.ProtocolTCP,
	}
}

// GetDefaultServiceMonitorRelabelConfig returns the default relabel config for a ServiceMonitor.
func GetDefaultServiceMonitorRelabelConfig() []monv1.RelabelConfig {
	separator := "/"
	return []monv1.RelabelConfig{
		{
			Action:       "replace",
			Separator:    &separator,
			SourceLabels: []monv1.LabelName{"namespace", "pod"},
			TargetLabel:  "instance",
		},
	}
}
