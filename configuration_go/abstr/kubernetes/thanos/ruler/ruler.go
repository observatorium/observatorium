package ruler

import (
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/objstore"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// NewAlertRelabelConfigFile returns a new alertRelabelConfigFile option
func NewAlertRelabelConfigFile(value *relabel.Config) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/relabel", "config.yaml", "relabel", "observatorium-rule-relabel")
	if value != nil {
		valueYaml, err := yaml.Marshal(value)
		if err != nil {
			panic(err)
		}

		ret.WithValue(string(valueYaml))
	}
	return ret
}

// NewAlertmanagersConfigFile returns a new alertmanagersConfigFile option
func NewAlertmanagersConfigFile(value *AlertingConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/alertmanagers", "config.yaml", "alertmanagers", "observatorium-rule-alertmanagers")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewTracingConfigFile returns a new tracing config file.
func NewTracingConfigFile(value *trclient.TracingConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/tracing", "config.yaml", "tracing", "observatorium-rule-tracing")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewObjstoreConfigFile returns a new alertRelabelConfigFile option
func NewObjstoreConfigFile(value *objstore.BucketConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/objstore", "config.yaml", "objstore", "observatorium-rule-objstore")
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
	AlertLabelDrop             []string                       `opt:"alert.label-drop"`
	AlertQeuryTemplate         string                         `opt:"alert.query-template"`
	AlertQueryUrl              string                         `opt:"alert.query-url"`
	AlertRelabelConfig         *relabel.Config                `opt:"alert.relabel-config"`
	AlertRelabelConfigFile     containeropts.ContainerUpdater `opt:"alert.relabel-config-file"`
	AlertmanagersConfig        *AlertingConfig                `opt:"alertmanagers.config"` // check
	AlertmanagersConfigFile    containeropts.ContainerUpdater `opt:"alertmanagers.config-file"`
	AlertmanagersSdDnsInterval time.Duration                  `opt:"alertmanagers.sd-dns-interval"`
	AlertmanagersSendTimeout   time.Duration                  `opt:"alertmanagers.send-timeout"`
	AlertmanagersUrl           []string                       `opt:"alertmanagers.url"`
	DataDir                    string                         `opt:"data-dir"`
	EvalInterval               time.Duration                  `opt:"eval-interval"`
	ForGracePeriod             time.Duration                  `opt:"for-grace-period"`
	ForOutageTolerance         time.Duration                  `opt:"for-outage-tolerance"`
	GrpcAddress                *net.TCPAddr                   `opt:"grpc-address"`
	GrpcGracePeriod            time.Duration                  `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge time.Duration                  `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert          string                         `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa      string                         `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey           string                         `opt:"grpc-server-tls-key"`
	HashFunc                   HashFunc                       `opt:"hash-func"`
	HttpAddress                *net.TCPAddr                   `opt:"http-address"`
	HttpGracePeriod            time.Duration                  `opt:"http-grace-period"`
	HttpConfig                 string                         `opt:"http.config"`
	Label                      []Label                        `opt:"label"`
	LogFormat                  log.Format                     `opt:"log.format"`
	LogLevel                   log.Level                      `opt:"log.level"`
	ObjstoreConfig             string                         `opt:"objstore.config"`
	ObjstoreConfigFile         containeropts.ContainerUpdater `opt:"objstore.config-file"`
	Query                      []string                       `opt:"query"`
	QueryConfig                string                         `opt:"query.config"`      //todo
	QueryConfigFile            string                         `opt:"query.config-file"` //todo
	QueryDefaultStep           time.Duration                  `opt:"query.default-step"`
	QueryHttpMethod            QueryHttpMethod                `opt:"query.http-method"`
	QuerySdDnsInterval         time.Duration                  `opt:"query.sd-dns-interval"`
	QuerySdFiles               []string                       `opt:"query.sd-files"`
	QuerySdInterval            time.Duration                  `opt:"query.sd-interval"`
	RemoteWriteConfig          string                         `opt:"remote-write.config"`         //todo
	RemoteWriteConfigFile      string                         `opt:"remote-write.config-file"`    //todo
	RequestLoggingConfig       string                         `opt:"request.logging-config"`      //todo
	RequestLoggingConfigFile   string                         `opt:"request.logging-config-file"` //todo
	ResendDelay                time.Duration                  `opt:"resend-delay"`
	RestoreIgnoredLabel        []string                       `opt:"restore-ignored-label"`
	RuleFile                   []RuleFileOption               `opt:"rule-file"`
	ShipperMetaFileName        string                         `opt:"shipper.meta-file-name"`
	ShipperUploadCompacted     bool                           `opt:"shipper.upload-compacted,noval"`
	StoreLimitsRequestSamples  int                            `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries   int                            `opt:"store.limits.request-series"`
	TracingConfig              *trclient.TracingConfig        `opt:"tracing.config"`      //todo
	TracingConfigFile          containeropts.ContainerUpdater `opt:"tracing.config-file"` //todo
	TsdbBlockDuration          time.Duration                  `opt:"tsdb.block-duration"`
	TsdbNoLockfile             bool                           `opt:"tsdb.no-lockfile,noval"`
	TsdbRetention              time.Duration                  `opt:"tsdb.retention"`
	TsdbWalCompression         bool                           `opt:"tsdb.wal-compression,noval"`
	WebDisableCors             bool                           `opt:"web.disable-cors,noval"`
	WebExternalPrefix          string                         `opt:"web.external-prefix"`
	WebPrefixHeader            string                         `opt:"web.prefix-header"`
	WebRoutePrefix             string                         `opt:"web.route-prefix"`

	// Extra options not officially supported.
	cmdopt.ExtraOpts
}

type RulerStatefulSet struct {
	options *RulerOptions
	workload.StatefulSetWorkload
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
		workload.NameLabel:      "thanos-rule",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "rule-evaluation-engine",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	ssWorkload := workload.StatefulSetWorkload{
		Replicas:   1,
		VolumeSize: "50Gi",
		PodConfig: workload.PodConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-thanos-ruler",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("500m", "1", "200Mi", "400Mi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: kghelpers.NewProbe("/-/healthy", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: kghelpers.NewProbe("/-/ready", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 20,
				PeriodSeconds:    5,
			}),
			TerminationGracePeriodSeconds: 120,
			Env: []corev1.EnvVar{
				kghelpers.NewEnvFromField("POD_NAME", "metadata.name"),
			},
			ConfigMaps: make(map[string]map[string]string),
			Secrets:    make(map[string]map[string][]byte),
		},
	}

	return &RulerStatefulSet{
		options:             opts,
		StatefulSetWorkload: ssWorkload,
	}
}

func (r *RulerStatefulSet) Objects() []runtime.Object {
	container := r.makeContainer()
	return r.StatefulSetWorkload.Objects(container)
}

func (s *RulerStatefulSet) makeContainer() *workload.Container {
	httpPort := kghelpers.GetPortOrDefault(defaultHTTPPort, s.options.HttpAddress)
	grpcPort := kghelpers.GetPortOrDefault(defaultGRPCPort, s.options.GrpcAddress)

	kghelpers.CheckProbePort(httpPort, s.LivenessProbe)
	kghelpers.CheckProbePort(httpPort, s.ReadinessProbe)

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
		kghelpers.NewServicePort("http", httpPort, httpPort),
		kghelpers.NewServicePort("grpc", grpcPort, grpcPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}
	ret.VolumeClaims = append(ret.VolumeClaims, workload.PersistentVolumeClaim{
		Name:  dataVolumeName,
		Size:  s.VolumeSize,
		Class: s.VolumeType,
	})
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: s.options.DataDir,
		},
	}

	if s.options.AlertRelabelConfigFile != nil {
		s.options.AlertRelabelConfigFile.Update(ret)
	}

	if s.options.AlertmanagersConfigFile != nil {
		s.options.AlertmanagersConfigFile.Update(ret)
	}

	if s.options.TracingConfigFile != nil {
		s.options.TracingConfigFile.Update(ret)
	}

	if s.options.ObjstoreConfigFile != nil {
		s.options.ObjstoreConfigFile.Update(ret)
	}

	for _, ruleFile := range s.options.RuleFile {
		if ruleFile.VolumeName != "" {
			ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
				Name:      ruleFile.VolumeName,
				MountPath: ruleFile.FilePath(),
				ReadOnly:  true,
			})
		} else if ruleFile.ConfigMapName != "" {
			ret.Volumes = append(ret.Volumes, kghelpers.NewPodVolumeFromConfigMap(ruleFile.ConfigMapName, ruleFile.ConfigMapName))

			ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
				Name:      ruleFile.ConfigMapName,
				MountPath: ruleFile.FilePath(),
			})
		}
	}

	return ret
}
