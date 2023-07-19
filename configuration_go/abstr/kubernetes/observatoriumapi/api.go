package observatoriumapi

import (
	"fmt"

	"github.com/bwplotka/mimic"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type observatoriumAPI struct {
	// API configs.
	logLevel          string
	rbacYAML          string
	tenantsSecret     map[string]string
	metrics           MetricsBackend
	logs              LogsBackend
	traces            TracesBackend
	rateLimiter       string
	additionalAPIArgs []string

	// Embedded K8s config struct which exposes override methods.
	k8sutil.DeploymentGenericConfig
}

type MetricsBackend struct {
	ReadEndpoint  string
	WriteEndpoint string
	RulesEndpoint string
}

type LogsBackend struct {
	ReadEndpoint  string
	WriteEndpoint string
	TailEndpoint  string
	RulesEndpoint string
}

type TracesBackend struct {
	ReadEndpoint  string
	WriteEndpoint string
}

// Allows specifying external functions for overriding Observatorium API options.
type ObservatoriumAPIOption func(a *observatoriumAPI)

// WithAdditionalAPIArgs allows including additional arguments to the Observatorium API deployment.
func WithAdditionalAPIArgs(additionalAPIArgs []string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.additionalAPIArgs = additionalAPIArgs
	}
}

// WithRBACYAML overrides the empty RBAC in ConfigMap with the passed in one.
// Takes in string instead of https://pkg.go.dev/github.com/observatorium/api/rbac types in
// case this needs to be empty/substituted.
func WithRBACYAML(rbacYAML string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.rbacYAML = rbacYAML
	}
}

// WithTenantsSecret overrides the empty tenants Secret with the passed in map.
// Ensure that the secret has the required tenants file in the "tenants.yaml" key.
// Takes in string map, in case secret needs to have multiple fields or has to be substituted.
func WithTenantsSecret(tenants map[string]string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.tenantsSecret = tenants
	}
}

// WithLogLevel overrides the default log level of Observatorium API.
func WithLogLevel(logLevel string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.logLevel = logLevel
	}
}

// WithRateLimiter emables rate limiting by passing gRPC address arg as flag to Observatorium API.
func WithRateLimiter(rateLimiter string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.rateLimiter = rateLimiter
	}
}

// WithMetrics enables metrics signal for Observatorium API and includes passed in URL args as flags.
// If any of the URLs are empty, the flags are not included.
func WithMetrics(m MetricsBackend) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.metrics = m
	}
}

// WithLogs enables logging signal for Observatorium API and includes passed in URL args as flags.
// If any of the URLs are empty, the flags are not included.
func WithLogs(l LogsBackend) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.logs = l
	}
}

// WithTraces enables tracing signal for Observatorium API and includes passed in URL args as flags.
// If any of the URLs are empty, the flags are not included.
func WithTraces(t TracesBackend) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.traces = t
	}
}

// Default metadata labels.
var DefaultLabels map[string]string = map[string]string{
	k8sutil.ComponentLabel: "api",
	k8sutil.InstanceLabel:  "observatorium",
	k8sutil.NameLabel:      "observatorium-api",
	k8sutil.PartOfLabel:    "observatorium",
}

// NewObservatoriumAPI returns a new instance of Observatorium API, customized with options.
// Also includes options for adding sidecars.
// Returns the following K8s Objects:
// Deployment
// Service
// ServiceAccount
// ConfigMap
// Secret
//
// NOTE: You need to call K8sConfig() to customize k8s-native options.
func NewObservatoriumAPI(opts ...ObservatoriumAPIOption) *observatoriumAPI {
	c := &observatoriumAPI{
		logLevel: "info",
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:           "quay.io/observatorium/api",
			ImageTag:        "latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Name:            "observatorium-api",
			Namespace:       "observatorium",
			CommonLabels:    DefaultLabels,
			Replicas:        3,
			DeploymentStrategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &[]intstr.IntOrString{intstr.FromInt(0)}[0],
					MaxUnavailable: &[]intstr.IntOrString{intstr.FromInt(1)}[0],
				},
			},
			PodResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("3"),
					corev1.ResourceMemory: resource.MustParse("3000Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("2000Mi"),
				},
			},
			Affinity: corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							Weight: 100,
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      k8sutil.NameLabel,
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{"observatorium-api"},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
			},
			LivenessProbe: corev1.Probe{
				FailureThreshold: 10,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/live",
						Port:   intstr.FromInt(8080),
						Scheme: corev1.URISchemeHTTP,
					},
				},
				PeriodSeconds: 30,
			},
			ReadinessProbe: corev1.Probe{
				FailureThreshold: 12,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/ready",
						Port:   intstr.FromInt(8081),
						Scheme: corev1.URISchemeHTTP,
					},
				},
				PeriodSeconds: 5,
			},
			SecurityContext: corev1.PodSecurityContext{
				RunAsUser: &[]int64{65534}[0],
				FSGroup:   &[]int64{65534}[0],
			},
			TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		},
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// K8sConfig overrides the default K8s options for Observatorium API.
func (c *observatoriumAPI) K8sConfig(opts ...k8sutil.DeploymentOption) *observatoriumAPI {
	for _, o := range opts {
		o(&c.DeploymentGenericConfig)
	}

	c.CommonLabels[k8sutil.InstanceLabel] = c.Name
	c.Name = c.Name + "-" + c.CommonLabels[k8sutil.NameLabel]

	return c
}

