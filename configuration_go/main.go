package main

import (
	"fmt"
	"time"

	"github.com/bwplotka/mimic"
	"github.com/ghodss/yaml"
	api "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/observatoriumapi"
	"github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/thanos/compactor"
	query "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/thanos/query"
	"github.com/observatorium/observatorium/configuration_go/generator"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	k8sutilv2 "github.com/observatorium/observatorium/configuration_go/k8sutil/v2"
	"github.com/observatorium/observatorium/configuration_go/openshift"
	apiprovider "github.com/observatorium/observatorium/configuration_go/providers/api"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	jaeger "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/jaeger"

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
			api.WithMetrics(api.MetricsBackend{
				ReadEndpoint:  "http://observatorium-xyz-thanos-query-frontend.observatorium.svc.cluster.local:9090",
				WriteEndpoint: "http://observatorium-xyz-thanos-receive.observatorium.svc.cluster.local:19291",
				RulesEndpoint: "http://observatorium-xyz-rules-objstore.observatorium.svc.cluster.local:8080",
			}),
			api.WithLogs(api.LogsBackend{
				ReadEndpoint:  "http://observatorium-xyz-loki-query-frontend-http.observatorium.svc.cluster.local:3100",
				WriteEndpoint: "http://observatorium-xyz-loki-distributor-http.observatorium.svc.cluster.local:3100",
				RulesEndpoint: "http://observatorium-xyz-loki-ruler-http.observatorium.svc.cluster.local:3100",
				TailEndpoint:  "http://observatorium-xyz-loki-querier-http.observatorium.svc.cluster.local:3100",
			}),
			api.WithTraces(api.TracesBackend{
				ReadEndpoint:  "http://observatorium-xyz-jaeger-query.observatorium.svc.cluster.local:16686/",
				WriteEndpoint: "observatorium-xyz-otel-collector:4317",
			}),
			api.WithRateLimiter("observatorium-xyz-gubernator.observatorium.svc.cluster.local:8081"),
			api.WithRBACYAML(string(rbacData)),
			api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),
		).K8sConfig(
			k8sutil.WithImage("quay.io/observatorium/api", "main-2023-01-24-v0.1.2-318-g5f4fdf4"),
			k8sutil.WithName("observatorium-xyz"),
			k8sutil.WithNamespace("observatorium"),
			k8sutil.WithReplicas(3),
			k8sutil.WithResources(apiResources),
			k8sutil.WithServiceMonitor(),
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

	sidecar := k8sutil.ExtraConfig{
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
			api.WithMetrics(api.MetricsBackend{
				ReadEndpoint:  "http://observatorium-xyz-thanos-query-frontend.observatorium.svc.cluster.local:9090",
				WriteEndpoint: "http://observatorium-xyz-thanos-receive.observatorium.svc.cluster.local:19291",
				RulesEndpoint: "http://observatorium-xyz-rules-objstore.observatorium.svc.cluster.local:8080",
			}),
			api.WithLogs(api.LogsBackend{
				ReadEndpoint:  "http://observatorium-xyz-loki-query-frontend-http.observatorium.svc.cluster.local:3100",
				WriteEndpoint: "http://observatorium-xyz-loki-distributor-http.observatorium.svc.cluster.local:3100",
				RulesEndpoint: "http://observatorium-xyz-loki-ruler-http.observatorium.svc.cluster.local:3100",
				TailEndpoint:  "http://observatorium-xyz-loki-querier-http.observatorium.svc.cluster.local:3100",
			}),
			api.WithTraces(api.TracesBackend{
				ReadEndpoint:  "http://observatorium-xyz-jaeger-query.observatorium.svc.cluster.local:16686/",
				WriteEndpoint: "observatorium-xyz-otel-collector:4317",
			}),
			api.WithRateLimiter("observatorium-xyz-gubernator.observatorium.svc.cluster.local:8081"),
			api.WithRBACYAML(string(rbacData)),
			api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),
		).K8sConfig(
			k8sutil.WithImage("quay.io/observatorium/api", "main-2023-01-24-v0.1.2-318-g5f4fdf4"),
			k8sutil.WithName("observatorium-xyz"),
			k8sutil.WithNamespace("observatorium"),
			k8sutil.WithReplicas(3),
			k8sutil.WithResources(apiResources),
			k8sutil.WithServiceMonitor(),
			// Add dummy-sidecar stuff
			k8sutil.WithExtras(sidecar),
		).Manifests(),
		"config-w-sidecar",
	)

	// Thanos Query sample deployment.
	generator.GenerateWithMimic(
		g,
		query.NewThanosQuery(
			query.WithGRPCOptions(query.GRPCOptions{
				ServerAddress: "0.0.0.0:10901",
				ProxyStrategy: "eager",
			}),
			query.WithHTTPOptions(query.HTTPOptions{
				BindAddress: "0.0.0.0:9090",
			}),
			query.WithLogging("debug", "logfmt"),
			query.WithReplicaLabels([]string{"replica", "rule_replica", "prometheus_replica"}),
			query.WithWebOptions(query.WebOptions{
				PrefixHeaderName: "X-Forwarded-Prefix",
			}),
			query.WithEndpoints([]string{
				"dnssrv+_grpc._tcp.observatorium-thanos-store-shard-0.observatorium.svc.cluster.local",
				"dnssrv+_grpc._tcp.observatorium-thanos-store-shard-1.observatorium.svc.cluster.local",
				"dnssrv+_grpc._tcp.observatorium-thanos-store-shard-2.observatorium.svc.cluster.local",
				"dnssrv+_grpc._tcp.observatorium-thanos-receive.observatorium.svc.cluster.local",
			}...),
			query.WithQueryTimeout("15m"),
			query.WithLookbackDelta("15m"),
			query.WithAutoDownsampling(),
			query.WithMaxConcurrentQueries(20),
			query.WithEngine("prometheus"),
			query.WithTracingConfig(trclient.TracingConfig{
				Type: trclient.Jaeger,
				Config: jaeger.Config{
					SamplerParam: 2,
					SamplerType:  jaeger.SamplerTypeRateLimiting,
					ServiceName:  "thanos-query",
				},
			}),
			query.WithStoreOptions(query.StoreOptions{
				SDFiles: []query.SDFile{{
					Name: "rule-sd",
					Data: "dnssrv+_grpc._tcp.observatorium-thanos-rule.observatorium.svc.cluster.local",
				}},
			}),
			query.WithQueryTelemetry(query.QueryTelemetry{
				DurationQuantiles: []float64{0.1, 0.25, 0.75, 1.25, 1.75, 2, 3, 5, 10, 15, 30, 60, 120},
			}),
		).K8sConfig(
			k8sutil.WithImage("quay.io/thanos-io/thanos", "v0.31"),
			k8sutil.WithName("observatorium-xyz"),
			k8sutil.WithNamespace("observatorium"),
			k8sutil.WithReplicas(3),
			k8sutil.WithResources(apiResources),
			k8sutil.WithServiceMonitor(),
			// Add dummy-sidecar stuff
			k8sutil.WithExtras(sidecar),
		).Manifests(),
		"config-w-sidecar",
	)

	// Example
	// Thanos Compactor deployment as statefulSet with sidecar.

	// Set the compactor options.
	compactorSatefulset := compactor.NewCompactor()
	compactorSatefulset.Options.LogLevel = "debug"
	compactorSatefulset.Options.AddExtraOpts("--debug.accept-malformed-index")

	// Set the kube config.
	compactorSatefulset.ImageTag = "v0.32"
	compactorSatefulset.Env = append(compactorSatefulset.Env, k8sutilv2.NewEnvFromSecret("AWS_ACCESS_KEY_ID", "aws_access_key_id", "name"))
	compactorSatefulset.Env = append(compactorSatefulset.Env, k8sutilv2.NewEnvFromSecret("AWS_SECRET_ACCESS_KEY", "aws_secret_access_key", "name"))

	// Add sidecar.
	compactorSatefulset.Sidecars = append(compactorSatefulset.Sidecars, makeOauthProxyContainer(8443, 10902, compactorSatefulset.Name))

	// Add configmap.
	compactorSatefulset.ConfigMaps["objStore"] = map[string]string{
		"data": "dummy",
	}

	// Generate manifests and do some post processing.
	compactorManifests := compactorSatefulset.Manifests()
	compactorManifests["compactor-serviceMonitor"].(*monv1.ServiceMonitor).ObjectMeta.Namespace = "openshift-monitoring"
	compactorManifests["compactor-configMap-objStore"].(*corev1.ConfigMap).ObjectMeta.Annotations["tls"] = "platform-tls"
	generator.GenerateWithMimic(g, compactorManifests, "compactor")

	// Derive base configuration from the statefulset for another deployment.
	compactorSatefulset.Name = "compactor-tenant-a"
	compactorSatefulset.Options.RetentionResolutionRaw = time.Hour * 24 * 365
	compactorSatefulset.Options.DeleteExtraOpts()
	compactorSatefulset.ConfigMaps["objStore"] = map[string]string{
		"data": "tenant-a",
	}
	generator.GenerateWithMimic(g, compactorSatefulset.Manifests(), "compactor-tenant-a")

	// Example 3
	// Observatorium API with no sidecar, packaged as Observatorium template.
	generator.GenerateWithMimic(
		g,
		openshift.WrapInTemplate(
			"observatorium-template",
			api.NewObservatoriumAPI(
				api.WithLogLevel("${OBSERVATORIUM_API_LOG_LEVEL}"),
				api.WithMetrics(api.MetricsBackend{
					ReadEndpoint:  "http://observatorium-xyz-thanos-query-frontend.${NAMESPACE}.svc.cluster.local:9090",
					WriteEndpoint: "http://observatorium-xyz-thanos-receive.${NAMESPACE}.svc.cluster.local:19291",
					RulesEndpoint: "http://observatorium-xyz-rules-objstore.${NAMESPACE}.svc.cluster.local:8080",
				}),
				api.WithLogs(api.LogsBackend{
					ReadEndpoint:  "http://observatorium-xyz-loki-query-frontend-http.${NAMESPACE}.svc.cluster.local:3100",
					WriteEndpoint: "http://observatorium-xyz-loki-distributor-http.${NAMESPACE}.svc.cluster.local:3100",
					RulesEndpoint: "http://observatorium-xyz-loki-ruler-http.${NAMESPACE}.svc.cluster.local:3100",
					TailEndpoint:  "http://observatorium-xyz-loki-querier-http.${NAMESPACE}.svc.cluster.local:3100",
				}),
				api.WithTraces(api.TracesBackend{
					ReadEndpoint:  "http://observatorium-xyz-jaeger-query.${NAMESPACE}.svc.cluster.local:16686/",
					WriteEndpoint: "observatorium-xyz-otel-collector:4317",
				}),
				api.WithRateLimiter("observatorium-xyz-gubernator.${NAMESPACE}.svc.cluster.local:8081"),
				api.WithRBACYAML(string(rbacData)),
				api.WithTenantsSecret(map[string]string{"tenants.yaml": string(tenantsData)}),
			).K8sConfig(
				k8sutil.WithImage("${OBSERVATORIUM_API_IMAGE}", "${OBSERVATORIUM_API_IMAGE_TAG}"),
				k8sutil.WithName("observatorium-xyz"),
				k8sutil.WithNamespace("${NAMESPACE}"),
				k8sutil.WithResources(apiResources),
				k8sutil.WithReplicas(3),
				k8sutil.WithServiceMonitor(),
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

func makeOauthProxyContainer(httpsPort, upstreamPort int, serviceAccount string) *k8sutilv2.Container {
	return &k8sutilv2.Container{
		Name:            "oauth-proxy",
		Image:           "quay.io/oauth-proxy",
		ImageTag:        "4.13.0",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			fmt.Sprintf("-https-address=:%d", httpsPort),
			fmt.Sprintf("-upstream=http://localhost:%d", upstreamPort),
			fmt.Sprintf("-openshift-service-account=%s", serviceAccount),
			"-tls-cert=/etc/tls/private/tls.crt",
			"-tls-key=/etc/tls/private/tls.key",
			"-client-secret-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
			"-cookie-secret-file=/etc/proxy/secrets/session_secret",
			"-openshift-ca=/etc/pki/tls/cert.pem",
			"-openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: int32(httpsPort),
			},
		},
		Resources: k8sutilv2.NewResourcesRequirements("25m", "50m", "200Mi", "200Mi"),
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/etc/tls/private",
				Name:      "compact-tls",
				ReadOnly:  false,
			},
			{
				MountPath: "/etc/proxy/secrets",
				Name:      "compact-proxy",
				ReadOnly:  false,
			},
		},
		Volumes: []corev1.Volume{
			k8sutilv2.NewPodVolumeFromSecret("compact-tls", "compact-tls"),
			k8sutilv2.NewPodVolumeFromSecret("compact-proxy", "compact-proxy"),
		},
		ServicePorts: []corev1.ServicePort{
			k8sutilv2.NewServicePort("https", httpsPort, httpsPort),
		},
	}
}
