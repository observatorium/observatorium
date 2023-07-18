package query

import (
	"fmt"

	"github.com/bwplotka/mimic"
	"github.com/ghodss/yaml"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type query struct {
	logLevel, logFormat string
	g                   GRPCOptions
	h                   HTTPOptions
	w                   WebOptions

	activeQueryPath                   string
	autoDownsampling                  bool
	connectionMetricLabels            []string
	selectorLabels                    labels.Labels
	defaultEvaluationInterval         model.Duration
	defaultStep                       model.Duration
	lookbackDelta                     model.Duration
	dynamicLookbackDelta              bool
	maxConcurrentQueries              int
	maxConcurrentSelects              int
	instantDefaultMaxSourceResolution model.Duration
	metadataDefaultTimeRange          model.Duration
	partialResponse                   bool
	promQLEngine                      string
	promQLQueryMode                   string
	replicaLabels                     []string
	telemetry                         QueryTelemetry
	timeout                           model.Duration
	tracingConfig                     string
	alertQueryURL                     string

	endpoints           []string
	endpointsStrict     []string
	endpointGroup       []string
	endpointGroupStrict []string

	s StoreOptions

	additionalQueryArgs []string

	k8sutil.GenericDeploymentConfig
}

type StoreOptions struct {
	RequestSampleLimit int
	RequestSeriesLimit int

	ResponseTimeout  model.Duration
	UnhealthyTimeout model.Duration

	SDDNSInterval model.Duration
	SDDNSResolver string
	SDInterval    model.Duration
	SDFiles       []SDFile
}

type SDFile struct {
	Data string
	Name string
}

type GRPCOptions struct {
	ClientSecure     bool
	ClientSkipVerify bool
	ClientCert       string
	ClientKey        string
	ClientCACert     string
	ClientServerName string
	Compression      string

	ServerAddress          string
	ServerTLSCert          string
	ServerTLSKey           string
	ServerTLSClientCA      string
	ServerMaxConnectionAge model.Duration
	ServerGracePeriod      model.Duration

	ProxyStrategy string
}

type HTTPOptions struct {
	BindAddress string
	GracePeriod model.Duration
	TLSConfig   string
}

type WebOptions struct {
	DisableCORS      bool
	RoutePrefix      string
	ExternalPrefix   string
	PrefixHeaderName string
}

type QueryTelemetry struct {
	DurationQuantiles []float64
	SampleQuantiles   []float64
	SeriesQuantiles   []float64
}

// Allows specifying external functions for overriding Thanos Querier options.
type ThanosQueryOption func(q *query)

// WithLogging overrides the default log level & format of Thanos Querier.
func WithLogging(logLevel, logFormat string) ThanosQueryOption {
	return func(q *query) {
		q.logLevel = logLevel
		q.logFormat = logFormat
	}
}

func WithGRPCOptions(opts GRPCOptions) ThanosQueryOption {
	return func(q *query) {
		q.g = opts
	}
}

func WithHTTPOptions(opts HTTPOptions) ThanosQueryOption {
	return func(q *query) {
		q.h = opts
	}
}

func WithWebOptions(opts WebOptions) ThanosQueryOption {
	return func(q *query) {
		q.w = opts
	}
}

func WithStoreOptions(opts StoreOptions) ThanosQueryOption {
	return func(q *query) {
		q.s = opts
	}
}

// WithEndpoints allows passing the needed StoreAPI endpoints to Thanos Querier.
func WithEndpoints(endpoints ...string) ThanosQueryOption {
	return func(q *query) {
		q.endpoints = endpoints
	}
}

// WithEndpointsStrict allows passing the needed StoreAPI endpoints to Thanos Querier
// that are always used even when health check fails.
func WithEndpointsStrict(endpoints ...string) ThanosQueryOption {
	return func(q *query) {
		q.endpointsStrict = endpoints
	}
}

func WithAutoDownsampling() ThanosQueryOption {
	return func(q *query) {
		q.autoDownsampling = true
	}
}

func WithActiveQueryPath(path string) ThanosQueryOption {
	return func(q *query) {
		q.activeQueryPath = path
	}
}

func WithConnectionMetricLabels(labels []string) ThanosQueryOption {
	return func(q *query) {
		q.connectionMetricLabels = labels
	}
}

func WithSelectorLabels(labels labels.Labels) ThanosQueryOption {
	return func(q *query) {
		q.selectorLabels = labels
	}
}

func WithDefaultEvaluationInterval(interval string) ThanosQueryOption {
	return func(q *query) {
		pi, err := model.ParseDuration(interval)
		if err != nil {
			panic(err)
		}
		q.defaultEvaluationInterval = pi
	}
}

func WithDefaultStep(step string) ThanosQueryOption {
	return func(q *query) {
		ps, err := model.ParseDuration(step)
		if err != nil {
			panic(err)
		}
		q.defaultStep = ps
	}
}

func WithLookbackDelta(delta string) ThanosQueryOption {
	return func(q *query) {
		pd, err := model.ParseDuration(delta)
		if err != nil {
			panic(err)
		}
		q.lookbackDelta = pd
	}
}

func WithDynamicLookbackDelta() ThanosQueryOption {
	return func(q *query) {
		q.dynamicLookbackDelta = true
	}
}

func WithMaxConcurrentQueries(max int) ThanosQueryOption {
	return func(q *query) {
		q.maxConcurrentQueries = max
	}
}

func WithMaxConcurrentSelects(max int) ThanosQueryOption {
	return func(q *query) {
		q.maxConcurrentSelects = max
	}
}

func WithInstantDefaultMaxSourceResolution(resolution string) ThanosQueryOption {
	return func(q *query) {
		pr, err := model.ParseDuration(resolution)
		if err != nil {
			panic(err)
		}
		q.instantDefaultMaxSourceResolution = pr
	}
}

func WithMetadataDefaultTimeRange(timeRange string) ThanosQueryOption {
	return func(q *query) {
		pr, err := model.ParseDuration(timeRange)
		if err != nil {
			panic(err)
		}
		q.metadataDefaultTimeRange = pr
	}
}

func WithPartialResponse() ThanosQueryOption {
	return func(q *query) {
		q.partialResponse = true
	}
}

func WithEngine(engine string) ThanosQueryOption {
	return func(q *query) {
		q.promQLEngine = engine
	}
}

func WithQueryMode(mode string) ThanosQueryOption {
	return func(q *query) {
		q.promQLQueryMode = mode
	}
}

func WithReplicaLabels(labels []string) ThanosQueryOption {
	return func(q *query) {
		q.replicaLabels = labels
	}
}

func WithQueryTelemetry(qt QueryTelemetry) ThanosQueryOption {
	return func(q *query) {
		q.telemetry = qt
	}
}

func WithQueryTimeout(timeout string) ThanosQueryOption {
	return func(q *query) {
		pt, err := model.ParseDuration(timeout)
		if err != nil {
			panic(err)
		}
		q.timeout = pt
	}
}

func WithTracingConfig(config trclient.TracingConfig) ThanosQueryOption {
	return func(q *query) {
		b, err := yaml.Marshal(config)
		if err != nil {
			panic(err)
		}
		q.tracingConfig = string(b)
	}
}

func WithAlertQueryURL(url string) ThanosQueryOption {
	return func(q *query) {
		q.alertQueryURL = url
	}
}

// WithAdditionalQueryArgs allows including additional arguments to the Thanos Query deployment.
func WithAdditionalQueryArgs(additionalQueryArgs []string) ThanosQueryOption {
	return func(q *query) {
		q.additionalQueryArgs = additionalQueryArgs
	}
}

// Default metadata labels.
var DefaultLabels map[string]string = map[string]string{
	k8sutil.ComponentLabel: "query-layer",
	k8sutil.InstanceLabel:  "observatorium",
	k8sutil.NameLabel:      "thanos-query",
	k8sutil.PartOfLabel:    "observatorium",
}

func NewThanosQuery(opts ...ThanosQueryOption) *query {
	q := query{
		logLevel:  "info",
		logFormat: "logfmt",
		GenericDeploymentConfig: k8sutil.GenericDeploymentConfig{
			Image:           "quay.io/thanos/thanos",
			ImageTag:        "latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Name:            "thanos-query",
			Namespace:       "observatorium",
			CommonLabels:    DefaultLabels,
			Replicas:        5,
			DeploymentStrategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &[]intstr.IntOrString{intstr.FromInt(0)}[0],
					MaxUnavailable: &[]intstr.IntOrString{intstr.FromInt(1)}[0],
				},
			},
			PodResources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("3"),
					corev1.ResourceMemory: resource.MustParse("5Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("3Gi"),
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
											Values:   []string{"thanos-query"},
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
				FailureThreshold: 4,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/-/healthy",
						Port:   intstr.FromInt(9090),
						Scheme: corev1.URISchemeHTTP,
					},
				},
				PeriodSeconds: 30,
			},
			ReadinessProbe: corev1.Probe{
				FailureThreshold: 20,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/-/ready",
						Port:   intstr.FromInt(9090),
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
		o(&q)
	}

	return &q
}

