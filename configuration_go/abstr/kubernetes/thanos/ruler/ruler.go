package ruler

import (
	"net"
	"path/filepath"
	"strings"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/objstore"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

type HashFunc string
type QueryHttpMethod string

const (
	defaultHTTPPort     int             = 10902
	defaultGRPCPort     int             = 10901
	dataVolumeName      string          = "data"
	HashFuncSHA256      HashFunc        = "SHA256"
	QueryHttpMethodGET  QueryHttpMethod = "GET"
	QueryHttpMethodPOST QueryHttpMethod = "POST"
	baseRulesDir        string          = "/etc/thanos/rules"
)

type alertRelabelConfigFile = k8sutil.ConfigFile

// NewAlertRelabelConfigFile returns a new alertRelabelConfigFile option
func NewAlertRelabelConfigFile(value *relabel.Config) *alertRelabelConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/relabel", "config.yaml", "relabel", "observatorium-rule-relabel")
	if value != nil {
		valueYaml, err := yaml.Marshal(value)
		if err != nil {
			panic(err)
		}

		ret.WithValue(string(valueYaml))
	}
	return ret
}

type alertmanagersConfigFile = k8sutil.ConfigFile

// NewAlertmanagersConfigFile returns a new alertmanagersConfigFile option
func NewAlertmanagersConfigFile(value *AlertingConfig) *alertmanagersConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/alertmanagers", "config.yaml", "alertmanagers", "observatorium-rule-alertmanagers")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type tracingConfigFile = k8sutil.ConfigFile

// NewTracingConfigFile returns a new tracing config file k8sutil.
func NewTracingConfigFile(value *trclient.TracingConfig) *tracingConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/tracing", "config.yaml", "tracing", "observatorium-rule-tracing")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type objstoreConfigFile = k8sutil.ConfigFile

// NewObjstoreConfigFile returns a new alertRelabelConfigFile option
func NewObjstoreConfigFile(value *objstore.BucketConfig) *objstoreConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/objstore", "config.yaml", "objstore", "observatorium-rule-objstore")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type RuleFileOption struct {
	FileName string
	// If the rules are contained in a shared volume, specify the volume name.
	VolumeName string
	// If the rules are in a configMap, specify the configMap name.
	ConfigMapName string
	// Optionally specify the parent directory of the rule file to avoid path conflicts.
	ParentDir string
}

func (r RuleFileOption) FilePath() string {
	parentDir := r.ParentDir
	if parentDir == "" {
		parentDir = strings.TrimSuffix(r.FileName, filepath.Ext(r.FileName))
	}

	return filepath.Join(baseRulesDir, parentDir)
}

func (r RuleFileOption) String() string {
	return filepath.Join(r.FilePath(), r.FileName)
}

