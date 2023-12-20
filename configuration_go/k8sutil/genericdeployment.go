package k8sutil

import (
	"maps"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeploymentGenericConfig represents certain config fields
// that might be useful to add/override in a Deployment/StatefulSet. It contains
// fields of both DeploymentSpec and PodSpec.
type DeploymentGenericConfig struct {
	// Container fields
	Image                string
	ImageTag             string
	ImagePullPolicy      corev1.PullPolicy
	Name                 string
	Namespace            string
	CommonLabels         map[string]string
	Replicas             int32
	DeploymentStrategy   appsv1.DeploymentStrategy // Only applies to Deployment kind
	PodResources         corev1.ResourceRequirements
	Affinity             *corev1.Affinity
	SecurityContext      *corev1.PodSecurityContext
	EnableServiceMonitor bool
	Env                  []corev1.EnvVar
	LivenessProbe        *corev1.Probe
	ReadinessProbe       *corev1.Probe

	// Pod fields
	TerminationMessagePolicy      corev1.TerminationMessagePolicy
	TerminationGracePeriodSeconds int64

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
		Resources:       d.PodResources,
		LivenessProbe:   d.LivenessProbe,
		ReadinessProbe:  d.ReadinessProbe,
		ConfigMaps:      d.ConfigMaps,
		Secrets:         d.Secrets,
	}
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

func (d DeploymentGenericConfig) Pod(container *Container) *Pod {
	return &Pod{
		TerminationGracePeriodSeconds: &d.TerminationGracePeriodSeconds,
		Affinity:                      d.Affinity,
		SecurityContext:               d.SecurityContext,
		ServiceAccountName:            d.Name,
		ContainerProviders:            append([]ContainerProvider{container}, d.Sidecars...),
	}
}

func (d DeploymentGenericConfig) Deployment(pod *Pod) runtime.Object {
	dep := &Deployment{
		MetaConfig: *d.ObjectMeta(),
		Replicas:   int32(d.Replicas),
		Pod:        pod,
	}

	return dep.MakeManifest()
}

func (d DeploymentGenericConfig) StatefulSet(pod *Pod) runtime.Object {
	statefulset := &StatefulSet{
		MetaConfig: *d.ObjectMeta(),
		Replicas:   int32(d.Replicas),
		Pod:        pod,
	}

	return statefulset.MakeManifest()
}

func (d DeploymentGenericConfig) Service(pod *Pod) runtime.Object {
	service := &Service{
		MetaConfig:   *d.ObjectMeta(),
		ServicePorts: pod,
	}

	return service.MakeManifest()
}

func (d DeploymentGenericConfig) ServiceMonitor(pod *Pod) runtime.Object {
	serviceMonitor := &ServiceMonitor{
		MetaConfig:              *d.ObjectMeta(),
		ServiceMonitorEndpoints: pod,
	}

	return serviceMonitor.MakeManifest()
}

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

func (d DeploymentGenericConfig) ConfigMapsAndSecrets(pod *Pod) []runtime.Object {
	ret := []runtime.Object{}
	for name, data := range d.ConfigMaps {
		cm := &corev1.ConfigMap{
			TypeMeta:   ConfigMapMeta,
			ObjectMeta: d.ObjectMeta().MakeMeta(),
			Data:       data,
		}
		cm.Name = name
		ret = append(ret, cm)
	}

	for name, data := range d.Secrets {
		secret := &corev1.Secret{
			TypeMeta:   SecretMeta,
			ObjectMeta: d.ObjectMeta().MakeMeta(),
			Data:       data,
		}
		secret.Name = name
		ret = append(ret, secret)
	}

	return ret
}

func (d DeploymentGenericConfig) GenerateObjects(container *Container) []runtime.Object {
	pod := d.Pod(container)
	ret := d.generateCommonObjects(pod)
	ret = append(ret, d.Deployment(pod))

	return ret
}

func (d DeploymentGenericConfig) GenerateObjectsStatefulSet(container *Container) []runtime.Object {
	pod := d.Pod(container)
	ret := d.generateCommonObjects(pod)
	ret = append(ret, d.StatefulSet(pod))

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

// type ObjectProcessor func(obj runtime.Object)
type DeploymentOption func(d *DeploymentGenericConfig)

// WithImage overrides the default image.
func WithImage(image, imageTag string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Image = image
		d.ImageTag = imageTag
	}
}

// WithImagePullPolicy overrides default image pull policy.
func WithImagePullPolicy(imagePullPolicy corev1.PullPolicy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.ImagePullPolicy = imagePullPolicy
	}
}

// WithName overrides the default name of all the individual objects.
func WithName(name string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Name = name
	}
}

// WithNamespace overrides the default namespace of all the individual objects.
func WithNamespace(namespace string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Namespace = namespace
	}
}

// WithCommonLabels overrides the default K8s metadata labels and selectors.
func WithCommonLabels(commonLabels map[string]string) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.CommonLabels = commonLabels
	}
}

// WithReplicas overrides the default number of replicas to run.
func WithReplicas(replicas int32) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Replicas = replicas
	}
}

// WithDeploymentStrategy overrides the default deployment strategy of pods.
func WithDeploymentStrategy(ds appsv1.DeploymentStrategy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.DeploymentStrategy = ds
	}
}

// WithResources overrides the default Pod resource config.
func WithResources(resource corev1.ResourceRequirements) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.PodResources = resource
	}
}

// WithAffinity overrides the default Pod scheduling affinity rules.
func WithAffinity(affinity *corev1.Affinity) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Affinity = affinity
	}
}

// WithSecurityContext overrides the default Pod security context.
func WithSecurityContext(sc *corev1.PodSecurityContext) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.SecurityContext = sc
	}
}

// WithServiceMonitor enables generation of a ServiceMonitor to scrape the deployment.
func WithServiceMonitor() DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.EnableServiceMonitor = true
	}
}

// WithProbe overrides the default K8s liveness and readiness probes of main deployment container.
func WithProbes(livenessProbe, readinessProbe *corev1.Probe) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.LivenessProbe = livenessProbe
		d.ReadinessProbe = readinessProbe
	}
}

// WithTerminationMessagePolicy overrides the default termination message policy of main deployment container.
func WithTerminationMessagePolicy(terminationMessagePolicy corev1.TerminationMessagePolicy) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.TerminationMessagePolicy = terminationMessagePolicy
	}
}

// WithTerminationGracePeriod overrides the default termination grace period of main deployment container.
func WithTerminationGracePeriod(duration int64) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.TerminationGracePeriodSeconds = duration
	}
}

// WithSideCars overrides the default pod sidecars.
func WithSidecars(sidecars ...ContainerProvider) DeploymentOption {
	return func(d *DeploymentGenericConfig) {
		d.Sidecars = sidecars
	}
}
