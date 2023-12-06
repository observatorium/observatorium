package ruler

import (
	"fmt"
	"net"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	thanoslog "github.com/observatorium/observatorium/configuration_go/schemas/thanos/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/option"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	corev1 "k8s.io/api/core/v1"
)

// --alert.label-drop=ALERT.LABEL-DROP ...
//                                  Labels by name to drop before sending
//                                  to alertmanager. This allows alert to be
//                                  deduplicated on replica label (repeated).
//                                  Similar Prometheus alert relabelling
//       --alert.query-template="/graph?g0.expr={{.Expr}}&g0.tab=1"
//                                  Template to use in alerts source field.
//                                  Need only include {{.Expr}} parameter
//       --alert.query-url=ALERT.QUERY-URL
//                                  The external Thanos Query URL that would be set
//                                  in all alerts 'Source' field
//       --alert.relabel-config=<content>
//                                  Alternative to 'alert.relabel-config-file' flag
//                                  (mutually exclusive). Content of YAML file that
//                                  contains alert relabelling configuration.
//       --alert.relabel-config-file=<file-path>
//                                  Path to YAML file that contains alert
//                                  relabelling configuration.
//       --alertmanagers.config=<content>
//                                  Alternative to 'alertmanagers.config-file'
//                                  flag (mutually exclusive). Content
//                                  of YAML file that contains alerting
//                                  configuration. See format details:
//                                  https://thanos.io/tip/components/rule.md/#configuration.
//                                  If defined, it takes precedence
//                                  over the '--alertmanagers.url' and
//                                  '--alertmanagers.send-timeout' flags.
//       --alertmanagers.config-file=<file-path>
//                                  Path to YAML file that contains alerting
//                                  configuration. See format details:
//                                  https://thanos.io/tip/components/rule.md/#configuration.
//                                  If defined, it takes precedence
//                                  over the '--alertmanagers.url' and
//                                  '--alertmanagers.send-timeout' flags.
//       --alertmanagers.sd-dns-interval=30s
//                                  Interval between DNS resolutions of
//                                  Alertmanager hosts.
//       --alertmanagers.send-timeout=10s
//                                  Timeout for sending alerts to Alertmanager
//       --alertmanagers.url=ALERTMANAGERS.URL ...
//                                  Alertmanager replica URLs to push firing
//                                  alerts. Ruler claims success if push to
//                                  at least one alertmanager from discovered
//                                  succeeds. The scheme should not be empty
//                                  e.g `http` might be used. The scheme may be
//                                  prefixed with 'dns+' or 'dnssrv+' to detect
//                                  Alertmanager IPs through respective DNS
//                                  lookups. The port defaults to 9093 or the
//                                  SRV record's value. The URL path is used as a
//                                  prefix for the regular Alertmanager API path.
//       --data-dir="data/"         data directory
//       --eval-interval=1m         The default evaluation interval to use.
//       --for-grace-period=10m     Minimum duration between alert and restored
//                                  "for" state. This is maintained only for alerts
//                                  with configured "for" time greater than grace
//                                  period.
//       --for-outage-tolerance=1h  Max time to tolerate prometheus outage for
//                                  restoring "for" state of alert.
//       --grpc-address="0.0.0.0:10901"
//                                  Listen ip:port address for gRPC endpoints
//                                  (StoreAPI). Make sure this address is routable
//                                  from other components.
//       --grpc-grace-period=2m     Time to wait after an interrupt received for
//                                  GRPC Server.
//       --grpc-server-max-connection-age=60m
//                                  The grpc server max connection age. This
//                                  controls how often to re-establish connections
//                                  and redo TLS handshakes.
//       --grpc-server-tls-cert=""  TLS Certificate for gRPC server, leave blank to
//                                  disable TLS
//       --grpc-server-tls-client-ca=""
//                                  TLS CA to verify clients against. If no
//                                  client CA is specified, there is no client
//                                  verification on server side. (tls.NoClientCert)
//       --grpc-server-tls-key=""   TLS Key for the gRPC server, leave blank to
//                                  disable TLS
//       --hash-func=               Specify which hash function to use when
//                                  calculating the hashes of produced files.
//                                  If no function has been specified, it does not
//                                  happen. This permits avoiding downloading some
//                                  files twice albeit at some performance cost.
//                                  Possible values are: "", "SHA256".
//   -h, --help                     Show context-sensitive help (also try
//                                  --help-long and --help-man).
//       --http-address="0.0.0.0:10902"
//                                  Listen host:port for HTTP endpoints.
//       --http-grace-period=2m     Time to wait after an interrupt received for
//                                  HTTP Server.
//       --http.config=""           [EXPERIMENTAL] Path to the configuration file
//                                  that can enable TLS or authentication for all
//                                  HTTP endpoints.
//       --label=<name>="<value>" ...
//                                  Labels to be applied to all generated metrics
//                                  (repeated). Similar to external labels for
//                                  Prometheus, used to identify ruler and its
//                                  blocks as unique source.
//       --log.format=logfmt        Log format to use. Possible options: logfmt or
//                                  json.
//       --log.level=info           Log filtering level.
//       --objstore.config=<content>
//                                  Alternative to 'objstore.config-file'
//                                  flag (mutually exclusive). Content of
//                                  YAML file that contains object store
//                                  configuration. See format details:
//                                  https://thanos.io/tip/thanos/storage.md/#configuration
//       --objstore.config-file=<file-path>
//                                  Path to YAML file that contains object
//                                  store configuration. See format details:
//                                  https://thanos.io/tip/thanos/storage.md/#configuration
//       --query=<query> ...        Addresses of statically configured query
//                                  API servers (repeatable). The scheme may be
//                                  prefixed with 'dns+' or 'dnssrv+' to detect
//                                  query API servers through respective DNS
//                                  lookups.
//       --query.config=<content>   Alternative to 'query.config-file' flag
//                                  (mutually exclusive). Content of YAML
//                                  file that contains query API servers
//                                  configuration. See format details:
//                                  https://thanos.io/tip/components/rule.md/#configuration.
//                                  If defined, it takes precedence over the
//                                  '--query' and '--query.sd-files' flags.
//       --query.config-file=<file-path>
//                                  Path to YAML file that contains query API
//                                  servers configuration. See format details:
//                                  https://thanos.io/tip/components/rule.md/#configuration.
//                                  If defined, it takes precedence over the
//                                  '--query' and '--query.sd-files' flags.
//       --query.default-step=1s    Default range query step to use. This is
//                                  only used in stateless Ruler and alert state
//                                  restoration.
//       --query.http-method=POST   HTTP method to use when sending queries.
//                                  Possible options: [GET, POST]
//       --query.sd-dns-interval=30s
//                                  Interval between DNS resolutions.
//       --query.sd-files=<path> ...
//                                  Path to file that contains addresses of query
//                                  API servers. The path can be a glob pattern
//                                  (repeatable).
//       --query.sd-interval=5m     Refresh interval to re-read file SD files.
//                                  (used as a fallback)
//       --remote-write.config=<content>
//                                  Alternative to 'remote-write.config-file'
//                                  flag (mutually exclusive). Content
//                                  of YAML config for the remote-write
//                                  configurations, that specify servers
//                                  where samples should be sent to (see
//                                  https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write).
//                                  This automatically enables stateless mode
//                                  for ruler and no series will be stored in the
//                                  ruler's TSDB. If an empty config (or file) is
//                                  provided, the flag is ignored and ruler is run
//                                  with its own TSDB.
//       --remote-write.config-file=<file-path>
//                                  Path to YAML config for the remote-write
//                                  configurations, that specify servers
//                                  where samples should be sent to (see
//                                  https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write).
//                                  This automatically enables stateless mode
//                                  for ruler and no series will be stored in the
//                                  ruler's TSDB. If an empty config (or file) is
//                                  provided, the flag is ignored and ruler is run
//                                  with its own TSDB.
//       --request.logging-config=<content>
//                                  Alternative to 'request.logging-config-file'
//                                  flag (mutually exclusive). Content
//                                  of YAML file with request logging
//                                  configuration. See format details:
//                                  https://thanos.io/tip/thanos/logging.md/#configuration
//       --request.logging-config-file=<file-path>
//                                  Path to YAML file with request logging
//                                  configuration. See format details:
//                                  https://thanos.io/tip/thanos/logging.md/#configuration
//       --resend-delay=1m          Minimum amount of time to wait before resending
//                                  an alert to Alertmanager.
//       --restore-ignored-label=RESTORE-IGNORED-LABEL ...
//                                  Label names to be ignored when restoring alerts
//                                  from the remote storage. This is only used in
//                                  stateless mode.
//       --rule-file=rules/ ...     Rule files that should be used by rule
//                                  manager. Can be in glob format (repeated).
//                                  Note that rules are not automatically detected,
//                                  use SIGHUP or do HTTP POST /-/reload to re-read
//                                  them.
//       --shipper.meta-file-name="thanos.shipper.json"
//                                  the file to store shipper metadata in
//       --shipper.upload-compacted
//                                  If true shipper will try to upload compacted
//                                  blocks as well. Useful for migration purposes.
//                                  Works only if compaction is disabled on
//                                  Prometheus. Do it once and then disable the
//                                  flag when done.
//       --store.limits.request-samples=0
//                                  The maximum samples allowed for a single
//                                  Series request, The Series call fails if
//                                  this limit is exceeded. 0 means no limit.
//                                  NOTE: For efficiency the limit is internally
//                                  implemented as 'chunks limit' considering each
//                                  chunk contains a maximum of 120 samples.
//       --store.limits.request-series=0
//                                  The maximum series allowed for a single Series
//                                  request. The Series call fails if this limit is
//                                  exceeded. 0 means no limit.
//       --tracing.config=<content>
//                                  Alternative to 'tracing.config-file' flag
//                                  (mutually exclusive). Content of YAML file
//                                  with tracing configuration. See format details:
//                                  https://thanos.io/tip/thanos/tracing.md/#configuration
//       --tracing.config-file=<file-path>
//                                  Path to YAML file with tracing
//                                  configuration. See format details:
//                                  https://thanos.io/tip/thanos/tracing.md/#configuration
//       --tsdb.block-duration=2h   Block duration for TSDB block.
//       --tsdb.no-lockfile         Do not create lockfile in TSDB data directory.
//                                  In any case, the lockfiles will be deleted on
//                                  next startup.
//       --tsdb.retention=48h       Block retention time on local disk.
//       --tsdb.wal-compression     Compress the tsdb WAL.
//       --version                  Show application version.
//       --web.disable-cors         Whether to disable CORS headers to be set by
//                                  Thanos. By default Thanos sets CORS headers to
//                                  be allowed by all.
//       --web.external-prefix=""   Static prefix for all HTML links and redirect
//                                  URLs in the bucket web UI interface.
//                                  Actual endpoints are still served on / or the
//                                  web.route-prefix. This allows thanos bucket
//                                  web UI to be served behind a reverse proxy that
//                                  strips a URL sub-path.
//       --web.prefix-header=""     Name of HTTP request header used for dynamic
//                                  prefixing of UI links and redirects.
//                                  This option is ignored if web.external-prefix
//                                  argument is set. Security risk: enable
//                                  this option only if a reverse proxy in
//                                  front of thanos is resetting the header.
//                                  The --web.prefix-header=X-Forwarded-Prefix
//                                  option can be useful, for example, if Thanos
//                                  UI is served via Traefik reverse proxy with
//                                  PathPrefixStrip option enabled, which sends the
//                                  stripped prefix value in X-Forwarded-Prefix
//                                  header. This allows thanos UI to be served on a
//                                  sub-path.
//       --web.route-prefix=""      Prefix for API and UI endpoints. This allows
//                                  thanos UI to be served on a sub-path. This
//                                  option is analogous to --web.route-prefix of
//                                  Prometheus.

