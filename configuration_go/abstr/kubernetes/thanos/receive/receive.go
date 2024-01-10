package receive

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/reqlogging"
	trclient "github.com/observatorium/observatorium/configuration_go/schemas/thanos/tracing/client"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/relabel"
	corev1 "k8s.io/api/core/v1"
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

type receiveLimitsConfigFile = k8sutil.ConfigFile

// NewReceiveLimitsConfigFile returns a new receive limits config file option.
func NewReceiveLimitsConfigFile(value *ReceiveLimitsConfig) *receiveLimitsConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/receive-limits", "limits.yaml", "receive-limits", "observatorium-thanos-receive-limits")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type receiveHashringConfigFile = k8sutil.ConfigFile

// NewReceiveHashringConfigFile returns a new receive hashring config file option.
func NewReceiveHashringConfigFile(value *HashRingsConfig) *receiveHashringConfigFile {
	ret := k8sutil.NewConfigFile("/etc/thanos/hashring", "hashrings.json", "hashring", "observatorium-thanos-receive-hashring")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// ReceiveOptions represents the options/flags for the receive.
// See https://thanos.io/tip/components/receive.md/#flags for details.
type ReceiveOptions struct {
	GrpcAddress                         *net.TCPAddr               `opt:"grpc-address"`
	GrpcGracePeriod                     time.Duration              `opt:"grpc-grace-period"`
	GrpcServerMaxConnectionAge          time.Duration              `opt:"grpc-server-max-connection-age"`
	GrpcServerTlsCert                   string                     `opt:"grpc-server-tls-cert"`
	GrpcServerTlsClientCa               string                     `opt:"grpc-server-tls-client-ca"`
	GrpcServerTlsKey                    string                     `opt:"grpc-server-tls-key"`
	HashFunc                            string                     `opt:"hash-func"`
	HttpAddress                         *net.TCPAddr               `opt:"http-address"`
	HttpGracePeriod                     time.Duration              `opt:"http-grace-period"`
	HttpConfig                          string                     `opt:"http.config"`
	Label                               []Label                    `opt:"label"`
	LogFormat                           log.LogFormat              `opt:"log.format"`
	LogLevel                            log.LogLevel               `opt:"log.level"`
	ObjstoreConfig                      string                     `opt:"objstore.config"`
	ObjstoreConfigFile                  string                     `opt:"objstore.config-file"`
	ReceiveDefaultTenantID              string                     `opt:"receive.default-tenant-id"`
	ReceiveGrpcCompression              GrpcCompressionType        `opt:"receive.grpc-compression"`
	ReceiveHashringsAlgorithm           string                     `opt:"receive.hashrings-algorithm"`
	ReceiveHashrings                    *HashRingsConfig           `opt:"receive.hashrings"`
	ReceiveHashringsFile                *receiveHashringConfigFile `opt:"receive.hashrings-file"`
	ReceiveHashringsFileRefreshInterval time.Duration              `opt:"receive.hashrings-file-refresh-interval"`
	ReceiveLimitsConfig                 *ReceiveLimitsConfig       `opt:"receive.limits-config"`
	ReceiveLimitsConfigFile             *receiveLimitsConfigFile   `opt:"receive.limits-config-file"`
	ReceiveLocalEndpoint                string                     `opt:"receive.local-endpoint"`
	ReceiveRelabelConfig                *relabel.Config            `opt:"receive.relabel-config"`
	ReceiveRelabelConfigFile            string                     `opt:"receive.relabel-config-file"`
	ReceiveReplicaHeader                string                     `opt:"receive.replica-header"`
	ReceiveReplicationFactor            int                        `opt:"receive.replication-factor"`
	ReceiveTenantCertificateField       string                     `opt:"receive.tenant-certificate-field"`
	ReceiveTenantHeader                 string                     `opt:"receive.tenant-header"`
	ReceiveTenantLabelName              string                     `opt:"receive.tenant-label-name"`
	RemoteWriteAddress                  *net.TCPAddr               `opt:"remote-write.address"`
	RemoteWriteClientServerName         string                     `opt:"remote-write.client-server-name"`
	RemoteWriteClientTlsCa              string                     `opt:"remote-write.client-tls-ca"`
	RemoteWriteClientTlsCert            string                     `opt:"remote-write.client-tls-cert"`
	RemoteWriteClientTlsKey             string                     `opt:"remote-write.client-tls-key"`
	RemoteWriteServerTlsCert            string                     `opt:"remote-write.server-tls-cert"`
	RemoteWriteServerTlsClientCa        string                     `opt:"remote-write.server-tls-client-ca"`
	RemoteWriteServerTlsKey             string                     `opt:"remote-write.server-tls-key"`
	RequestLoggingConfig                *reqlogging.RequestConfig  `opt:"request.logging-config"`
	RequestLoggingConfigFile            string                     `opt:"request.logging-config-file"`
	StoreLimitsRequestSamples           int                        `opt:"store.limits.request-samples"`
	StoreLimitsRequestSeries            int                        `opt:"store.limits.request-series"`
	TracingConfig                       *trclient.TracingConfig    `opt:"tracing.config"`
	TracingConfigFile                   string                     `opt:"tracing.config-file"`
	TsdbAllowOverlappingBlocks          bool                       `opt:"tsdb.allow-overlapping-blocks"`
	TsdbMaxExemplars                    int                        `opt:"tsdb.max-exemplars"`
	TsdbNoLockfile                      bool                       `opt:"tsdb.no-lockfile"`
	TsdbPath                            string                     `opt:"tsdb.path"`
	TsdbRetention                       time.Duration              `opt:"tsdb.retention"`
	TsdbTooFarInFutureTimeWindow        time.Duration              `opt:"tsdb.too-far-in-future.time-window"`
	TsdbWalCompression                  bool                       `opt:"tsdb.wal-compression"`

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
	ro.LogLevel = log.LogLevelWarn
	ro.LogFormat = log.LogFormatLogfmt
	ro.HttpAddress = &net.TCPAddr{Port: defaultHTTPPort, IP: net.ParseIP("0.0.0.0")}
	ro.GrpcAddress = &net.TCPAddr{Port: defaultGRPCPort, IP: net.ParseIP("0.0.0.0")}
	ro.RemoteWriteAddress = &net.TCPAddr{Port: defaultReceivePort, IP: net.ParseIP("0.0.0.0")}

	return ro
}

// Router represents a receive component with router configuration.
// It is deployed as a Deployment.
type Router struct {
	baseReceive
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
		k8sutil.NameLabel:      "thanos-receive-router",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-router",
		k8sutil.VersionLabel:   imageTag,
	}

	baseReceive := newBaseReceive(opts, namespace, imageTag, commonLabels)

	return &Router{
		baseReceive: *baseReceive,
	}
}

