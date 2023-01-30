package api

import (
	"fmt"

	"github.com/bwplotka/mimic"
	"github.com/go-openapi/swag"
	"github.com/observatorium/observatorium/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type observatoriumAPI struct {
	// API configs.
	logLevel                                 string
	rbacYAML                                 string
	tenantsSecret                            map[string]string
	metricsRead, metricsWrite, metricsRules  string
	logsRead, logsWrite, logsRules, logsTail string
	tracesRead, tracesWrite                  string
	rateLimiter                              string
	additionalAPIArgs                        []string

	// K8s specific configs.
	image                string
	imageTag             string
	imagePullPolicy      corev1.PullPolicy
	name                 string
	namespace            string
	commonLabels         map[string]string
	replicas             int32
	apiPodResources      corev1.ResourceRequirements
	enableServiceMonitor bool

	//Sidecar configs.
	sidecars k8sutil.SidecarConfig
}

// Allows specifying external functions for overriding Observatorium API options.
type ObservatoriumAPIOption func(a *observatoriumAPI)

// WithImage overrides default Observatorium latest image. K8s config.
func WithImage(image, imageTag string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.image = image
		a.imageTag = imageTag
	}
}

// WithImagePullPolicy overrides default (IfNotPresent) image pull policy. K8s config.
func WithImagePullPolicy(imagePullPolicy corev1.PullPolicy) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.imagePullPolicy = imagePullPolicy
	}
}

// WithImage overrides the default number of API replicas to run. Default is 1. K8s config.
func WithReplicas(replicas int32) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.replicas = replicas
	}
}

// WithAdditionalAPIArgs allows including additional arguments to the Observatorium API deployment.
func WithAdditionalAPIArgs(additionalAPIArgs []string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.additionalAPIArgs = additionalAPIArgs
	}
}

// WithCommonLabels overrides the default K8s metadata labels and selectors. K8s config.
func WithCommonLabels(commonLabels map[string]string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.commonLabels = commonLabels
	}
}

// WithName overrides the default name of all the individual objects. K8s config.
func WithName(name string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.name = name
	}
}

// WithNamespace overrides the default namespace of all the individual objects. K8s config.
func WithNamespace(namespace string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.namespace = namespace
	}
}

// WithAPIResources overrides the default API Pod resource config (empty) .
func WithAPIResources(resource corev1.ResourceRequirements) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.apiPodResources = resource
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
func WithMetrics(metricsRead, metricsWrite, metricsRules string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.metricsRead = metricsRead
		a.metricsWrite = metricsWrite
		a.metricsRules = metricsRules
	}
}

// WithLogs enables logging signal for Observatorium API and includes passed in URL args as flags.
// If any of the URLs are empty, the flags are not included.
func WithLogs(logsRead, logsWrite, logsRules, logsTail string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.logsRead = logsRead
		a.logsWrite = logsWrite
		a.logsRules = logsRules
		a.logsTail = logsTail
	}
}

// WithTraces enables tracing signal for Observatorium API and includes passed in URL args as flags.
// If any of the URLs are empty, the flags are not included.
func WithTraces(tracesRead, tracesWrite string) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.tracesRead = tracesRead
		a.tracesWrite = tracesWrite
	}
}

// WithSidecars allows overriding default K8s Deployment and adding additional containers,
// alongside Observatorium API.
// Use SidecarConfig struct to attach additonal volumes, service ports and service monitor
// scrape endpoints for these sidecars.
func WithSidecars(s k8sutil.SidecarConfig) ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.sidecars = s
	}
}