type HashFunc string
type QueryHttpMethod string

const (
	defaultNamespace    string          = "observatorium"
	defaultHTTPPort     int             = 10902
	defaultGRPCPort     int             = 10901
	dataVolumeName      string          = "data"
	HashFuncSHA256      HashFunc        = "SHA256"
	QueryHttpMethodGET  QueryHttpMethod = "GET"
	QueryHttpMethodPOST QueryHttpMethod = "POST"
)

type alertRelabelConfigFile = option.ConfigFile[relabel.Config]

// NewAlertRelabelConfigFile returns a new alertRelabelConfigFile option
func NewAlertRelabelConfigFile(name string, value relabel.Config) *alertRelabelConfigFile {
	return option.NewConfigFile("/etc/thanos/relabel", "config.yaml", name, value)
}

type alertmanagersConfigFile = option.ConfigFile[AlertingConfig]

// NewAlertmanagersConfigFile returns a new alertmanagersConfigFile option
func NewAlertmanagersConfigFile(name string, value AlertingConfig) *alertmanagersConfigFile {
	return option.NewConfigFile("/etc/thanos/alertmanagers", "config.yaml", name, value)
}

type tracingConfigFile = option.ConfigFile[trclient.TracingConfig]

// NewReceiveLimitsConfigFile returns a new tracing config file option.
func NewTracingConfigFile(name string, value trclient.TracingConfig) *tracingConfigFile {
	return option.NewConfigFile("/etc/thanos/tracing", "config.yaml", name, value)
}