type RulerOptions struct {
	AlertLabelDrop             []string                 `opt:"alert.label-drop"`
	AlertQeuryTemplate         string                   `opt:"alert.query-template"`
	AlertQueryUrl              string                   `opt:"alert.query-url"`
	AlertRelabelConfig         *relabel.Config          `opt:"alert.relabel-config"`
	AlertRelabelConfigFile     *alertRelabelConfigFile  `opt:"alert.relabel-config-file"`
	AlertmanagersConfig        *AlertingConfig          `opt:"alertmanagers.config"` // check
	AlertmanagersConfigFile    *alertmanagersConfigFile `opt:"alertmanagers.config-file"`
	AlertmanagersSdDnsInterval time.Duration            `opt:"alertmanagers.sd-dns-interval"`
	AlertmanagersSendTimeout   time.Duration            `opt:"alertmanagers.send-timeout"`
	AlertmanagersUrl           []string                 `opt:"alertmanagers.url"`
	DataDir                    string                   `opt:"data-dir"`
	EvalInterval               time.Duration            `opt:"eval-interval"`
	ForGracePeriod             time.Duration            `opt:"for-grace-period"`
	ForOutageTolerance         time.Duration            `opt:"for-outage-tolerance"`
	GrpcAddress                *net.TCPAddr             `opt:"grpc-address"`
	GrpcGracePeriod            time.Duration            `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge time.Duration            `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert          string                   `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa      string                   `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey           string                   `opt:"grpc-server-tls-key"`
	HashFunc                   HashFunc                 `opt:"hash-func"`
	HttpAddress                *net.TCPAddr             `opt:"http-address"`
	HttpGracePeriod            time.Duration            `opt:"http-grace-period"`
	HttpConfig                 string                   `opt:"http.config"`
	Label                      []Label                  `opt:"label"`
	LogFormat                  log.LogFormat            `opt:"log.format"`
	LogLevel                   log.LogLevel             `opt:"log.level"`
	ObjstoreConfig             string                   `opt:"objstore.config"`
	ObjstoreConfigFile         *objstoreConfigFile      `opt:"objstore.config-file"`
	Query                      []string                 `opt:"query"`
	QueryConfig                string                   `opt:"query.config"`      //todo
	QueryConfigFile            string                   `opt:"query.config-file"` //todo
	QueryDefaultStep           time.Duration            `opt:"query.default-step"`
	QueryHttpMethod            QueryHttpMethod          `opt:"query.http-method"`
	QuerySdDnsInterval         time.Duration            `opt:"query.sd-dns-interval"`
	QuerySdFiles               []string                 `opt:"query.sd-files"`
	QuerySdInterval            time.Duration            `opt:"query.sd-interval"`
	RemoteWriteConfig          string                   `opt:"remote-write.config"`         //todo
	RemoteWriteConfigFile      string                   `opt:"remote-write.config-file"`    //todo
	RequestLoggingConfig       string                   `opt:"request.logging-config"`      //todo
	RequestLoggingConfigFile   string                   `opt:"request.logging-config-file"` //todo
	ResendDelay                time.Duration            `opt:"resend-delay"`
	RestoreIgnoredLabel        []string                 `opt:"restore-ignored-label"`
	RuleFile                   []RuleFileOption         `opt:"rule-file"`
	ShipperMetaFileName        string                   `opt:"shipper.meta-file-name"`
	ShipperUploadCompacted     bool                     `opt:"shipper.upload-compacted,noval"`
	StoreLimitsRequestSamples  int                      `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries   int                      `opt:"store.limits.request-series"`
	TracingConfig              *trclient.TracingConfig  `opt:"tracing.config"`      //todo
	TracingConfigFile          *tracingConfigFile       `opt:"tracing.config-file"` //todo
	TsdbBlockDuration          time.Duration            `opt:"tsdb.block-duration"`
	TsdbNoLockfile             bool                     `opt:"tsdb.no-lockfile,noval"`
	TsdbRetention              time.Duration            `opt:"tsdb.retention"`
	TsdbWalCompression         bool                     `opt:"tsdb.wal-compression,noval"`
	WebDisableCors             bool                     `opt:"web.disable-cors,noval"`
	WebExternalPrefix          string                   `opt:"web.external-prefix"`
	WebPrefixHeader            string                   `opt:"web.prefix-header"`
	WebRoutePrefix             string                   `opt:"web.route-prefix"`
}

type RulerStatefulSet struct {
	options    *RulerOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewDefaultOptions() *RulerOptions {
	return &RulerOptions{
		LogLevel:       "warn",
		LogFormat:      "logfmt",
		DataDir:        "/var/thanos/ruler",
		ObjstoreConfig: "$(OBJSTORE_CONFIG)",
	}
}

func NewRuler(opts *RulerOptions, namespace, imageTag string) *RulerStatefulSet {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-rule",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "rule-evaluation-engine",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	return &RulerStatefulSet{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-ruler",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("500m", "1", "200Mi", "400Mi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", probePort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", probePort, k8sutil.ProbeConfig{
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

func (r *RulerStatefulSet) Manifests() k8sutil.ObjectMap {
	container := r.makeContainer()

	ret := k8sutil.ObjectMap{}
	ret.AddAll(r.GenerateObjectsStatefulSet(container))

	return ret
}

func (s *RulerStatefulSet) makeContainer() *k8sutil.Container {
	httpPort := k8sutil.GetPortOrDefault(defaultHTTPPort, s.options.HttpAddress)
	grpcPort := k8sutil.GetPortOrDefault(defaultGRPCPort, s.options.GrpcAddress)

	k8sutil.CheckProbePort(httpPort, s.LivenessProbe)
	k8sutil.CheckProbePort(httpPort, s.ReadinessProbe)

	if s.options.DataDir == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"rule"}, cmdopt.GetOpts(s.options)...)
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
			MountPath: s.options.DataDir,
		},
	}

	if s.options.AlertRelabelConfigFile != nil {
		s.options.AlertRelabelConfigFile.AddToContainer(ret)
	}

	if s.options.AlertmanagersConfigFile != nil {
		s.options.AlertmanagersConfigFile.AddToContainer(ret)
	}

	if s.options.TracingConfigFile != nil {
		s.options.TracingConfigFile.AddToContainer(ret)
	}

	if s.options.ObjstoreConfigFile != nil {
		s.options.ObjstoreConfigFile.AddToContainer(ret)
	}

	for _, ruleFile := range s.options.RuleFile {
		if ruleFile.VolumeName != "" {
			ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
				Name:      ruleFile.VolumeName,
				MountPath: ruleFile.FilePath(),
				ReadOnly:  true,
			})
		} else if ruleFile.ConfigMapName != "" {
			ret.Volumes = append(ret.Volumes, k8sutil.NewPodVolumeFromConfigMap(ruleFile.ConfigMapName, ruleFile.ConfigMapName))

			ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
				Name:      ruleFile.ConfigMapName,
				MountPath: ruleFile.FilePath(),
			})
		}
	}

	return ret
}