// WithServiceMonitor enables generation of a ServiceMonitor to scrape Observatorium API.
func WithServiceMonitor() ObservatoriumAPIOption {
	return func(a *observatoriumAPI) {
		a.enableServiceMonitor = true
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
func NewObservatoriumAPI(opts ...ObservatoriumAPIOption) *observatoriumAPI {
	c := &observatoriumAPI{
		logLevel:        "info",
		image:           "quay.io/observatorium/api:latest",
		replicas:        1,
		commonLabels:    DefaultLabels,
		namespace:       "observatorium",
		name:            "observatorium-api",
		imagePullPolicy: corev1.PullIfNotPresent,
	}

	for _, o := range opts {
		o(c)
	}

	c.commonLabels[k8sutil.InstanceLabel] = c.name
	c.name = c.name + "-" + c.commonLabels[k8sutil.NameLabel]

	return c
}

// Manifests creates resultant K8s YAML manifests for customized Observatorium API deployment
// as a map, containing manifest name and runtime.Object.
// This provides the ability to override options of the returned manifests
// in a specific manner (in case the exported ObservatoriumAPIOption functions do not suffice).
func (c *observatoriumAPI) Manifests(opts ...ObservatoriumAPIOption) k8sutil.ObjectMap {
	if c.image == "" || c.logLevel == "" || c.name == "" {
		mimic.Panicf("required params missing")
	}

	if _, ok := c.tenantsSecret["tenants.yaml"]; !ok {
		mimic.Panicf("tenant secret does not have tenants.yaml key")
	}

	podLabelSelectors := make(map[string]string)
	for k, v := range c.commonLabels {
		podLabelSelectors[k] = v
	}

	podLabelSelectors[k8sutil.VersionLabel] = c.imageTag

	commonObjectMeta := metav1.ObjectMeta{
		Name:      c.name,
		Labels:    c.commonLabels,
		Namespace: c.namespace,
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

	apiArgs := k8sutil.ArgList(
		k8sutil.FlagArg("web.listen", "0.0.0.0:8080"),
		k8sutil.FlagArg("web.internal.listen", "0.0.0.0:8081"),
		k8sutil.FlagArg("grpc.listen", "0.0.0.0:8090"),
		k8sutil.FlagArg("internal.tracing.endpoint", "0.0.0.0:6831"),
		k8sutil.FlagArg("log.level", c.logLevel),
		k8sutil.FlagArg("rbac.config", "/etc/observatorium/rbac.yaml"),
		k8sutil.FlagArg("tenants.config", "/etc/observatorium/tenants.yaml"),

		k8sutil.FlagArg("middleware.rate-limiter.grpc-address", c.rateLimiter),

		k8sutil.FlagArg("metrics.read.endpoint", c.metricsRead),
		k8sutil.FlagArg("metrics.write.endpoint", c.metricsWrite),
		k8sutil.FlagArg("metrics.rules.endpoint", c.metricsRules),

		k8sutil.FlagArg("logs.read.endpoint", c.logsRead),
		k8sutil.FlagArg("logs.tail.endpoint", c.logsTail),
		k8sutil.FlagArg("logs.write.endpoint", c.logsWrite),
		k8sutil.FlagArg("logs.rules.endpoint", c.logsRules),

		k8sutil.FlagArg("traces.write.endpoint", c.tracesWrite),
		k8sutil.FlagArg("experimental.traces.read.endpoint-template", c.tracesRead),
	)

	apiArgs = append(apiArgs, c.additionalAPIArgs...)

	// Instantiate Observatorium API container.
	observatoriumAPIContainer := corev1.Container{
		Name:            "observatorium-api",
		Args:            apiArgs,
		Image:           fmt.Sprintf("%s:%s", c.image, c.imageTag),
		ImagePullPolicy: c.imagePullPolicy,
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
		Resources: c.apiPodResources,
		LivenessProbe: &corev1.Probe{
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
		ReadinessProbe: &corev1.Probe{
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
	containers = append(containers, c.sidecars.Sidecars...)
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
	volumes = append(volumes, c.sidecars.AdditionalPodVolumes...)

	// Pointers for deployment strategy.
	maxSurge := intstr.FromInt(0)
	maxUnavail := intstr.FromInt(1)
	apiDeployment := appsv1.Deployment{
		TypeMeta:   k8sutil.DeploymentMeta,
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: swag.Int32(c.replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabelSelectors,
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavail,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    c.commonLabels,
					Namespace: c.namespace,
				},
				Spec: corev1.PodSpec{
					Containers:         containers,
					ServiceAccountName: apiServiceAccount.Name,
					Volumes:            volumes,
				},
			},
		},
	}

	// Pointers for app protocol. K8s does not have types for it.
	h2cProtocol := "h2c"
	httpProtocol := "http"

	// Attach any configured ports.
	ports := []corev1.ServicePort{
		{
			AppProtocol: &h2cProtocol,
			Name:        "grpc-public",
			Port:        8090,
			TargetPort:  intstr.FromInt(8090),
		},
		{
			AppProtocol: &httpProtocol,
			Name:        "internal",
			Port:        8081,
			TargetPort:  intstr.FromInt(8081),
		},
		{
			AppProtocol: &httpProtocol,
			Name:        "public",
			Port:        8080,
			TargetPort:  intstr.FromInt(8080),
		},
	}
	ports = append(ports, c.sidecars.AdditionalServicePorts...)

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
	if c.enableServiceMonitor {
		endpoints := []monv1.Endpoint{
			{
				Port: "internal",
			},
		}
		endpoints = append(endpoints, c.sidecars.AdditionalServiceMonitorPorts...)

		apiServiceMonitor := monv1.ServiceMonitor{
			TypeMeta: k8sutil.ServiceMonitorMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.name,
				Namespace: c.namespace,
				Labels: map[string]string{
					"name": "observatorium-api",
				},
			},
			Spec: monv1.ServiceMonitorSpec{
				Endpoints: endpoints,
				NamespaceSelector: monv1.NamespaceSelector{
					MatchNames: []string{c.namespace},
				},
				Selector: metav1.LabelSelector{
					MatchLabels: c.commonLabels,
				},
			},
		}

		manifests["api-serviceMonitor"] = &apiServiceMonitor
	}

	return manifests
}
