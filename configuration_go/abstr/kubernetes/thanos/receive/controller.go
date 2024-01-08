package receive

import (
	"fmt"
	"strings"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ThanosReceiveControllerImage string = "quay.io/observatorium/thanos-receive-controller"
)

type ControllerOptions struct {
	ConfigMapName          string `opt:"configmap-name"`
	ConfigMapGeneratedName string `opt:"configmap-generated-name"`
	FileName               string `opt:"file-name"`
	Namespace              string `opt:"namespace"`
}

type Controller struct {
	Options ControllerOptions
	k8sutil.DeploymentGenericConfig
}

func NewController() *Controller {
	options := ControllerOptions{
		ConfigMapName:          "observatorium-thanos-receive-controller",
		ConfigMapGeneratedName: "observatorium-thanos-receive-controller-generated",
		FileName:               "hashrings.json",
		Namespace:              "observatorium",
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-receive-controller",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "kubernetes-controller",
	}

	genericDeployment := k8sutil.DeploymentGenericConfig{
		Name:            fmt.Sprintf("%s-%s", commonLabels[k8sutil.InstanceLabel], commonLabels[k8sutil.NameLabel]),
		Image:           ThanosReceiveControllerImage,
		ImageTag:        "main",
		Replicas:        1,
		CommonLabels:    commonLabels,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"),
		},
		ContainerResources: k8sutil.NewResourcesRequirements("10m", "24Mi", "64m", "128Mi"),
		ConfigMaps:         map[string]map[string]string{},
		Secrets:            map[string]map[string][]byte{},
	}

	return &Controller{
		options,
		genericDeployment,
	}
}

func (c *Controller) Manifests() k8sutil.ObjectMap {
	container := c.makeContainer()
	ret := k8sutil.ObjectMap{}

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      c.Name,
		Labels:    c.CommonLabels,
		Namespace: c.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &c.TerminationGracePeriodSeconds,
		SecurityContext:               c.SecurityContext,
		ServiceAccountName:            c.Name,
		ContainerProviders:            []k8sutil.ContainerProvider{container},
	}

	deployment := &k8sutil.Deployment{
		MetaConfig: commonObjectMeta,
		Replicas:   1,
		Pod:        pod,
	}

	ret["controller-deployment"] = deployment.MakeManifest()

	// service := &k8sutil.Service{
	// 	MetaConfig:   commonObjectMeta.Clone(),
	// 	ServicePorts: pod,
	// }
	// ret["controller-service"] = service.MakeManifest()

	if c.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["controller-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["controller-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["controller-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["controller-secret-"+name] = secret.MakeManifest()
	}

	// create role
	ret["controller-role"] = &rbacv1.Role{
		TypeMeta: k8sutil.RoleMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Labels:    c.CommonLabels,
			Namespace: c.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"list", "watch", "get", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},

				Verbs: []string{"get", "update"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets"},
				Verbs:     []string{"list", "watch", "get"},
			},
		},
	}

	// create role binding
	apiGroup := strings.Split(k8sutil.RoleMeta.APIVersion, "/")[0]
	ret["controller-rolebinding"] = &rbacv1.RoleBinding{
		TypeMeta: k8sutil.RoleBindingMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Labels:    c.CommonLabels,
			Namespace: c.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      c.Name,
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     k8sutil.RoleMeta.Kind,
			Name:     c.Name,
			APIGroup: apiGroup,
		},
	}

	return ret
}

func (c *Controller) makeContainer() *k8sutil.Container {
	container := c.DeploymentGenericConfig.ToContainer()
	container.Env = append(container.Env, k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"))

	container.Args = cmdopt.GetOpts(c.Options)

	return container
}
