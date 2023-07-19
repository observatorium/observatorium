package query

import (
	"fmt"

	"github.com/bwplotka/mimic"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"gopkg.in/yaml.v2"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type query struct {
	// Query configs.
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

	// Embedded K8s config struct which exposes override methods.
	k8sutil.DeploymentGenericConfig
}

// StoreOptions specify the Store API, and Store service discovery options
// for Thanos Querier.
type StoreOptions struct {
	// The maximum series allowed for a single Series request. The Series call fails if this limit is exceeded. 0 means no limit.
	RequestSampleLimit int

	// The maximum samples allowed for a single Series request, The Series call fails if this limit is exceeded. 0 means no limit.
	// NOTE: For efficiency the limit is internally implemented as 'chunks limit' considering each chunk contains a maximum of 120 samples
	RequestSeriesLimit int

	// If a Store doesn't send any data in this specified duration then a Store will be ignored and partial data will be returned if it's enabled.
	// 0 (default) disables timeout
	ResponseTimeout model.Duration

	// Timeout before an unhealthy store is cleaned from the store UI page. Default is 5m.
	UnhealthyTimeout model.Duration

	// Interval between DNS resolutions for Store endpoints. Default is 30s.
	SDDNSInterval model.Duration

	// DNS Resolver to use for Store endpoints. Possible options: golang, miekgdns. Default is miekgdns.
	SDDNSResolver string

	// Refresh interval to re-read file SD files. It is used as a resync fallback. Default is 5m.
	SDInterval model.Duration

	// Path to files that contain addresses of store API servers. The path can be a glob pattern.
	SDFiles []SDFile
}

// SDFile represents a static service discovery file to be mounted to Querier.
type SDFile struct {
	Data string
	Name string
}

// GRPCOptions specify the client and server gRPC config options for Querier.
type GRPCOptions struct {
	ClientSecure     bool
	ClientSkipVerify bool
	ClientCert       string
	ClientKey        string
	ClientCACert     string
	ClientServerName string

	// Compression algorithm to use for gRPC requests to other clients. Must be one of: ["none", "snappy"]
	// Default is none.
	Compression string

	ServerAddress          string
	ServerTLSCert          string
	ServerTLSKey           string
	ServerTLSClientCA      string
	ServerMaxConnectionAge model.Duration
	ServerGracePeriod      model.Duration

	// Strategy to use when proxying Series requests to leaf nodes. Hidden and only used for testing, will be removed after lazy becomes the default
	ProxyStrategy string
}

// HTTPOptions specify the HTTP config options for Querier.
type HTTPOptions struct {
	BindAddress string
	GracePeriod model.Duration
	TLSConfig   string
}

// WebOptions specify the Web/UI/URL config options for Querier.
type WebOptions struct {
	// 	Whether to disable CORS headers to be set by Thanos. By default Thanos sets CORS headers to be allowed by all.
	DisableCORS bool

	// Prefix for API and UI endpoints. This allows thanos UI to be served on a sub-path. Defaults to the value of --web.external-prefix.
	// This option is analogous to --web.route-prefix of Prometheus.
	RoutePrefix string

	// Static prefix for all HTML links and redirect URLs in the UI query web interface. Actual endpoints are still served on / or the
	// web.route-prefix. This allows thanos UI to be served behind a reverse proxy that strips a URL sub-path.
	ExternalPrefix string

	// Name of HTTP request header used for dynamic prefixing of UI links and redirects.
	// This option is ignored if web.external-prefix argument is set.
	//
	// Security risk: enable this option only if a reverse proxy in front of thanos is resetting the header.
	//
	// The --web.prefix-header=X-Forwarded-Prefix option can be useful, for example, if Thanos UI is served via Traefik reverse proxy
	// with PathPrefixStrip option enabled, which sends the stripped prefix value in X-Forwarded-Prefix header. This allows thanos UI
	// to be served on a sub-path.
	PrefixHeaderName string
}

// QueryTelemetry specifies the query telemetry quantiles for Querier.
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

// WithGRPCOptions allows overriding the default gRPC options for Thanos Querier.
func WithGRPCOptions(opts GRPCOptions) ThanosQueryOption {
	return func(q *query) {
		q.g = opts
	}
}

// WithHTTPOptions allows overriding the default HTTP options for Thanos Querier.
func WithHTTPOptions(opts HTTPOptions) ThanosQueryOption {
	return func(q *query) {
		q.h = opts
	}
}

// WithWebOptions allows overriding the default Web/UI/URL options for Thanos Querier.
func WithWebOptions(opts WebOptions) ThanosQueryOption {
	return func(q *query) {
		q.w = opts
	}
}

// WithStoreOptions allows overriding the default Store (API or SD) options for Thanos Querier.
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

// WithEndpointGroup allows configuring DNS name of statically configured Thanos API
// server groups. Targets resolved from the DNS name will be queried in a round-robin,
// instead of a fanout manner. This option should be used when connecting a Thanos Query
// to HA groups of Thanos components
func WithEndpointGroup(group ...string) ThanosQueryOption {
	return func(q *query) {
		q.endpointGroup = group
	}
}

// WithStrictEndpointGroup allows configuring DNS name of statically configured Thanos API
// server groups that will be always used even if health check fails. Targets resolved from
// the DNS name will be queried in a round-robin, instead of a fanout manner. This option
// should be used when connecting a Thanos Query to HA groups of Thanos components
func WithStrictEndpointGroup(group ...string) ThanosQueryOption {
	return func(q *query) {
		q.endpointGroupStrict = group
	}
}

// WithAutoDownsampling enables auto downsampling by default for Thanos Querier.
func WithAutoDownsampling() ThanosQueryOption {
	return func(q *query) {
		q.autoDownsampling = true
	}
}

// WithActiveQueryPath allows setting a path for active query file for Thanos Querier.
func WithActiveQueryPath(path string) ThanosQueryOption {
	return func(q *query) {
		q.activeQueryPath = path
	}
}

// WithConnectionMetricLabels allows overriding the default selection of query connection
// metric labels to be collected from endpoint set.
func WithConnectionMetricLabels(labels []string) ThanosQueryOption {
	return func(q *query) {
		q.connectionMetricLabels = labels
	}
}

// WithSelectorLabels allows setting selector labels that will be exposed in info endpoint.
func WithSelectorLabels(labels labels.Labels) ThanosQueryOption {
	return func(q *query) {
		q.selectorLabels = labels
	}
}

// WithDefaultEvaluationInterval allows overriding the default evaluation interval for sub queries in Thanos Querier.
func WithDefaultEvaluationInterval(interval string) ThanosQueryOption {
	return func(q *query) {
		pi, err := model.ParseDuration(interval)
		if err != nil {
			panic(err)
		}
		q.defaultEvaluationInterval = pi
	}
}

// WithDefaultStep allows overriding the default step for range queries in Thanos Querier.
// Default step is only used when step is not set in UI. In such cases, Thanos UI will use
// default step to calculate resolution (resolution = max(rangeSeconds / 250, defaultStep)).
// This will not work from Grafana, but Grafana has __step variable which can be used.
func WithDefaultStep(step string) ThanosQueryOption {
	return func(q *query) {
		ps, err := model.ParseDuration(step)
		if err != nil {
			panic(err)
		}
		q.defaultStep = ps
	}
}

// WithLookbackDelta allows overriding the maximum lookback delta used by Thanos Querier.
// for retrieving metrics during expression evaluations. PromQL always evaluates the query
// for the certain timestamp (query range timestamps are deduced by step).
//
// Since scrape intervals might be different, PromQL looks back for given amount of time to
// get latest  sample. If it exceeds the maximum lookback delta it assumes series is stale
// and returns none (a gap). This is why lookback delta should be set to at least 2 times
// of the slowest scrape interval. If unset it will use the promql default of 5m
func WithLookbackDelta(delta string) ThanosQueryOption {
	return func(q *query) {
		pd, err := model.ParseDuration(delta)
		if err != nil {
			panic(err)
		}
		q.lookbackDelta = pd
	}
}

// WithDynamicLookbackDelta allows for larger lookback duration for queries based on resolution
// for Thanos Querier.
func WithDynamicLookbackDelta() ThanosQueryOption {
	return func(q *query) {
		q.dynamicLookbackDelta = true
	}
}

// WithMaxConcurrentQueries allows overriding the default (20) maximum number of concurrent queries
// that can be executed by Thanos Querier.
func WithMaxConcurrentQueries(max int) ThanosQueryOption {
	return func(q *query) {
		q.maxConcurrentQueries = max
	}
}

// WithMaxConcurrentSelects allows overriding the default (4) maximum number of concurrent selects
// that can be executed by Thanos Querier while evaluating a single query.
func WithMaxConcurrentSelects(max int) ThanosQueryOption {
	return func(q *query) {
		q.maxConcurrentSelects = max
	}
}

// WithInstantDefaultMaxSourceResolution allows overriding the default value for max_source_resolution
// for instant queries. If not set, defaults to 0s only taking raw resolution into account. 1h can be a
// good value if you use instant queries over time ranges that incorporate times outside of your raw-retention.
func WithInstantDefaultMaxSourceResolution(resolution string) ThanosQueryOption {
	return func(q *query) {
		pr, err := model.ParseDuration(resolution)
		if err != nil {
			panic(err)
		}
		q.instantDefaultMaxSourceResolution = pr
	}
}

// WithMetadataDefaultTimeRange allows overriding default metadata time range duration for retrieving
// labels through Labels and Series API when the range parameters are not specified. The zero value
// means range covers the time since the beginning.
func WithMetadataDefaultTimeRange(timeRange string) ThanosQueryOption {
	return func(q *query) {
		pr, err := model.ParseDuration(timeRange)
		if err != nil {
			panic(err)
		}
		q.metadataDefaultTimeRange = pr
	}
}

// WithPartialResponse allows enabling partial response for queries if no partial_response param is specified.
func WithPartialResponse() ThanosQueryOption {
	return func(q *query) {
		q.partialResponse = true
	}
}

// WithEngine allows overriding the default PromQL engine for Thanos Querier. Default is thanos.
func WithEngine(engine string) ThanosQueryOption {
	return func(q *query) {
		q.promQLEngine = engine
	}
}

// WithQueryMode allows overriding the default PromQL query mode for Thanos Querier. Default is local.
// Can be set to either local or distributed.
func WithQueryMode(mode string) ThanosQueryOption {
	return func(q *query) {
		q.promQLQueryMode = mode
	}
}

// WithReplicaLabels allows setting labels to treat as a replica indicator along which data is deduplicated.
// Still you will be able to query without deduplication using 'dedup=false' parameter. Data includes
// time series, recording rules, and alerting rules.replica labels in Thanos Querier.
func WithReplicaLabels(labels []string) ThanosQueryOption {
	return func(q *query) {
		q.replicaLabels = labels
	}
}

// WithQueryTelemetry allows overriding the default query request telemetry quantiles for Thanos Querier.
func WithQueryTelemetry(qt QueryTelemetry) ThanosQueryOption {
	return func(q *query) {
		q.telemetry = qt
	}
}

// WithQueryTimeout allows overriding the default (2m) query processing timeout for Thanos Querier.
func WithQueryTimeout(timeout string) ThanosQueryOption {
	return func(q *query) {
		pt, err := model.ParseDuration(timeout)
		if err != nil {
			panic(err)
		}
		q.timeout = pt
	}
}

// WithTracingConfig allows passing in-line YAML tracing configuration for Thanos Querier.
// See format details: https://thanos.io/tip/thanos/tracing.md/#configuration.
// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
//
// Please use the exported TracingConfig and provider configuration structs from
// "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing" package.
func WithTracingConfig(config trclient.TracingConfig) ThanosQueryOption {
	return func(q *query) {
		b, err := yaml.Marshal(config)
		if err != nil {
			panic(err)
		}
		q.tracingConfig = string(b)
	}
}

// WithAlertQueryURL allows setting external Thanos Query URL that would be set in all alerts 'Source' field.
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

// NewThanosQuery returns a new instance of Thanos Query, customized with options.
// Also includes options for adding sidecars.
// Returns the following K8s Objects:
// Deployment
// Service
// ServiceAccount
// Service Discovery ConfigMap (if option is used)
//
// NOTE: You need to call K8sConfig() to customize k8s-native options.
func NewThanosQuery(opts ...ThanosQueryOption) *query {
	q := query{
		logLevel:  "info",
		logFormat: "logfmt",
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
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
		o(&q.DeploymentGenericConfig)
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

	// Instantiate SD file list and configmap if passed.
	sdFileList := []string{}
	sdData := map[string]string{}
	for _, f := range q.s.SDFiles {
		sdFileList = append(sdFileList, "etc/thanos/sd/"+f.Name+".yaml")
		if f.Data != "" {
			sdData[f.Name+".yaml"] = f.Data
		}
	}

	// Build argument list for the Query container.
	args := []string{
		"query",

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

		// Store.
		k8sutil.FlagArg("store.limits.request-series", fmt.Sprint(q.s.RequestSeriesLimit)),
		k8sutil.FlagArg("store.limits.request-samples", fmt.Sprint(q.s.RequestSampleLimit)),
		k8sutil.FlagArg("store.response-timeout", q.s.ResponseTimeout.String()),
		k8sutil.FlagArg("store.unhealthy-timeout", q.s.UnhealthyTimeout.String()),
		k8sutil.FlagArg("store.sd-dns-interval", q.s.SDDNSInterval.String()),
		k8sutil.FlagArg("store.sd-dns-resolver", q.s.SDDNSResolver),
		k8sutil.FlagArg("store.sd-interval", q.s.SDInterval.String()),
	}

	// Labels
	args = append(args, k8sutil.RepeatableFlagArg("query.conn-metric.label", q.connectionMetricLabels)...)
	args = append(args, k8sutil.RepeatableLabelFlagArg("selector-label", q.selectorLabels)...)
	args = append(args, k8sutil.RepeatableFlagArg("query.replica-label", q.replicaLabels)...)

	// Telemetry
	args = append(args, k8sutil.RepeatableFloatFlagArg("query.telemetry.request-duration-seconds-quantiles", q.telemetry.DurationQuantiles)...)
	args = append(args, k8sutil.RepeatableFloatFlagArg("query.telemetry.request-samples-quantiles", q.telemetry.SampleQuantiles)...)
	args = append(args, k8sutil.RepeatableFloatFlagArg("query.telemetry.request-series-seconds-quantiles", q.telemetry.SeriesQuantiles)...)

	// Endpoints.
	args = append(args, k8sutil.RepeatableFlagArg("endpoint", q.endpoints)...)
	args = append(args, k8sutil.RepeatableFlagArg("endpoint-strict", q.endpointsStrict)...)
	args = append(args, k8sutil.RepeatableFlagArg("endpoint-group", q.endpointGroup)...)
	args = append(args, k8sutil.RepeatableFlagArg("endpoint-group-strict", q.endpointGroupStrict)...)

	// SD Files.
	args = append(args, k8sutil.RepeatableFlagArg("store.sd-files", sdFileList)...)

	queryArgs := k8sutil.ArgList(args)

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

	if len(sdFileList) != 0 {
		thanosQueryContainer.VolumeMounts = append(thanosQueryContainer.VolumeMounts,
			corev1.VolumeMount{
				MountPath: "etc/thanos/sd",
				Name:      "file-sd",
			})
	}

	// Attach any configured sidecars.
	containers := []corev1.Container{thanosQueryContainer}
	containers = append(containers, q.Extras.Sidecars...)

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
	ports = append(ports, q.Extras.AdditionalServicePorts...)

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
		"query-service":        &queryService,
	}

	volumes := []corev1.Volume{}

	if len(sdData) != 0 {
		sdConfigMap := corev1.ConfigMap{
			TypeMeta:   k8sutil.ConfigMapMeta,
			ObjectMeta: commonObjectMeta,
			Data:       sdData,
		}

		manifests["query-sd-configmap"] = &sdConfigMap

		volumes = append(volumes, corev1.Volume{
			Name: "file-sd",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sdConfigMap.Name,
					},
				},
			},
		})
	}

	// Attach any configured volumes.
	volumes = append(volumes, q.Extras.AdditionalPodVolumes...)
	queryDeployment.Spec.Template.Spec.Volumes = volumes

	manifests["query-deployment"] = &queryDeployment

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
		endpoints = append(endpoints, q.Extras.AdditionalServiceMonitorPorts...)

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
