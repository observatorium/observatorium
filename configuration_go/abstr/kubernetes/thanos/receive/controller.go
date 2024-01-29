package receive

import (
	"fmt"
	"strings"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	workload.DeploymentWorkload
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
		workload.NameLabel:      "thanos-receive-controller",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "kubernetes-controller",
		workload.VersionLabel:   imageTag,
	}

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Name:            fmt.Sprintf("%s-%s", commonLabels[workload.InstanceLabel], commonLabels[workload.NameLabel]),
			Image:           ThanosReceiveControllerImage,
			ImageTag:        imageTag,
			Namespace:       namespace,
			CommonLabels:    commonLabels,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env: []corev1.EnvVar{
				kghelpers.NewEnvFromField("NAMESPACE", "metadata.namespace"),
			},
			ContainerResources: kghelpers.NewResourcesRequirements("10m", "24Mi", "64m", "128Mi"),
			ConfigMaps:         map[string]map[string]string{},
			Secrets:            map[string]map[string][]byte{},
		},
	}

	return &Controller{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (c *Controller) Objects() []runtime.Object {
	container := c.makeContainer()

	ret := c.DeploymentWorkload.Objects(container)

	// create role
	ret = append(ret, &rbacv1.Role{
		TypeMeta: workload.RoleMeta,
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
	apiGroup := strings.Split(workload.RoleMeta.APIVersion, "/")[0]
	ret = append(ret, &rbacv1.RoleBinding{
		TypeMeta: workload.RoleBindingMeta,
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
			Kind:     workload.RoleMeta.Kind,
			Name:     c.Name,
			APIGroup: apiGroup,
		},
	})

	return ret
}

func (c *Controller) makeContainer() *workload.Container {
	container := c.ToContainer()
	container.Env = append(container.Env, kghelpers.NewEnvFromField("NAMESPACE", "metadata.namespace"))

	container.Args = cmdopt.GetOpts(c.options)

	return container
}
