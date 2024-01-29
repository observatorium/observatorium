package main

import (
	"time"

	"github.com/bwplotka/mimic"

	obsrbac "github.com/observatorium/api/rbac"
	"github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/observatorium/api"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/kubeyaml"
	"github.com/observatorium/observatorium/configuration_go/kubegen/openshift"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	templatev1 "github.com/openshift/api/template/v1"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	rbac := &api.RBAC{
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

	// Create tenants. Ideally, this should come from secret management tool like vault.
	tenants := &api.Tenants{
		Tenants: []api.Tenant{
			{
				Name: "test-oidc",
				ID:   "1610b0c3-c509-4592-a256-a1871353dbfa",
				OIDC: &api.TenantOIDC{
					ClientID:  "observatorium",
					IssuerURL: "http://hydra.hydra.svc.cluster.local:4444/",
				},
				RateLimits: []api.TenantRateLimits{
					{
						Endpoint: "/api/metrics/v1/.+/api/v1/receive",
						Limit:    1000,
						Window:   time.Duration(time.Second),
					},
					{
						Endpoint: "/api/logs/v1/.*",
						Limit:    1000,
						Window:   time.Duration(time.Second),
					},
				},
			},
		},
	}

	// Example 1.
	// Observatorium API with no sidecar and a serviceMonitor.

	// Configure Observatorium API application.
	apiOpts := &api.ObservatoriumAPIOptions{
		LogLevel: log.LevelDebug,
		// Metrics endpoints.
		MetricsReadEndpoint:  "http://observatorium-xyz-thanos-query-frontend.observatorium.svc.cluster.local:9090",
		MetricsWriteEndpoint: "http://observatorium-xyz-thanos-receive.observatorium.svc.cluster.local:19291",
		MetricsRulesEndpoint: "http://observatorium-xyz-rules-objstore.observatorium.svc.cluster.local:8080",
		// Logs endpoints.
		LogsReadEndpoint:  "http://observatorium-xyz-loki-query-frontend-http.observatorium.svc.cluster.local:3100",
		LogsWriteEndpoint: "http://observatorium-xyz-loki-distributor-http.observatorium.svc.cluster.local:3100",
		LogsRulesEndpoint: "http://observatorium-xyz-loki-ruler-http.observatorium.svc.cluster.local:3100",
		LogsTailEndpoint:  "http://observatorium-xyz-loki-querier-http.observatorium.svc.cluster.local:3100",
		// Traces endpoints.
		TracesReadEndpoint:  "http://observatorium-xyz-jaeger-query.observatorium.svc.cluster.local:16686/",
		TracesWriteEndpoint: "observatorium-xyz-otel-collector:4317",
		// Rate limiter.
		MiddlewareRateLimiterGrpcAddress: "observatorium-xyz-gubernator.observatorium.svc.cluster.local:8081",
		// RBAC and tenants.
		RbacConfig:    api.NewRbacConfig(rbac).WithResourceName("observatorium-xyz-rbac"),
		TenantsConfig: api.NewTenantsConfig(tenants).WithResourceName("observatorium-xyz-tenants"),
	}

	// Configure Kubernetes resources.
	apiK8s := api.NewObservatoriumAPI(apiOpts, "observatorium", "latest")
	apiK8s.Name = "observatorium-xyz"
	apiK8s.Replicas = 3
	apiK8s.ContainerResources = kghelpers.NewResourcesRequirements("2", "3", "2Gi", "3Gi")

	// Generate manifests.
	kubeyaml.GenerateWithMimic(g, apiK8s.Objects(), "config-new")

	// Example 2
	// Create sidecar container.
	// It uses k8sutil Container provider that encapsulates all resources needed for a container.
	// Including configMaps, volumes, servicePorts, etc.
	dummyContainer := &workload.Container{
		Name:            "dummy-sidecar",
		Image:           "docker.io/dummy",
		ImageTag:        "latest",
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
		ServicePorts: []corev1.ServicePort{
			{
				Name:       "dummy",
				Port:       7080,
				TargetPort: intstr.FromInt(7080),
			},
		},
		Volumes: []corev1.Volume{kghelpers.NewPodVolumeFromConfigMap("config", "dummy-sidecar-config-new")},
		MonitorPorts: []monv1.Endpoint{
			{
				Port: "dummy-metrics",
			},
		},
		ConfigMaps: map[string]map[string]string{
			"dummy-sidecar-config": {
				"config.yaml": "dummy: true",
			},
		},
	}

	// Reusing the same API config as above for simplicity.
	// And adding the dummy sidecar.
	apiK8s.Sidecars = append(apiK8s.Sidecars, dummyContainer)

	// Generate manifests.
	kubeyaml.GenerateWithMimic(g, apiK8s.Objects(), "config-w-sidecar")

	// Example 3
	// Observatorium API with no sidecar, packaged as Observatorium template.

	apiOpts = &api.ObservatoriumAPIOptions{
		// Metrics endpoints.
		MetricsReadEndpoint:  "http://observatorium-xyz-thanos-query-frontend.${NAMESPACE}.svc.cluster.local:9090",
		MetricsWriteEndpoint: "http://observatorium-xyz-thanos-receive.${NAMESPACE}.svc.cluster.local:19291",
		MetricsRulesEndpoint: "http://observatorium-xyz-rules-objstore.${NAMESPACE}.svc.cluster.local:8080",
		// Logs endpoints.
		LogsReadEndpoint:  "http://observatorium-xyz-loki-query-frontend-http.${NAMESPACE}.svc.cluster.local:3100",
		LogsWriteEndpoint: "http://observatorium-xyz-loki-distributor-http.${NAMESPACE}.svc.cluster.local:3100",
		LogsRulesEndpoint: "http://observatorium-xyz-loki-ruler-http.${NAMESPACE}.svc.cluster.local:3100",
		LogsTailEndpoint:  "http://observatorium-xyz-loki-querier-http.${NAMESPACE}.svc.cluster.local:3100",
		// Traces endpoints.
		TracesReadEndpoint:  "http://observatorium-xyz-jaeger-query.${NAMESPACE}.svc.cluster.local:16686/",
		TracesWriteEndpoint: "observatorium-xyz-otel-collector:4317",
		// Rate limiter.
		MiddlewareRateLimiterGrpcAddress: "observatorium-xyz-gubernator.${NAMESPACE}.svc.cluster.local:8081",
		// RBAC and tenants.
		RbacConfig:    api.NewRbacConfig(rbac),
		TenantsConfig: api.NewTenantsConfig(tenants),
	}
	apiOpts.AddExtraOpts("--log-level=${OBSERVATORIUM_API_LOG_LEVEL}")

	// Configure Kubernetes resources.
	apiK8s = api.NewObservatoriumAPI(apiOpts, "${NAMESPACE}", "${OBSERVATORIUM_API_IMAGE_TAG}")
	apiK8s.Replicas = 3
	apiK8s.ContainerResources = kghelpers.NewResourcesRequirements("2", "3", "2Gi", "3Gi")
	apiK8s.Image = "${OBSERVATORIUM_API_IMAGE}"

	// Generate manifests.
	objects := []runtime.Object{
		openshift.WrapInTemplate(apiK8s.Objects(), metav1.ObjectMeta{
			Name: "observatorium",
		}, []templatev1.Parameter{
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
		}),
	}
	kubeyaml.GenerateWithMimic(g, objects, "openshift-config-new")
}
