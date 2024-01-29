package receive

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	dataVolumeName     string = "data"
	defaultHTTPPort    int    = 10902
	defaultGRPCPort    int    = 10901
	defaultReceivePort int    = 19291
)

type GrpcCompressionType string
type HashRingAlgorithm string

const (
	GrpcCompressionSnappy    GrpcCompressionType = "snappy"
	GrpcCompressionNone      GrpcCompressionType = "none"
	HashRingAlgorithmHashmod HashRingAlgorithm   = "hashmod"
	HashRingAlgorithmKetama  HashRingAlgorithm   = "ketama"
)

// Label represents a single label configuration.
type Label struct {
	Key   string
	Value string
}

func (l Label) String() string {
	return l.Key + "=" + l.Value
}

// HashringConfig represents a single hashring configuration.
type HashringConfig struct {
	Hashring       string            `json:"hashring,omitempty"`
	Tenants        []string          `json:"tenants,omitempty"`
	Endpoints      []Endpoint        `json:"endpoints,omitempty"`
	Algorithm      HashRingAlgorithm `json:"algorithm,omitempty"`
	ExternalLabels map[string]string `json:"external_labels,omitempty"`
}

// Endpoint represents a endpoint configuration in a hashring.
type Endpoint struct {
	Address string `json:"address,omitempty"`
	AZ      string `json:"az,omitempty"`
}

// HashRingsConfig represents the hashring configuration.
type HashRingsConfig []HashringConfig

// String returns a string representation of the HashRingsConfig as JSON.
// It implements the Stringer interface that is used by the cmdopt package.
func (h HashRingsConfig) String() string {
	ret, err := json.Marshal(h)
	if err != nil {
		panic(err)
	}

	return string(ret)
}

