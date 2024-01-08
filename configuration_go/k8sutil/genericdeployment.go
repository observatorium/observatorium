package k8sutil

import (
	"maps"
	"unicode/utf8"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeploymentGenericConfig represents a generic deployment configuration with common options.
// It groups those option fields in a flat structure for all the different Kubernetes objects that are created.
// It also provides helpers to generate the runtime objects from the configuration.
//
// Usage example:
//
//	// Create a new DeploymentGenericConfig object with your desired options
//	deploymentConfig := k8sutil.DeploymentGenericConfig{}
//
//	// Create the main container from the DeploymentGenericConfig object
//	// Customize the container if needed (e.g., add volumes, ports)
//	container := deploymentConfig.ToContainer()
//
//	// Generate the runtime objects from the DeploymentGenericConfig object
//	runtimeObjects := deploymentConfig.GenerateObjects(container)
type DeploymentGenericConfig struct {
	// Container fields
	ContainerResources corev1.ResourceRequirements
	Env                []corev1.EnvVar
	Image              string
	ImageTag           string
	ImagePullPolicy    corev1.PullPolicy
	LivenessProbe      *corev1.Probe
	ReadinessProbe     *corev1.Probe

	// Pod fields
	Affinity                      *corev1.Affinity
	SecurityContext               *corev1.PodSecurityContext
	TerminationGracePeriodSeconds int64

	// Deployment fields
	CommonLabels       map[string]string
	DeploymentStrategy appsv1.DeploymentStrategy // Only applies to Deployment kind
	Name               string
	Namespace          string
	Replicas           int32

	EnableServiceMonitor bool

	// Container dependencies
	// ConfigMaps and Secrets are the ones required by the main container, others are directly defined in Sidecars
	ConfigMaps map[string]map[string]string // maps a configmap name to its data of type map[string]string
	Secrets    map[string]map[string][]byte // maps a secret name to its data of type map[string][]byte
	Sidecars   []ContainerProvider
}

func (d DeploymentGenericConfig) ToContainer() *Container {
	return &Container{
		Name:            d.Name,
		Image:           d.Image,
		ImageTag:        d.ImageTag,
		ImagePullPolicy: d.ImagePullPolicy,
		Env:             d.Env,
		Resources:       d.ContainerResources,
		LivenessProbe:   d.LivenessProbe,
		ReadinessProbe:  d.ReadinessProbe,
		ConfigMaps:      d.ConfigMaps,
		Secrets:         d.Secrets,
	}
}

// GenerateObjects returns the list of runtime objects for the given container in a Deployment.
func (d DeploymentGenericConfig) GenerateObjects(container *Container) []runtime.Object {
	pod := d.Pod(container)
	ret := d.generateCommonObjects(pod)
	ret = append(ret, d.Deployment(pod))

	return ret
}

// GenerateObjectsStatefulSet returns the list of runtime objects for the given container in a StatefulSet.
func (d DeploymentGenericConfig) GenerateObjectsStatefulSet(container *Container) []runtime.Object {
	pod := d.Pod(container)
	ret := d.generateCommonObjects(pod)
	ret = append(ret, d.StatefulSet(pod))

	return ret
}

func (d DeploymentGenericConfig) ObjectMeta() *MetaConfig {
	labels := maps.Clone(d.CommonLabels)
	if d.ImageTag != "" {
		labels[VersionLabel] = d.ImageTag
	}

	return &MetaConfig{
		Name:      d.Name,
		Namespace: d.Namespace,
		Labels:    labels,
	}
}

// Pod returns a Pod object with the given container and sidecars.
func (d DeploymentGenericConfig) Pod(container *Container) *Pod {
	return &Pod{
		TerminationGracePeriodSeconds: &d.TerminationGracePeriodSeconds,
		Affinity:                      d.Affinity,
		SecurityContext:               d.SecurityContext,
		ServiceAccountName:            d.Name,
		ContainerProviders:            append([]ContainerProvider{container}, d.Sidecars...),
	}
}

// Deployment returns a Deployment object with the given pod.
func (d DeploymentGenericConfig) Deployment(pod *Pod) runtime.Object {
	dep := &Deployment{
		MetaConfig: *d.ObjectMeta(),
		Replicas:   int32(d.Replicas),
		Pod:        pod,
	}

	return dep.MakeManifest()
}

// StatefulSet returns a StatefulSet object with the given pod.
func (d DeploymentGenericConfig) StatefulSet(pod *Pod) runtime.Object {
	statefulset := &StatefulSet{
		MetaConfig: *d.ObjectMeta(),
		Replicas:   int32(d.Replicas),
		Pod:        pod,
	}

	return statefulset.MakeManifest()
}

// Service returns a Service object for the given pod.
func (d DeploymentGenericConfig) Service(pod *Pod) runtime.Object {
	service := &Service{
		MetaConfig:   *d.ObjectMeta(),
		ServicePorts: pod,
	}

	return service.MakeManifest()
}

// ServiceMonitor returns a ServiceMonitor object for the given pod.
func (d DeploymentGenericConfig) ServiceMonitor(pod *Pod) runtime.Object {
	serviceMonitor := &ServiceMonitor{
		MetaConfig:              *d.ObjectMeta(),
		ServiceMonitorEndpoints: pod,
	}

	return serviceMonitor.MakeManifest()
}

// ServiceAccount returns a ServiceAccount object.
func (d DeploymentGenericConfig) ServiceAccount() runtime.Object {
	serviceAccount := &ServiceAccount{
		MetaConfig: *d.ObjectMeta(),
		Name:       d.Name,
	}

	return serviceAccount.MakeManifest()
}

func (d DeploymentGenericConfig) RBACRole(rules []rbacv1.PolicyRule) runtime.Object {
	return &rbacv1.Role{
		TypeMeta:   RoleMeta,
		ObjectMeta: d.ObjectMeta().MakeMeta(),
		Rules:      rules,
	}
}

func (d DeploymentGenericConfig) RBACRoleBinding(subjects []runtime.Object, role runtime.Object) runtime.Object {
	subs := make([]rbacv1.Subject, len(subjects))
	for i, s := range subjects {
		subMeta, ok := s.(metav1.Object)
		if !ok {
			panic("subject does not implement metav1.Object")
		}

		subs[i] = rbacv1.Subject{
			Kind:      s.GetObjectKind().GroupVersionKind().Kind,
			Name:      subMeta.GetName(),
			Namespace: subMeta.GetNamespace(),
		}
	}

	return &rbacv1.RoleBinding{
		TypeMeta:   RoleBindingMeta,
		ObjectMeta: d.ObjectMeta().MakeMeta(),
		Subjects:   subs,
		RoleRef: rbacv1.RoleRef{
			Kind:     role.GetObjectKind().GroupVersionKind().Kind,
			APIGroup: role.GetObjectKind().GroupVersionKind().Group,
			Name:     role.(metav1.Object).GetName(),
		},
	}
}

// ConfigMapsAndSecrets returns the list of ConfigMap and Secret objects for the given pod.
func (d DeploymentGenericConfig) ConfigMapsAndSecrets(pod *Pod) []runtime.Object {
	ret := []runtime.Object{}
	for name, data := range pod.GetConfigMaps() {
		cm := &corev1.ConfigMap{
			TypeMeta:   ConfigMapMeta,
			ObjectMeta: d.ObjectMeta().MakeMeta(),
			Data:       data,
		}
		cm.Name = name
		ret = append(ret, cm)
	}

	for name, data := range pod.GetSecrets() {
		secret := &corev1.Secret{
			TypeMeta:   SecretMeta,
			ObjectMeta: d.ObjectMeta().MakeMeta(),
		}
		secret.Name = name

		// check if data is a string and store it as a stringData if possible for better readability
		stringData := map[string]string{}
		isStringData := true
		for k, v := range data {
			if utf8.Valid(v) {
				stringData[k] = string(v)
			} else {
				isStringData = false
				break
			}
		}

		if isStringData {
			secret.StringData = stringData
		} else {
			secret.Data = data
		}
		ret = append(ret, secret)
	}

	return ret
}

func (d DeploymentGenericConfig) generateCommonObjects(pod *Pod) []runtime.Object {
	ret := []runtime.Object{
		d.ServiceAccount(),
	}

	if len(pod.GetServicePorts()) > 0 {
		ret = append(ret, d.Service(pod))
	}

	if d.EnableServiceMonitor && len(pod.GetServiceMonitorEndpoints()) > 0 {
		ret = append(ret, d.ServiceMonitor(pod))
	}

	ret = append(ret, d.ConfigMapsAndSecrets(pod)...)

	return ret
}