// Manifests creates resultant K8s YAML manifests for customized Observatorium API deployment
// as a map, containing manifest name and runtime.Object.
// This provides the ability to override options of the returned manifests
// in a specific manner (in case the exported ObservatoriumAPIOption functions do not suffice).
func (c *observatoriumAPI) Manifests(opts ...ObservatoriumAPIOption) k8sutil.ObjectMap {
	if c.Image == "" || c.logLevel == "" || c.Name == "" {
		mimic.Panicf("required params missing")
	}

	if _, ok := c.tenantsSecret["tenants.yaml"]; !ok {
		mimic.Panicf("tenant secret does not have tenants.yaml key")
	}

	podLabelSelectors := make(map[string]string)
	for k, v := range c.CommonLabels {
		podLabelSelectors[k] = v
	}

	podLabelSelectors[k8sutil.VersionLabel] = c.ImageTag

	commonObjectMeta := metav1.ObjectMeta{
		Name:      c.Name,
		Labels:    c.CommonLabels,
		Namespace: c.Namespace,
	}

	// Instantiate service account.
	apiServiceAccount := corev1.ServiceAccount{
		TypeMeta:   k8sutil.ServiceAccountMeta,
		ObjectMeta: commonObjectMeta,
	}

	// Instantiate RBAC ConfigMap.
	rbacConfigMap := corev1.ConfigMap{
		TypeMeta:   k8sutil.ConfigMapMeta,
		ObjectMeta: commonObjectMeta,
		Data: map[string]string{
			"rbac.yaml": c.rbacYAML,
		},
	}

	// Instantiate tenants file secret.
	tenantsSecret := corev1.Secret{
		TypeMeta:   k8sutil.SecretMeta,
		ObjectMeta: commonObjectMeta,
		StringData: c.tenantsSecret,
	}

	apiArgs := k8sutil.ArgList([]string{
		k8sutil.FlagArg("web.listen", "0.0.0.0:8080"),
		k8sutil.FlagArg("web.internal.listen", "0.0.0.0:8081"),
		k8sutil.FlagArg("grpc.listen", "0.0.0.0:8090"),
		k8sutil.FlagArg("internal.tracing.endpoint", "0.0.0.0:6831"),
		k8sutil.FlagArg("log.level", c.logLevel),
		k8sutil.FlagArg("rbac.config", "/etc/observatorium/rbac.yaml"),
		k8sutil.FlagArg("tenants.config", "/etc/observatorium/tenants.yaml"),

		k8sutil.FlagArg("middleware.rate-limiter.grpc-address", c.rateLimiter),

		k8sutil.FlagArg("metrics.read.endpoint", c.metrics.ReadEndpoint),
		k8sutil.FlagArg("metrics.write.endpoint", c.metrics.WriteEndpoint),
		k8sutil.FlagArg("metrics.rules.endpoint", c.metrics.RulesEndpoint),

		k8sutil.FlagArg("logs.read.endpoint", c.logs.ReadEndpoint),
		k8sutil.FlagArg("logs.tail.endpoint", c.logs.TailEndpoint),
		k8sutil.FlagArg("logs.write.endpoint", c.logs.WriteEndpoint),
		k8sutil.FlagArg("logs.rules.endpoint", c.logs.RulesEndpoint),

		k8sutil.FlagArg("traces.write.endpoint", c.traces.WriteEndpoint),
		k8sutil.FlagArg("experimental.traces.read.endpoint-template", c.traces.ReadEndpoint),
	})

	apiArgs = append(apiArgs, c.additionalAPIArgs...)

	// Instantiate Observatorium API container.
	observatoriumAPIContainer := corev1.Container{
		Name:            "observatorium-api",
		Args:            apiArgs,
		Image:           fmt.Sprintf("%s:%s", c.Image, c.ImageTag),
		ImagePullPolicy: c.ImagePullPolicy,
		Ports: []corev1.ContainerPort{
			{
				Name:          "grpc-public",
				ContainerPort: 8090,
			},
			{
				Name:          "internal",
				ContainerPort: 8081,
			},
			{
				Name:          "public",
				ContainerPort: 8080,
			},
		},
		Resources:      c.PodResources,
		LivenessProbe:  &c.LivenessProbe,
		ReadinessProbe: &c.ReadinessProbe,
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/etc/observatorium/rbac.yaml",
				Name:      "rbac",
				ReadOnly:  true,
				SubPath:   "rbac.yaml",
			},
			{
				MountPath: "/etc/observatorium/tenants.yaml",
				Name:      "tenants",
				ReadOnly:  true,
				SubPath:   "tenants.yaml",
			},
		},
	}

	// Attach any configured sidecars.
	containers := []corev1.Container{observatoriumAPIContainer}
	containers = append(containers, c.Extras.Sidecars...)
	// Attach any configured volumes.
	volumes := []corev1.Volume{
		{
			Name: "rbac",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rbacConfigMap.Name,
					},
				},
			},
		},
		{
			Name: "tenants",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: tenantsSecret.Name,
				},
			},
		},
	}
	volumes = append(volumes, c.Extras.AdditionalPodVolumes...)

	apiDeployment := appsv1.Deployment{
		TypeMeta:   k8sutil.DeploymentMeta,
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{c.Replicas}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabelSelectors,
			},
			Strategy: c.DeploymentStrategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    c.CommonLabels,
					Namespace: c.Namespace,
				},
				Spec: corev1.PodSpec{
					SecurityContext:    &c.SecurityContext,
					Affinity:           &c.Affinity,
					Containers:         containers,
					ServiceAccountName: apiServiceAccount.Name,
					Volumes:            volumes,
				},
			},
		},
	}

	// Attach any configured ports.
	ports := []corev1.ServicePort{
		{
			AppProtocol: &[]string{"h2c"}[0],
			Name:        "grpc-public",
			Port:        8090,
			TargetPort:  intstr.FromInt(8090),
		},
		{
			AppProtocol: &[]string{"http"}[0],
			Name:        "internal",
			Port:        8081,
			TargetPort:  intstr.FromInt(8081),
		},
		{
			AppProtocol: &[]string{"http"}[0],
			Name:        "public",
			Port:        8080,
			TargetPort:  intstr.FromInt(8080),
		},
	}
	ports = append(ports, c.Extras.AdditionalServicePorts...)

	// Instantiate API Service.
	apiService := corev1.Service{
		TypeMeta:   k8sutil.ServiceMeta,
		ObjectMeta: commonObjectMeta,
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: podLabelSelectors,
		},
	}

	manifests := k8sutil.ObjectMap{
		"api-serviceAccount": &apiServiceAccount,
		"api-deployment":     &apiDeployment,
		"api-service":        &apiService,
		"api-configMap":      &rbacConfigMap,
		"api-secret":         &tenantsSecret,
	}

	// If enabled, instantiate API Service.
	if c.EnableServiceMonitor {
		endpoints := []monv1.Endpoint{
			{
				Port: "internal",
			},
		}
		endpoints = append(endpoints, c.Extras.AdditionalServiceMonitorPorts...)

		apiServiceMonitor := monv1.ServiceMonitor{
			TypeMeta: k8sutil.ServiceMonitorMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.Name,
				Namespace: c.Namespace,
				Labels: map[string]string{
					"name": "observatorium-api",
				},
			},
			Spec: monv1.ServiceMonitorSpec{
				Endpoints: endpoints,
				NamespaceSelector: monv1.NamespaceSelector{
					MatchNames: []string{c.Namespace},
				},
				Selector: metav1.LabelSelector{
					MatchLabels: c.CommonLabels,
				},
			},
		}

		manifests["api-serviceMonitor"] = &apiServiceMonitor
	}

	return manifests
}
