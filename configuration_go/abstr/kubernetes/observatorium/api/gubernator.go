package api

import (
	"fmt"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	gubernatorHttpPort = 8080
	gubernatorGrpcPort = 8081
)

type GubernatorDeployment struct {
	k8sutil.DeploymentGenericConfig
}

func NewGubernatorDeployment(namespace, imageTag string) *GubernatorDeployment {
	commonLabels := map[string]string{
		k8sutil.NameLabel:      "gubernator",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "rate-limiter",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &GubernatorDeployment{
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "ghcr.io/mailgun/gubernator",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-gubernator",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("300m", "600m", "100Mi", "200Mi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", gubernatorHttpPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", gubernatorHttpPort, k8sutil.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				k8sutil.NewEnvFromField("GUBER_K8S_NAMESPACE", "metadata.namespace"),
				k8sutil.NewEnvFromField("GUBER_K8S_POD_IP", "status.podIP"),
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
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}
}

func (g *GubernatorDeployment) Manifests() k8sutil.ObjectMap {
	container := g.makeContainer()

	ret := k8sutil.ObjectMap{}

	ret.AddAll(g.GenerateObjects(container))
	k8sutil.GetObject[*corev1.Service](ret, g.Name).Spec.ClusterIP = corev1.ClusterIPNone

	rbacRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"endpoints"},
			Verbs:     []string{"list", "watch", "get"},
		},
	}
	rbacRole := g.RBACRole(rbacRules)
	ret.Add(rbacRole)
	sa := k8sutil.GetObject[*corev1.ServiceAccount](ret, g.Name)
	ret.Add(g.RBACRoleBinding([]runtime.Object{sa}, rbacRole))

	return ret
}

func (s *GubernatorDeployment) makeContainer() *k8sutil.Container {
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
		k8sutil.NewServicePort("http", gubernatorHttpPort, gubernatorHttpPort),
		k8sutil.NewServicePort("grpc", gubernatorGrpcPort, gubernatorGrpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	return ret
}