// NewReceiveLimitsConfigFile returns a new receive limits config file option.
func NewReceiveLimitsConfigFile(value *ReceiveLimitsConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/receive-limits", "limits.yaml", "receive-limits", "observatorium-thanos-receive-limits")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewReceiveHashringConfigFile returns a new receive hashring config file option.
func NewReceiveHashringConfigFile(value *HashRingsConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos/hashring", "hashrings.json", "hashring", "observatorium-thanos-receive-hashring")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// ReceiveOptions represents the options/flags for the receive.
// See https://thanos.io/tip/components/receive.md/#flags for details.
type ReceiveOptions struct {
	GrpcAddress                         *net.TCPAddr                   `opt:"grpc-address"`
	GrpcGracePeriod                     time.Duration                  `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge          time.Duration                  `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert                   string                         `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa               string                         `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey                    string                         `opt:"grpc-server-tls-key"`
	HashFunc                            string                         `opt:"hash-func"`
	HttpAddress                         *net.TCPAddr                   `opt:"http-address"`
	HttpGracePeriod                     time.Duration                  `opt:"http-grace-period"`
	HttpConfig                          string                         `opt:"http.config"`
	Label                               []Label                        `opt:"label"`
	LogFormat                           log.Format                     `opt:"log.format"`
	LogLevel                            log.Level                      `opt:"log.level"`
	ObjstoreConfig                      string                         `opt:"objstore.config"`
	ObjstoreConfigFile                  string                         `opt:"objstore.config-file"`
	ReceiveDefaultTenantID              string                         `opt:"receive.default-tenant-id"`
	ReceiveGrpcCompression              GrpcCompressionType            `opt:"receive.grpc-compression"`
	ReceiveHashringsAlgorithm           string                         `opt:"receive.hashrings-algorithm"`
	ReceiveHashrings                    *HashRingsConfig               `opt:"receive.hashrings"`
	ReceiveHashringsFile                containeropts.ContainerUpdater `opt:"receive.hashrings-file"`
	ReceiveHashringsFileRefreshInterval time.Duration                  `opt:"receive.hashrings-file-refresh-interval"`
	ReceiveLimitsConfig                 *ReceiveLimitsConfig           `opt:"receive.limits-config"`
	ReceiveLimitsConfigFile             containeropts.ContainerUpdater `opt:"receive.limits-config-file"`
	ReceiveLocalEndpoint                string                         `opt:"receive.local-endpoint"`
	ReceiveRelabelConfig                *relabel.Config                `opt:"receive.relabel-config"`
	ReceiveRelabelConfigFile            string                         `opt:"receive.relabel-config-file"`
	ReceiveReplicaHeader                string                         `opt:"receive.replica-header"`
	ReceiveReplicationFactor            int                            `opt:"receive.replication-factor"`
	ReceiveTenantCertificateField       string                         `opt:"receive.tenant-certificate-field"`
	ReceiveTenantHeader                 string                         `opt:"receive.tenant-header"`
	ReceiveTenantLabelName              string                         `opt:"receive.tenant-label-name"`
	RemoteWriteAddress                  *net.TCPAddr                   `opt:"remote-write.address"`
	RemoteWriteClientServerName         string                         `opt:"remote-write.client-server-name"`
	RemoteWriteClientTlsCa              string                         `opt:"remote-write.client-tls-ca"`
	RemoteWriteClientTlsCert            string                         `opt:"remote-write.client-tls-cert"`
	RemoteWriteClientTlsKey             string                         `opt:"remote-write.client-tls-key"`
	RemoteWriteServerTlsCert            string                         `opt:"remote-write.server-tls-cert"`
	RemoteWriteServerTlsClientCa        string                         `opt:"remote-write.server-tls-client-ca"`
	RemoteWriteServerTlsKey             string                         `opt:"remote-write.server-tls-key"`
	RequestLoggingConfig                *reqlogging.RequestConfig      `opt:"request.logging-config"`
	RequestLoggingConfigFile            string                         `opt:"request.logging-config-file"`
	StoreLimitsRequestSamples           int                            `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries            int                            `opt:"store.limits.request-series"`
	TracingConfig                       *trclient.TracingConfig        `opt:"tracing.config"`
	TracingConfigFile                   string                         `opt:"tracing.config-file"`
	TsdbAllowOverlappingBlocks          bool                           `opt:"tsdb.allow-overlapping-blocks"`
	TsdbMaxExemplars                    int                            `opt:"tsdb.max-exemplars"`
	TsdbNoLockfile                      bool                           `opt:"tsdb.no-lockfile"`
	TsdbPath                            string                         `opt:"tsdb.path"`
	TsdbRetention                       time.Duration                  `opt:"tsdb.retention"`
	TsdbTooFarInFutureTimeWindow        time.Duration                  `opt:"tsdb.too-far-in-future.time-window"`
	TsdbWalCompression                  bool                           `opt:"tsdb.wal-compression"`

	// Extra options not officially supported by the receive.
	cmdopt.ExtraOpts
}

func (ro *ReceiveOptions) withDefaultRouterOptions() *ReceiveOptions {
	ro.ReceiveHashringsFile = NewReceiveHashringConfigFile(nil)
	ro.ReceiveHashringsFileRefreshInterval = 5 * time.Second
	ro.ReceiveHashringsAlgorithm = "ketama"
	ro.Label = append(ro.Label, Label{Key: "receive", Value: "\"true\""})

	return ro
}

func (ro *ReceiveOptions) withDefaultIngestorOptions() *ReceiveOptions {
	ro.TsdbPath = "/var/thanos/receive"
	ro.Label = append(ro.Label, Label{Key: "replica", Value: "\"$(POD_NAME)\""})
	ro.ObjstoreConfig = "$(OBJSTORE_CONFIG)"

	return ro
}

func (ro *ReceiveOptions) withBaseOptions() *ReceiveOptions {
	ro.LogLevel = log.LevelWarn
	ro.LogFormat = log.FormatLogfmt
	ro.HttpAddress = &net.TCPAddr{Port: defaultHTTPPort, IP: net.ParseIP("0.0.0.0")}
	ro.GrpcAddress = &net.TCPAddr{Port: defaultGRPCPort, IP: net.ParseIP("0.0.0.0")}
	ro.RemoteWriteAddress = &net.TCPAddr{Port: defaultReceivePort, IP: net.ParseIP("0.0.0.0")}

	return ro
}

// Router represents a receive component with router configuration.
// It is deployed as a Deployment.
type Router struct {
	baseReceive
	workload.DeploymentWorkload
}

func NewDefaultRouterOptions() *ReceiveOptions {
	ret := &ReceiveOptions{}
	return ret.withBaseOptions().withDefaultRouterOptions()
}

// NewRouter returns a new Router with default configuration.
func NewRouter(opts *ReceiveOptions, namespace, imageTag string) *Router {
	if opts == nil {
		opts = NewDefaultRouterOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-receive-router",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "database-write-hashring-router",
		workload.VersionLabel:   imageTag,
	}

	baseReceive, podConfig := newBaseReceive(opts, namespace, imageTag, commonLabels)

	depWorkload := workload.DeploymentWorkload{
		Replicas:  1,
		PodConfig: podConfig,
	}

	return &Router{
		baseReceive:        *baseReceive,
		DeploymentWorkload: depWorkload,
	}
}

// Manifests returns the manifests for the Router.
func (r *Router) Objects() []runtime.Object {
	container := r.makeContainer(&r.PodConfig, r.withRouterContainer())
	return r.DeploymentWorkload.Objects(container)
}

// Ingestor represents a receive component with ingestor configuration.
// It is deployed as a StatefulSet.
type Ingestor struct {
	baseReceive
	workload.StatefulSetWorkload
}

func NewDefaultIngestorOptions() *ReceiveOptions {
	ret := &ReceiveOptions{}
	return ret.withBaseOptions().withDefaultIngestorOptions()
}

// NewIngestor returns a new Ingestor with default configuration.
func NewIngestor(opts *ReceiveOptions, namespace, imageTag string) *Ingestor {
	if opts == nil {
		opts = NewDefaultIngestorOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-receive-ingestor",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "database-write-hashring-ingestor",
		workload.VersionLabel:   imageTag,
	}

	baseReceive, podConfig := newBaseReceive(opts, namespace, imageTag, commonLabels)
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("OBJSTORE_CONFIG", "objectStore-secret"))
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("POD_NAME", "metadata.name"))

	ssWorkload := workload.StatefulSetWorkload{
		Replicas:   1,
		VolumeSize: "50Gi",
		PodConfig:  podConfig,
	}

	return &Ingestor{
		baseReceive:         *baseReceive,
		StatefulSetWorkload: ssWorkload,
	}
}

// Manifests returns the manifests for the Ingestor.
func (i *Ingestor) Objects() []runtime.Object {
	container := i.makeContainer(&i.PodConfig, i.withIngestorContainer(i.VolumeType, i.VolumeSize))
	return i.StatefulSetWorkload.Objects(container)
}

// IngestorRouter represents a receive component with ingestor and router configuration.
// It is deployed as a StatefulSet.
type IngestorRouter struct {
	baseReceive
	workload.StatefulSetWorkload
}

func NewDefaultIngestorRouterOptions() *ReceiveOptions {
	ret := &ReceiveOptions{}
	return ret.withBaseOptions().
		withDefaultRouterOptions().
		withDefaultIngestorOptions()
}

// NewIngestorRouter returns a new IngestorRouter with default configuration.
func NewIngestorRouter(opts *ReceiveOptions, namespace, imageTag string) *IngestorRouter {
	if opts == nil {
		opts = NewDefaultIngestorRouterOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "thanos-receive-ingestorrouter",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "database-write-hashring-ingestor-router",
		workload.VersionLabel:   imageTag,
	}

	baseReceive, podConfig := newBaseReceive(opts, namespace, imageTag, commonLabels)
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("OBJSTORE_CONFIG", "objectStore-secret"))
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("NAME", "metadata.name"))
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("NAMESPACE", "metadata.namespace"))
	podConfig.Env = append(podConfig.Env, kghelpers.NewEnvFromField("POD_NAME", "metadata.name"))

	return &IngestorRouter{
		baseReceive: *baseReceive,
		StatefulSetWorkload: workload.StatefulSetWorkload{
			VolumeSize: "50Gi",
		},
	}
}

// Manifests returns the manifests for the IngestorRouter.
func (ir *IngestorRouter) Objects() []runtime.Object {
	// Set the local endpoint at Manifests time, as it depends on the name of the resource and gRPC port.
	// This option, in addition to the router and receive options, is required to be set for the IngestorRouter.
	ir.options.ReceiveLocalEndpoint = fmt.Sprintf("$(NAME).%s.$(NAMESPACE).svc.cluster.local:%d", ir.Name, ir.options.GrpcAddress.Port)
	container := ir.makeContainer(
		&ir.PodConfig,
		ir.withIngestorContainer(ir.VolumeType, ir.VolumeSize),
		ir.withRouterContainer(),
	)
	return ir.StatefulSetWorkload.Objects(container)
}

// baseReceive is the base struct for all receive components.
// It contains their common configuration.
type baseReceive struct {
	options *ReceiveOptions
}

func newBaseReceive(opts *ReceiveOptions, namespace, imageTag string, commonLabels map[string]string) (*baseReceive, workload.PodConfig) {
	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	podCfg := workload.PodConfig{
		Image:                "quay.io/thanos/thanos",
		ImageTag:             imageTag,
		ImagePullPolicy:      corev1.PullIfNotPresent,
		Name:                 fmt.Sprintf("%s-%s", commonLabels[workload.InstanceLabel], commonLabels[workload.NameLabel]),
		Namespace:            namespace,
		CommonLabels:         commonLabels,
		ContainerResources:   kghelpers.NewResourcesRequirements("1", "2", "10Gi", "20Gi"),
		Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
		EnableServiceMonitor: true,
		LivenessProbe: kghelpers.NewProbe("/-/healthy", probePort, kghelpers.ProbeConfig{
			FailureThreshold: 8,
			PeriodSeconds:    30,
			TimeoutSeconds:   1,
		}),
		ReadinessProbe: kghelpers.NewProbe("/-/ready", probePort, kghelpers.ProbeConfig{
			FailureThreshold:    20,
			InitialDelaySeconds: 60,
			PeriodSeconds:       5,
		}),
		TerminationGracePeriodSeconds: 120,
		ConfigMaps:                    make(map[string]map[string]string),
		Secrets:                       make(map[string]map[string][]byte),
	}

	return &baseReceive{
		options: opts,
	}, podCfg
}

func (br *baseReceive) withRouterContainer() ContainerOption {
	return func(container *workload.Container) {
		if br.options.ReceiveHashringsFile == nil {
			panic(`hashrings file is not specified for the statefulset.`)
		}

		br.options.ReceiveHashringsFile.Update(container)

		if br.options.ReceiveLimitsConfigFile != nil {
			br.options.ReceiveLimitsConfigFile.Update(container)
		}
	}
}

func (br *baseReceive) withIngestorContainer(volumeType string, volumeSize string) ContainerOption {
	return func(container *workload.Container) {
		if br.options.TsdbPath == "" {
			panic(`data directory is not specified for the statefulset.`)
		}

		container.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      dataVolumeName,
				MountPath: br.options.TsdbPath,
			},
		}
		container.VolumeClaims = append(container.VolumeClaims, workload.PersistentVolumeClaim{
			Name:  dataVolumeName,
			Size:  volumeSize,
			Class: volumeType,
		})
	}
}

func (br *baseReceive) makeContainer(podConfig *workload.PodConfig, opts ...ContainerOption) *workload.Container {
	httpPort := kghelpers.GetPortOrDefault(defaultHTTPPort, br.options.HttpAddress)
	kghelpers.CheckProbePort(httpPort, podConfig.LivenessProbe)
	kghelpers.CheckProbePort(httpPort, podConfig.ReadinessProbe)

	grpcPort := kghelpers.GetPortOrDefault(defaultGRPCPort, br.options.GrpcAddress)

	ret := podConfig.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"receive"}, cmdopt.GetOpts(br.options)...)
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
		{
			Name:          "remote-write",
			ContainerPort: int32(br.options.RemoteWriteAddress.Port),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http", httpPort, httpPort),
		kghelpers.NewServicePort("grpc", grpcPort, grpcPort),
		kghelpers.NewServicePort("remote-write", br.options.RemoteWriteAddress.Port, br.options.RemoteWriteAddress.Port),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}

type ContainerOption func(*workload.Container)