// Manifests returns the manifests for the Router.
func (r *Router) Manifests() k8sutil.ObjectMap {
	r.makeContainer(r.withRouterContainer())
	return r.baseReceive.manifests()
}

// Ingestor represents a receive component with ingestor configuration.
// It is deployed as a StatefulSet.
type Ingestor struct {
	baseReceive
	VolumeType string
	VolumeSize string
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
		k8sutil.NameLabel:      "thanos-receive-ingestor",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-ingestor",
		k8sutil.VersionLabel:   imageTag,
	}

	baseReceive := newBaseReceive(opts, namespace, imageTag, commonLabels)
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("OBJSTORE_CONFIG", "objectStore-secret"))
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("POD_NAME", "metadata.name"))

	return &Ingestor{
		baseReceive: *baseReceive,
		VolumeSize:  "50Gi",
	}
}

// Manifests returns the manifests for the Ingestor.
func (i *Ingestor) Manifests() k8sutil.ObjectMap {
	i.makeContainer(i.withIngestorContainer(i.VolumeType, i.VolumeSize))
	return i.baseReceive.manifests()
}

// IngestorRouter represents a receive component with ingestor and router configuration.
// It is deployed as a StatefulSet.
type IngestorRouter struct {
	VolumeType string
	VolumeSize string
	baseReceive
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
		k8sutil.NameLabel:      "thanos-receive-ingestorrouter",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-ingestor-router",
		k8sutil.VersionLabel:   imageTag,
	}

	baseReceive := newBaseReceive(opts, namespace, imageTag, commonLabels)
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("OBJSTORE_CONFIG", "objectStore-secret"))
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("NAME", "metadata.name"))
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"))
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("POD_NAME", "metadata.name"))

	return &IngestorRouter{
		baseReceive: *baseReceive,
		VolumeSize:  "50Gi",
	}
}