// K8sConfig overrides the default K8s options for Thanos Query.
func (q *query) K8sConfig(opts ...k8sutil.DeploymentOption) *query {
	for _, o := range opts {
		o(&q.GenericDeploymentConfig)
	}

	q.CommonLabels[k8sutil.InstanceLabel] = q.Name
	q.Name = q.Name + "-" + q.CommonLabels[k8sutil.NameLabel]

	return q
}

// Manifests creates resultant K8s YAML manifests for customized Thanos Querier deployment
// as a map, containing manifest name and runtime.Object.
// This provides the ability to override options of the returned manifests
// in a specific manner (in case the exported ThanosQueryOptions functions do not suffice).
func (q *query) Manifests(opts ...ThanosQueryOption) k8sutil.ObjectMap {
	if q.Image == "" || q.logLevel == "" || q.Name == "" {
		mimic.Panicf("required params missing")
	}

	if len(q.endpoints) == 0 && len(q.endpointsStrict) == 0 {
		mimic.Panicf("no stores configured for querier")
	}

	podLabelSelectors := make(map[string]string)
	for k, v := range q.CommonLabels {
		podLabelSelectors[k] = v
	}

	podLabelSelectors[k8sutil.VersionLabel] = q.ImageTag

	commonObjectMeta := metav1.ObjectMeta{
		Name:      q.Name,
		Labels:    q.CommonLabels,
		Namespace: q.Namespace,
	}

	// Instantiate service account.
	queryServiceAccount := corev1.ServiceAccount{
		TypeMeta:   k8sutil.ServiceAccountMeta,
		ObjectMeta: commonObjectMeta,
	}

	sdFileList := []string{}
	sdData := map[string]string{}
	for _, f := range q.s.SDFiles {
		sdFileList = append(sdFileList, "etc/thanos/sd/"+f.Name+".yaml")
		sdData[f.Name+".yaml"] = f.Data
	}

	sdConfigMap := corev1.ConfigMap{
		TypeMeta:   k8sutil.ConfigMapMeta,
		ObjectMeta: commonObjectMeta,
		Data:       sdData,
	}

	queryArgs := k8sutil.ArgList(
		// Logging.
		k8sutil.FlagArg("log.level", q.logLevel),
		k8sutil.FlagArg("log.format", q.logFormat),

		// HTTP.
		k8sutil.FlagArg("http-address", q.h.BindAddress),
		k8sutil.FlagArg("http-grace-period", q.h.GracePeriod.String()),
		k8sutil.FlagArg("http.config", q.h.TLSConfig),

		// gRPC.
		k8sutil.FlagArg("grpc-address", q.g.ServerAddress),
		k8sutil.FlagArg("grpc-server-tls-cert", q.g.ServerTLSCert),
		k8sutil.FlagArg("grpc-server-tls-key", q.g.ServerTLSKey),
		k8sutil.FlagArg("grpc-server-tls-client-ca", q.g.ServerTLSClientCA),
		k8sutil.FlagArg("grpc-server-max-connection-age", q.g.ServerMaxConnectionAge.String()),
		k8sutil.FlagArg("grpc-grace-period", q.g.ServerGracePeriod.String()),
		k8sutil.BoolFlagArg("grpc-client-tls-secure", q.g.ClientSecure),
		k8sutil.BoolFlagArg("grpc-client-tls-skip-verify", q.g.ClientSkipVerify),
		k8sutil.FlagArg("grpc-client-tls-cert", q.g.ClientCert),
		k8sutil.FlagArg("grpc-client-tls-key", q.g.ClientKey),
		k8sutil.FlagArg("grpc-client-tls-ca", q.g.ClientCACert),
		k8sutil.FlagArg("grpc-client-server-name", q.g.ClientServerName),
		k8sutil.FlagArg("grpc-compression", q.g.Compression),
		k8sutil.FlagArg("grpc.proxy-strategy", q.g.ProxyStrategy),

		// Web.
		k8sutil.BoolFlagArg("web.disable-cors", q.w.DisableCORS),
		k8sutil.FlagArg("web.external-prefix", q.w.ExternalPrefix),
		k8sutil.FlagArg("web.prefix-header", q.w.PrefixHeaderName),
		k8sutil.FlagArg("web.route-prefix", q.w.RoutePrefix),

		// Query.
		k8sutil.FlagArg("query.active-query-path", q.activeQueryPath),
		k8sutil.BoolFlagArg("query.auto-downsampling", q.autoDownsampling),
		k8sutil.RepeatableFlagArg("query.conn-metric.label", q.connectionMetricLabels),
		k8sutil.RepeatableLabelFlagArg("selector-label", q.selectorLabels),
		k8sutil.RepeatableFlagArg("query.replica-label", q.replicaLabels),
		k8sutil.FlagArg("query.default-evaluation-interval", q.defaultEvaluationInterval.String()),
		k8sutil.FlagArg("query.default-step", q.defaultStep.String()),
		k8sutil.FlagArg("query.lookback-delta", q.lookbackDelta.String()),
		k8sutil.BoolFlagArg("query.dynamic-lookback-delta", q.dynamicLookbackDelta),
		k8sutil.FlagArg("query.max-concurrent", fmt.Sprint(q.maxConcurrentQueries)),
		k8sutil.FlagArg("query.max-concurrent-select", fmt.Sprint(q.maxConcurrentSelects)),
		k8sutil.FlagArg("query.instant.default.max_source_resolution", q.instantDefaultMaxSourceResolution.String()),
		k8sutil.FlagArg("query.metadata.default-time-range", q.metadataDefaultTimeRange.String()),
		k8sutil.BoolFlagArg("query.partial-response", q.partialResponse),
		k8sutil.FlagArg("query.promql-engine", string(q.promQLEngine)),
		k8sutil.FlagArg("query.mode", q.promQLQueryMode),
		k8sutil.FlagArg("query.timeout", q.timeout.String()),
		k8sutil.FlagArg("tracing.config", q.tracingConfig),
		k8sutil.FlagArg("alert.query-url", q.alertQueryURL),
		k8sutil.RepeatableFloatFlagArg("query.telemetry.request-duration-seconds-quantiles", q.telemetry.DurationQuantiles),
		k8sutil.RepeatableFloatFlagArg("query.telemetry.request-samples-quantiles", q.telemetry.SampleQuantiles),
		k8sutil.RepeatableFloatFlagArg("query.telemetry.request-series-seconds-quantiles", q.telemetry.SeriesQuantiles),

		// Endpoints.
		k8sutil.RepeatableFlagArg("endpoint", q.endpoints),
		k8sutil.RepeatableFlagArg("endpoint-strict", q.endpointsStrict),
		k8sutil.RepeatableFlagArg("endpoint-group", q.endpointGroup),
		k8sutil.RepeatableFlagArg("endpoint-group-strict", q.endpointGroupStrict),

		// Store.
		k8sutil.FlagArg("store.limits.request-series", fmt.Sprint(q.s.RequestSeriesLimit)),
		k8sutil.FlagArg("store.limits.request-samples", fmt.Sprint(q.s.RequestSampleLimit)),
		k8sutil.FlagArg("store.response-timeout", q.s.ResponseTimeout.String()),
		k8sutil.FlagArg("store.unhealthy-timeout", q.s.UnhealthyTimeout.String()),
		k8sutil.FlagArg("store.sd-dns-interval", q.s.SDDNSInterval.String()),
		k8sutil.FlagArg("store.sd-dns-resolver", q.s.SDDNSResolver),
		k8sutil.FlagArg("store.sd-interval", q.s.SDInterval.String()),
		k8sutil.RepeatableFlagArg("store.sd-files", sdFileList),
	)

	queryArgs = append(queryArgs, q.additionalQueryArgs...)

	// Instantiate Thanos Query container.
	thanosQueryContainer := corev1.Container{
		Name:            "thanos-query",
		Args:            queryArgs,
		Image:           fmt.Sprintf("%s:%s", q.Image, q.ImageTag),
		ImagePullPolicy: q.ImagePullPolicy,
		Ports: []corev1.ContainerPort{
			{
				Name:          "grpc",
				ContainerPort: 10901,
			},
			{
				Name:          "http",
				ContainerPort: 9090,
			},
			{
				Name:          "https",
				ContainerPort: 9091,
			},
		},
		Resources:      q.PodResources,
		LivenessProbe:  &q.LivenessProbe,
		ReadinessProbe: &q.ReadinessProbe,
	}

	thanosQueryContainer.VolumeMounts = append(thanosQueryContainer.VolumeMounts,
		corev1.VolumeMount{
			MountPath: "etc/thanos/sd",
			Name:      "file-sd",
		})

	// Attach any configured sidecars.
	containers := []corev1.Container{thanosQueryContainer}
	containers = append(containers, q.Sidecars.Sidecars...)

	// Attach any configured volumes.
	volumes := []corev1.Volume{
		{
			Name: "file-sd",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sdConfigMap.Name,
					},
				},
			},
		},
	}
	volumes = append(volumes, q.Sidecars.AdditionalPodVolumes...)

	queryDeployment := appsv1.Deployment{
		TypeMeta:   k8sutil.DeploymentMeta,
		ObjectMeta: commonObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{q.Replicas}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabelSelectors,
			},
			Strategy: q.DeploymentStrategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    q.CommonLabels,
					Namespace: q.Namespace,
				},
				Spec: corev1.PodSpec{
					SecurityContext:    &q.SecurityContext,
					Affinity:           &q.Affinity,
					Containers:         containers,
					ServiceAccountName: queryServiceAccount.Name,
					Volumes:            volumes,
				},
			},
		},
	}

	// Attach any configured ports.
	ports := []corev1.ServicePort{
		{
			AppProtocol: &[]string{"h2c"}[0],
			Name:        "grpc",
			Port:        10901,
			TargetPort:  intstr.FromInt(10901),
		},
		{
			AppProtocol: &[]string{"http"}[0],
			Name:        "http",
			Port:        9090,
			TargetPort:  intstr.FromInt(9090),
		},
		{
			AppProtocol: &[]string{"http"}[0],
			Name:        "https",
			Port:        9091,
			TargetPort:  intstr.FromInt(9091),
		},
	}
	ports = append(ports, q.Sidecars.AdditionalServicePorts...)

	// Instantiate Query Service.
	queryService := corev1.Service{
		TypeMeta:   k8sutil.ServiceMeta,
		ObjectMeta: commonObjectMeta,
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: podLabelSelectors,
		},
	}

	manifests := k8sutil.ObjectMap{
		"query-serviceAccount": &queryServiceAccount,
		"query-deployment":     &queryDeployment,
		"query-service":        &queryService,
		"query-sd-configmap":   &sdConfigMap,
	}

	// If enabled, instantiate Query ServiceMonitor.
	if q.EnableServiceMonitor {
		endpoints := []monv1.Endpoint{
			{
				Port: "http",
				RelabelConfigs: []*monv1.RelabelConfig{
					{
						Action:       "replace",
						Separator:    "/",
						SourceLabels: []monv1.LabelName{"namespace", "pod"},
						TargetLabel:  "instance",
					},
				},
			},
		}
		endpoints = append(endpoints, q.Sidecars.AdditionalServiceMonitorPorts...)

		apiServiceMonitor := monv1.ServiceMonitor{
			TypeMeta: k8sutil.ServiceMonitorMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      q.Name,
				Namespace: q.Namespace,
				Labels: map[string]string{
					"name": "thanos-query",
				},
			},
			Spec: monv1.ServiceMonitorSpec{
				Endpoints: endpoints,
				NamespaceSelector: monv1.NamespaceSelector{
					MatchNames: []string{q.Namespace},
				},
				Selector: metav1.LabelSelector{
					MatchLabels: q.CommonLabels,
				},
			},
		}

		manifests["query-serviceMonitor"] = &apiServiceMonitor
	}

	return manifests
}
