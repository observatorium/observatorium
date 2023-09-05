package k8sutil

import (
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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
	val := int64(65534)
	return corev1.PodSecurityContext{
		RunAsUser: &val,
		FSGroup:   &val,
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

// ProbeConfig represents the configuration of a container probe (liveness or readiness).
type ProbeConfig struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
}

// NewProbe returns a new probe.
func NewProbe(path string, port int, cfg ProbeConfig) corev1.Probe {
	return corev1.Probe{
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

	ret := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: HostnameLabel,
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

// NewVolumeClaimProvider returns a new volume claim.
func NewVolumeClaimProvider(name, volumeType, size string) VolumeClaim {
	return VolumeClaim{
		Name: name,
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
			StorageClassName: &volumeType,
		},
	}
}
