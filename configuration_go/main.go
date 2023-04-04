package main

import (
	"time"

	"github.com/bwplotka/mimic"
	"github.com/ghodss/yaml"
	"github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/api"
	"github.com/observatorium/observatorium/configuration_go/generator"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/openshift"
	apiprovider "github.com/observatorium/observatorium/configuration_go/providers/api"

	obsrbac "github.com/observatorium/api/rbac"
	templatev1 "github.com/openshift/api/template/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// This demonstrates 3 examples,
// Generating only API config
// Generating API config with sidecar
// Generating OpenShift template of API config
func main() {
	g := mimic.New().WithTopLevelComment(mimic.GeneratedComment)

	defer g.Generate()

	// Create rbac.
	rbac := apiprovider.RBAC{
		Roles: []obsrbac.Role{
			{
				Name:        "read-write",
				Permissions: []obsrbac.Permission{obsrbac.Read, obsrbac.Write},
				Resources:   []string{"metrics", "logs", "traces"},
				Tenants:     []string{"test-oidc"},
			},
		},
		RoleBindings: []obsrbac.RoleBinding{
			{
				Name:  "test",
				Roles: []string{"read-write"},
				Subjects: []obsrbac.Subject{
					{
						Kind: obsrbac.User,
						Name: "user",
					},
				},
			},
		},
	}

	rbacData, err := yaml.Marshal(rbac)
	mimic.PanicOnErr(err)

	// Create tenants. Ideally, this should come from secret management tool like vault.
	tenants := apiprovider.Tenants{
		Tenants: []*apiprovider.Tenant{
			{
				Name: "test-oidc",
				ID:   "1610b0c3-c509-4592-a256-a1871353dbfa",
				OIDC: &apiprovider.OIDC{
					ClientID:  "observatorium",
					IssuerURL: "http://hydra.hydra.svc.cluster.local:4444/",
				},
				RateLimits: []*apiprovider.Ratelimits{
					{
						Endpoint: "/api/metrics/v1/.+/api/v1/receive",
						Limit:    1000,
						Window:   apiprovider.Duration(time.Second),
					},
					{
						Endpoint: "/api/logs/v1/.*",
						Limit:    1000,
						Window:   apiprovider.Duration(time.Second),
					},
				},
			},
		},
	}

	tenantsData, err := yaml.Marshal(tenants)
	mimic.PanicOnErr(err)

	// Define API pod resources.
	apiResources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("3"),
			corev1.ResourceMemory: resource.MustParse("3000Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("2000Mi"),
		},
	}

	// Example 1.
	// Observatorium API with no sidecar and a serviceMonitor.
	generator.GenerateWithMimic(
		g,
		api.NewObservatoriumAPI(
			api.WithLogLevel("debug"),
			api.WithMetrics(
				"http://observatorium-xyz-thanos-query-frontend.observatorium.svc.cluster.local:9090",
				"http://observatorium-xyz-thanos-receive.observatorium.svc.cluster.local:19291",
				"http://observatorium-xyz-rules-objstore.observatorium.svc.cluster.local:8080",
			),
			api.WithLogs(
				"http://observatorium-xyz-loki-query-frontend-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-distributor-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-ruler-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-querier-http.observatorium.svc.cluster.local:3100",
			),
			api.WithTraces(
				"http://observatorium-xyz-jaeger-query.observatorium.svc.cluster.local:16686/",
				"observatorium-xyz-otel-collector:4317",
			),
			api.WithRateLimiter("observatorium-xyz-gubernator.observatorium.svc.cluster.local:8081"),
			api.WithRBACYAML(string(rbacData)),
			api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),

			api.WithImage("quay.io/observatorium/api", "main-2023-01-24-v0.1.2-318-g5f4fdf4"),
			api.WithName("observatorium-xyz"),
			api.WithNamespace("observatorium"),
			api.WithReplicas(3),
			api.WithAPIResources(apiResources),
			api.WithServiceMonitor(),
		).Manifests(),
		"config",
	)

	// Example 2
	// Create sidecar container.
	dummyContainer := corev1.Container{
		Name:            "dummy-sidecar",
		Image:           "docker.io/dummy:latest",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			"--start-dummy-sidecar",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "dummy-metrics",
				ContainerPort: 7081,
			},
			{
				Name:          "dummy",
				ContainerPort: 7080,
			},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold: 10,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/live",
					Port:   intstr.FromInt(7081),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			PeriodSeconds: 30,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/etc/dummy-sidecar/config.yaml",
				Name:      "config",
				ReadOnly:  true,
				SubPath:   "config.yaml",
			},
		},
	}

	sidecar := k8sutil.SidecarConfig{
		Sidecars: []corev1.Container{dummyContainer},
		AdditionalServicePorts: []corev1.ServicePort{
			{
				Name:       "dummy",
				Port:       7080,
				TargetPort: intstr.FromInt(7080),
			},
		},
		AdditionalPodVolumes: []corev1.Volume{
			{
				Name:         "config",
				VolumeSource: corev1.VolumeSource{},
			},
		},
		AdditionalServiceMonitorPorts: []monv1.Endpoint{
			{
				Port: "dummy-metrics",
			},
		},
	}

	// Observatorium API with a dummy sidecar and a serviceMonitor.
	generator.GenerateWithMimic(
		g,
		api.NewObservatoriumAPI(
			api.WithLogLevel("debug"),
			api.WithMetrics(
				"http://observatorium-xyz-thanos-query-frontend.observatorium.svc.cluster.local:9090",
				"http://observatorium-xyz-thanos-receive.observatorium.svc.cluster.local:19291",
				"http://observatorium-xyz-rules-objstore.observatorium.svc.cluster.local:8080",
			),
			api.WithLogs(
				"http://observatorium-xyz-loki-query-frontend-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-distributor-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-ruler-http.observatorium.svc.cluster.local:3100",
				"http://observatorium-xyz-loki-querier-http.observatorium.svc.cluster.local:3100",
			),
			api.WithTraces(
				"http://observatorium-xyz-jaeger-query.observatorium.svc.cluster.local:16686/",
				"observatorium-xyz-otel-collector:4317",
			),
			api.WithRateLimiter("observatorium-xyz-gubernator.observatorium.svc.cluster.local:8081"),
			api.WithRBACYAML(string(rbacData)),
			api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),

			api.WithImage("quay.io/observatorium/api", "main-2023-01-24-v0.1.2-318-g5f4fdf4"),
			api.WithName("observatorium-xyz"),
			api.WithNamespace("observatorium"),
			api.WithReplicas(3),
			api.WithAPIResources(apiResources),
			api.WithServiceMonitor(),
			// Add dummy-sidecar stuff
			api.WithSidecars(sidecar),
		).Manifests(),
		"config-w-sidecar",
	)

	// Example 3
	// Observatorium API with no sidecar, packaged as Observatorium template.
	generator.GenerateWithMimic(
		g,
		openshift.WrapInTemplate(
			"observatorium-template",
			api.NewObservatoriumAPI(
				api.WithLogLevel("${OBSERVATORIUM_API_LOG_LEVEL}"),
				api.WithMetrics(
					"http://observatorium-xyz-thanos-query-frontend.${NAMESPACE}.svc.cluster.local:9090",
					"http://observatorium-xyz-thanos-receive.${NAMESPACE}.svc.cluster.local:19291",
					"http://observatorium-xyz-rules-objstore.${NAMESPACE}.svc.cluster.local:8080",
				),
				api.WithLogs(
					"http://observatorium-xyz-loki-query-frontend-http.${NAMESPACE}.svc.cluster.local:3100",
					"http://observatorium-xyz-loki-distributor-http.${NAMESPACE}.svc.cluster.local:3100",
					"http://observatorium-xyz-loki-ruler-http.${NAMESPACE}.svc.cluster.local:3100",
					"http://observatorium-xyz-loki-querier-http.${NAMESPACE}.svc.cluster.local:3100",
				),
				api.WithTraces(
					"http://observatorium-xyz-jaeger-query.${NAMESPACE}.svc.cluster.local:16686/",
					"observatorium-xyz-otel-collector:4317",
				),
				api.WithRateLimiter("observatorium-xyz-gubernator.${NAMESPACE}.svc.cluster.local:8081"),
				api.WithRBACYAML(string(rbacData)),
				api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),

				api.WithImage("${OBSERVATORIUM_API_IMAGE}", "${OBSERVATORIUM_API_IMAGE_TAG}"),
				api.WithName("observatorium-xyz"),
				api.WithNamespace("${NAMESPACE}"),
				api.WithAPIResources(apiResources),
				api.WithReplicas(3),
				api.WithServiceMonitor(),
			).Manifests(),
			metav1.ObjectMeta{
				Name: "observatorium",
			},
			[]templatev1.Parameter{
				{
					Name:  "OBSERVATORIUM_API_IMAGE",
					Value: "quay.io/observatorium/api",
				},
				{
					Name:  "OBSERVATORIUM_API_IMAGE_TAG",
					Value: "main-2023-01-24-v0.1.2-318-g5f4fdf4",
				},
				{
					Name:  "NAMESPACE",
					Value: "observatorium",
				},
				{
					Name:  "OBSERVATORIUM_API_LOG_LEVEL",
					Value: "debug",
				},
			},
		),
		"openshift-config",
	)
}