// Manifests returns the manifests for the IngestorRouter.
func (ir *IngestorRouter) Manifests() k8sutil.ObjectMap {
	// Set the local endpoint at Manifests time, as it depends on the name of the resource and gRPC port.
	// This option, in addition to the router and receive options, is required to be set for the IngestorRouter.
	ir.options.ReceiveLocalEndpoint = fmt.Sprintf("$(NAME).%s.$(NAMESPACE).svc.cluster.local:%d", ir.Name, ir.options.GrpcAddress.Port)
	ir.makeContainer(
		ir.withIngestorContainer(ir.VolumeType, ir.VolumeSize),
		ir.withRouterContainer(),
	)
	return ir.baseReceive.manifests()
}

// baseReceive is the base struct for all receive components.
// It contains their common configuration.
type baseReceive struct {
	options *ReceiveOptions
	k8sutil.DeploymentGenericConfig
	container *k8sutil.Container
}

func newBaseReceive(opts *ReceiveOptions, namespace, imageTag string, commonLabels map[string]string) *baseReceive {
	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultHTTPPort, opts.HttpAddress)

	return &baseReceive{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/thanos/thanos",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 fmt.Sprintf("%s-%s", commonLabels[k8sutil.InstanceLabel], commonLabels[k8sutil.NameLabel]),
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("1", "2", "10Gi", "20Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,
			LivenessProbe: k8sutil.NewProbe("/-/healthy", probePort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", probePort, k8sutil.ProbeConfig{
				FailureThreshold:    20,
				InitialDelaySeconds: 60,
				PeriodSeconds:       5,
			}),
			TerminationGracePeriodSeconds: 120,
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}
}

func (br *baseReceive) withRouterContainer() ContainerOption {
	return func(container *k8sutil.Container) {
		if br.options.ReceiveHashringsFile == nil {
			panic(`hashrings file is not specified for the statefulset.`)
		}

		br.options.ReceiveHashringsFile.AddToContainer(container)

		if br.options.ReceiveLimitsConfigFile != nil {
			br.options.ReceiveLimitsConfigFile.AddToContainer(container)
		}
	}
}

func (br *baseReceive) withIngestorContainer(volumeType string, volumeSize string) ContainerOption {
	return func(container *k8sutil.Container) {
		if br.options.TsdbPath == "" {
			panic(`data directory is not specified for the statefulset.`)
		}

		container.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      dataVolumeName,
				MountPath: br.options.TsdbPath,
			},
		}
		container.VolumeClaims = []k8sutil.VolumeClaim{
			k8sutil.NewVolumeClaimProvider(dataVolumeName, volumeType, volumeSize),
		}
	}
}

func (br *baseReceive) manifests() k8sutil.ObjectMap {
	container := br.container
	if container == nil {
		panic("container is not initialized")
	}

	ret := k8sutil.ObjectMap{}

	// Create the statefulset or deployment based on the presence of the TSDB path.
	if br.options.TsdbPath != "" {
		ret.AddAll(br.GenerateObjectsStatefulSet(container))
	} else {
		ret.AddAll(br.GenerateObjectsDeployment(container))
	}

	return ret
}

func (br *baseReceive) makeContainer(opts ...ContainerOption) {
	httpPort := k8sutil.GetPortOrDefault(defaultHTTPPort, br.options.HttpAddress)
	k8sutil.CheckProbePort(httpPort, br.LivenessProbe)
	k8sutil.CheckProbePort(httpPort, br.ReadinessProbe)

	grpcPort := k8sutil.GetPortOrDefault(defaultGRPCPort, br.options.GrpcAddress)

	ret := br.ToContainer()
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
		k8sutil.NewServicePort("http", httpPort, httpPort),
		k8sutil.NewServicePort("grpc", grpcPort, grpcPort),
		k8sutil.NewServicePort("remote-write", br.options.RemoteWriteAddress.Port, br.options.RemoteWriteAddress.Port),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	for _, opt := range opts {
		opt(ret)
	}

	br.container = ret
}

type ContainerOption func(*k8sutil.Container)
