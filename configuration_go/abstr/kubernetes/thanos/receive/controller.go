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
	options *ControllerOptions

	k8sutil.DeploymentGenericConfig
}

func NewControllerDefaultOptions() *ControllerOptions {
	return &ControllerOptions{
		ConfigMapName:          "observatorium-thanos-receive-controller",
		ConfigMapGeneratedName: "observatorium-thanos-receive-controller-generated",
		FileName:               "hashrings.json",
		Namespace:              "observatorium",
	}
}

func NewController(opts *ControllerOptions, namespace, imageTag string) *Controller {
	if opts == nil {
		opts = NewControllerDefaultOptions()
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-receive-controller",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "kubernetes-controller",
		k8sutil.VersionLabel:   imageTag,
	}

	return &Controller{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Name:            fmt.Sprintf("%s-%s", commonLabels[k8sutil.InstanceLabel], commonLabels[k8sutil.NameLabel]),
			Image:           ThanosReceiveControllerImage,
			ImageTag:        imageTag,
			Namespace:       namespace,
			Replicas:        1,
			CommonLabels:    commonLabels,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{
				k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"),
			},
			ContainerResources: k8sutil.NewResourcesRequirements("10m", "24Mi", "64m", "128Mi"),
			ConfigMaps:         map[string]map[string]string{},
			Secrets:            map[string]map[string][]byte{},
		},
	}
}

func (c *Controller) Manifests() k8sutil.ObjectMap {
	container := c.makeContainer()

	ret := k8sutil.ObjectMap{}
	ret.AddAll(c.GenerateObjects(container))

	// create role
	ret.Add(&rbacv1.Role{
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
	})

	// create role binding
	apiGroup := strings.Split(k8sutil.RoleMeta.APIVersion, "/")[0]
	ret.Add(&rbacv1.RoleBinding{
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
	})

	return ret
}

func (c *Controller) makeContainer() *k8sutil.Container {
	container := c.DeploymentGenericConfig.ToContainer()
	container.Env = append(container.Env, k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"))

	container.Args = cmdopt.GetOpts(c.options)

	return container
}