// type objstoreConfigFile = option.ConfigFile[objstore.Config]

// // NewObjstoreConfigFile returns a new objstoreConfigFile option
// func NewObjstoreConfigFile(name string, value objstore.Config) objstoreConfigFile {
// 	return option.NewConfigFile("/etc/thanos/objstore", value)
// }

type RulerOptions struct {
	AlertLabelDrop             []string                 `opt:"alert.label-drop"`
	AlertQeuryTemplate         string                   `opt:"alert.query-template"`
	AlertQueryUrl              string                   `opt:"alert.query-url"`
	AlertRelabelConfig         *relabel.Config          `opt:"alert.relabel-config"`
	AlertRelabelConfigFile     *alertRelabelConfigFile  `opt:"alert.relabel-config-file"`
	AlertmanagersConfig        *AlertingConfig          `opt:"alertmanagers.config"` // check
	AlertmanagersConfigFile    *alertmanagersConfigFile `opt:"alertmanagers.config-file"`
	AlertmanagersSdDnsInterval model.Duration           `opt:"alertmanagers.sd-dns-interval"`
	AlertmanagersSendTimeout   model.Duration           `opt:"alertmanagers.send-timeout"`
	AlertmanagersUrl           []string                 `opt:"alertmanagers.url"`
	DataDir                    string                   `opt:"data-dir"`
	EvalInterval               model.Duration           `opt:"eval-interval"`
	ForGracePeriod             model.Duration           `opt:"for-grace-period"`
	ForOutageTolerance         model.Duration           `opt:"for-outage-tolerance"`
	GrpcAddress                *net.TCPAddr             `opt:"grpc-address"`
	GrpcGracePeriod            model.Duration           `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge model.Duration           `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert          string                   `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa      string                   `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey           string                   `opt:"grpc-server-tls-key"`
	HashFunc                   HashFunc                 `opt:"hash-func"`
	HttpAddress                *net.TCPAddr             `opt:"http-address"`
	HttpGracePeriod            model.Duration           `opt:"http-grace-period"`
	HttpConfig                 string                   `opt:"http.config"`
	Label                      []Label                  `opt:"label"`
	LogFormat                  thanoslog.LogFormat      `opt:"log.format"`
	LogLevel                   thanoslog.LogLevel       `opt:"log.level"`
	ObjstoreConfig             string                   `opt:"objstore.config"`
	ObjstoreConfigFile         string                   `opt:"objstore.config-file"`
	Query                      []string                 `opt:"query"`
	QueryConfig                string                   `opt:"query.config"`      //todo
	QueryConfigFile            string                   `opt:"query.config-file"` //todo
	QueryDefaultStep           model.Duration           `opt:"query.default-step"`
	QueryHttpMethod            QueryHttpMethod          `opt:"query.http-method"`
	QuerySdDnsInterval         model.Duration           `opt:"query.sd-dns-interval"`
	QuerySdFiles               []string                 `opt:"query.sd-files"`
	QuerySdInterval            model.Duration           `opt:"query.sd-interval"`
	RemoteWriteConfig          string                   `opt:"remote-write.config"`         //todo
	RemoteWriteConfigFile      string                   `opt:"remote-write.config-file"`    //todo
	RequestLoggingConfig       string                   `opt:"request.logging-config"`      //todo
	RequestLoggingConfigFile   string                   `opt:"request.logging-config-file"` //todo
	ResendDelay                model.Duration           `opt:"resend-delay"`
	RestoreIgnoredLabel        []string                 `opt:"restore-ignored-label"`
	RuleFile                   []string                 `opt:"rule-file"`
	ShipperMetaFileName        string                   `opt:"shipper.meta-file-name"`
	ShipperUploadCompacted     bool                     `opt:"shipper.upload-compacted,noval"`
	StoreLimitsRequestSamples  int                      `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries   int                      `opt:"store.limits.request-series"`
	TracingConfig              *trclient.TracingConfig  `opt:"tracing.config"`      //todo
	TracingConfigFile          *tracingConfigFile       `opt:"tracing.config-file"` //todo
	TsdbBlockDuration          model.Duration           `opt:"tsdb.block-duration"`
	TsdbNoLockfile             bool                     `opt:"tsdb.no-lockfile,noval"`
	TsdbRetention              model.Duration           `opt:"tsdb.retention"`
	TsdbWalCompression         bool                     `opt:"tsdb.wal-compression,noval"`
	WebDisableCors             bool                     `opt:"web.disable-cors,noval"`
	WebExternalPrefix          string                   `opt:"web.external-prefix"`
	WebPrefixHeader            string                   `opt:"web.prefix-header"`
	WebRoutePrefix             string                   `opt:"web.route-prefix"`
}

type RulerStatefulSet struct {
	Options    *RulerOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewRuler() *RulerStatefulSet {
	opts := &RulerOptions{
		LogLevel:       "warn",
		LogFormat:      "logfmt",
		DataDir:        "/var/thanos/ruler",
		ObjstoreConfig: "$(OBJSTORE_CONFIG)",
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-rule",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "rule-evaluation-engine",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &RulerStatefulSet{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-ruler",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("500m", "1", "200Mi", "400Mi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", defaultHTTPPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", defaultHTTPPort, k8sutil.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				k8sutil.NewEnvFromField("POD_NAME", "metadata.name"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
		VolumeSize: "50Gi",
	}
}

func (s *RulerStatefulSet) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      s.Name,
		Labels:    s.CommonLabels,
		Namespace: s.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &s.TerminationGracePeriodSeconds,
		Affinity:                      s.Affinity,
		SecurityContext:               s.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container}, s.Sidecars...),
	}

	statefulset := &k8sutil.StatefulSet{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   s.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"ruler-statefulSet": statefulset.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["ruler-service"] = service.MakeManifest()

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["ruler-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["ruler-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["ruler-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["ruler-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *RulerStatefulSet) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &RulerOptions{}
	}

	httpPort := defaultHTTPPort
	if s.Options.HttpAddress != nil && s.Options.HttpAddress.Port != 0 {
		httpPort = s.Options.HttpAddress.Port
	}

	grpcPort := defaultGRPCPort
	if s.Options.GrpcAddress != nil && s.Options.GrpcAddress.Port != 0 {
		grpcPort = s.Options.GrpcAddress.Port
	}

	livenessPort := s.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(httpPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, httpPort))
	}

	readinessPort := s.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(httpPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, httpPort))
	}

	if s.Options.DataDir == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"rule"}, cmdopt.GetOpts(s.Options)...)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "grpc",
			ContainerPort: int32(grpcPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("http", httpPort, httpPort),
		k8sutil.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}
	ret.VolumeClaims = []k8sutil.VolumeClaim{
		k8sutil.NewVolumeClaimProvider(dataVolumeName, s.VolumeType, s.VolumeSize),
	}
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: s.Options.DataDir,
		},
	}

	return ret
}
