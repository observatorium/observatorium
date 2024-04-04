package api

import (
	"fmt"

	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	gubernatorHttpPort = 8080
	gubernatorGrpcPort = 8081
)

type GubernatorDeployment struct {
	workload.DeploymentWorkload
}

func NewGubernatorDeployment(namespace, imageTag string) *GubernatorDeployment {
	commonLabels := map[string]string{
		workload.NameLabel:      "gubernator",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "rate-limiter",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Image:                "ghcr.io/mailgun/gubernator",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-gubernator",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("300m", "600m", "100Mi", "200Mi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: kghelpers.NewProbe("/-/healthy", gubernatorHttpPort, kghelpers.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: kghelpers.NewProbe("/-/ready", gubernatorHttpPort, kghelpers.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				kghelpers.NewEnvFromField("GUBER_K8S_NAMESPACE", "metadata.namespace"),
				kghelpers.NewEnvFromField("GUBER_K8S_POD_IP", "status.podIP"),
				{
					Name:  "GUBER_HTTP_ADDRESS",
					Value: fmt.Sprintf("0.0.0.0:%d", gubernatorHttpPort),
				},
				{
					Name:  "GUBER_GRPC_ADDRESS",
					Value: fmt.Sprintf("0.0.0.0:%d", gubernatorGrpcPort),
				},
				{
					Name:  "GUBER_K8S_POD_PORT",
					Value: fmt.Sprintf("%d", gubernatorGrpcPort),
				},
				{
					Name:  "GUBER_K8S_ENDPOINTS_SELECTOR",
					Value: "app.kubernetes.io/name=gubernator",
				},
				{
					Name:  "GUBER_PEER_DISCOVERY_TYPE",
					Value: "k8s",
				},
				{
					Name:  "GUBER_LOG_LEVEL",
					Value: "info",
				},
				{
					Name:  "OTEL_TRACES_EXPORTER",
					Value: "none",
				},
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}

	return &GubernatorDeployment{
		DeploymentWorkload: depWorkload,
	}
}

func (g *GubernatorDeployment) Objects() []runtime.Object {
	container := g.makeContainer()
	ret := g.DeploymentWorkload.Objects(container)

	kghelpers.GetObject[*corev1.Service](ret, g.Name).Spec.ClusterIP = corev1.ClusterIPNone

	rbacRole := &rbacv1.Role{
		TypeMeta:   workload.RoleMeta,
		ObjectMeta: g.ObjectMeta().MakeMeta(),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"list", "watch", "get"},
			},
		},
	}

	sa := kghelpers.GetObject[*corev1.ServiceAccount](ret, g.Name)
	roleBinding := &rbacv1.RoleBinding{
		TypeMeta:   workload.RoleBindingMeta,
		ObjectMeta: g.ObjectMeta().MakeMeta(),
		Subjects: []rbacv1.Subject{
			{
				Kind:      sa.GetObjectKind().GroupVersionKind().Kind,
				Name:      sa.GetName(),
				Namespace: sa.GetNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     rbacRole.GetObjectKind().GroupVersionKind().Kind,
			APIGroup: rbacRole.GetObjectKind().GroupVersionKind().Group,
			Name:     rbacRole.GetName(),
		},
	}

	ret = append(ret, rbacRole, roleBinding)

	return ret
}

func (s *GubernatorDeployment) makeContainer() *workload.Container {
	ret := s.ToContainer()
	ret.Name = "gubernator"
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(gubernatorHttpPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "grpc",
			ContainerPort: int32(gubernatorGrpcPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http", gubernatorHttpPort, gubernatorHttpPort),
		kghelpers.NewServicePort("grpc", gubernatorGrpcPort, gubernatorGrpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	return ret
}
