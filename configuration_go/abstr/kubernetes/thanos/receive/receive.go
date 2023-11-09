package receive

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/log"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/option"
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
	defaultNamespace   string = "observatorium"
	defaultImage       string = "quay.io/thanos/thanos"
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

type receiveLimitsConfigFile = option.ConfigFile[ReceiveLimitsConfig]

// NewReceiveLimitsConfigFile returns a new receive limits config file option.
func NewReceiveLimitsConfigFile(name string, value ReceiveLimitsConfig) *receiveLimitsConfigFile {
	return option.NewConfigFile("/etc/thanos/receive-limits", "limits.yaml", name, value)
}

type receiveHashringConfigFile = option.ConfigFile[HashRingsConfig]

// NewReceiveHashringConfigFile returns a new receive hashring config file option.
func NewReceiveHashringConfigFile(name string, value HashRingsConfig) *receiveHashringConfigFile {
	return option.NewConfigFile("/etc/thanos/hashring", "hashrings.json", name, value) //
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

// Router represents a receive component with router configuration.
// It is deployed as a Deployment.
type Router struct {
	baseReceive
}

// NewRouter returns a new Router with default configuration.
func NewRouter() *Router {
	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-receive-router",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-router",
	}

	baseReceive := newBaseReceive(commonLabels)
	baseReceive.withRouterConfig()

	return &Router{
		baseReceive: *baseReceive,
	}
}

// Manifests returns the manifests for the Router.
func (r *Router) Manifests() k8sutil.ObjectMap {
	r.withRouterContainer()
	return r.baseReceive.manifests()
}

// Ingestor represents a receive component with ingestor configuration.
// It is deployed as a StatefulSet.
type Ingestor struct {
	baseReceive
	VolumeType string
	VolumeSize string
}

// NewIngestor returns a new Ingestor with default configuration.
func NewIngestor() *Ingestor {
	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-receive-ingestor",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-ingestor",
	}

	baseReceive := newBaseReceive(commonLabels)
	baseReceive.withIngestorConfig()

	return &Ingestor{
		baseReceive: *baseReceive,
		VolumeSize:  "100Gi",
	}
}

// Manifests returns the manifests for the Ingestor.
func (i *Ingestor) Manifests() k8sutil.ObjectMap {
	i.withIngestorContainer(i.VolumeType, i.VolumeSize)
	return i.baseReceive.manifests()
}

// IngestorRouter represents a receive component with ingestor and router configuration.
// It is deployed as a StatefulSet.
type IngestorRouter struct {
	VolumeType string
	VolumeSize string
	baseReceive
}

// NewIngestorRouter returns a new IngestorRouter with default configuration.
func NewIngestorRouter() *IngestorRouter {
	commonLabels := map[string]string{
		k8sutil.NameLabel:      "thanos-receive-ingestorrouter",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "database-write-hashring-ingestor-router",
	}

	baseReceive := newBaseReceive(commonLabels)
	baseReceive.withIngestorConfig()
	baseReceive.withRouterConfig()
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("NAME", "metadata.name"))
	baseReceive.Env = append(baseReceive.Env, k8sutil.NewEnvFromField("NAMESPACE", "metadata.namespace"))

	return &IngestorRouter{
		baseReceive: *baseReceive,
		VolumeSize:  "100Gi",
	}
}

// Manifests returns the manifests for the IngestorRouter.
func (ir *IngestorRouter) Manifests() k8sutil.ObjectMap {
	// Set the local endpoint at Manifests time, as it depends on the name of the resource and gRPC port.
	// This option, in addition to the router and receive options, is required to be set for the IngestorRouter.
	ir.Options.ReceiveLocalEndpoint = fmt.Sprintf("$(NAME).%s.$(NAMESPACE).svc.cluster.local:%d", ir.Name, ir.Options.GrpcAddress.Port)
	ir.withIngestorContainer(ir.VolumeType, ir.VolumeSize)
	ir.withRouterContainer()
	return ir.baseReceive.manifests()
}

// baseReceive is the base struct for all receive components.
// It contains their common configuration.
type baseReceive struct {
	Options *ReceiveOptions
	k8sutil.DeploymentGenericConfig
	container *k8sutil.Container
}

func newBaseReceive(commonLabels map[string]string) *baseReceive {
	opts := &ReceiveOptions{
		LogLevel:           log.LogLevelWarn,
		LogFormat:          log.LogFormatLogfmt,
		HttpAddress:        &net.TCPAddr{Port: defaultHTTPPort, IP: net.ParseIP("0.0.0.0")},
		GrpcAddress:        &net.TCPAddr{Port: defaultGRPCPort, IP: net.ParseIP("0.0.0.0")},
		RemoteWriteAddress: &net.TCPAddr{Port: defaultReceivePort, IP: net.ParseIP("0.0.0.0")},
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &baseReceive{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                defaultImage,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 fmt.Sprintf("%s-%s", commonLabels[k8sutil.InstanceLabel], commonLabels[k8sutil.NameLabel]),
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("1", "2", "10Gi", "20Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,
			LivenessProbe: k8sutil.NewProbe("/-/healthy", defaultHTTPPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", defaultHTTPPort, k8sutil.ProbeConfig{
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

func (br *baseReceive) withRouterConfig() {
	br.Options.ReceiveHashringsFile = NewReceiveHashringConfigFile(br.Name+"-hashring", HashRingsConfig{})
	br.Options.ReceiveHashringsFileRefreshInterval = 5 * time.Second
	br.Options.ReceiveHashringsAlgorithm = "ketama"
	br.Options.Label = []Label{{Key: "receive", Value: "\"true\""}}
}

func (br *baseReceive) withRouterContainer() {
	if br.Options.ReceiveHashringsFile == nil {
		panic(`hashrings file is not specified for the statefulset.`)
	}

	container := br.makeContainer()

	// The configmap can be dyamically generated by the controller
	// We only create the config map if it is defined in the options.
	if len(br.Options.ReceiveHashringsFile.Value) > 0 {
		container.ConfigMaps[br.Options.ReceiveHashringsFile.Name] = map[string]string{
			br.Options.ReceiveHashringsFile.FileName(): br.Options.ReceiveHashringsFile.Value.String(),
		}
	}

	container.Volumes = append(container.Volumes, k8sutil.NewPodVolumeFromConfigMap("hashring-config", br.Options.ReceiveHashringsFile.Name))
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "hashring-config",
		MountPath: br.Options.ReceiveHashringsFile.MountPath(),
	})

	if br.Options.ReceiveLimitsConfigFile != nil {
		container.ConfigMaps[br.Options.ReceiveLimitsConfigFile.Name] = map[string]string{
			br.Options.ReceiveLimitsConfigFile.FileName(): br.Options.ReceiveLimitsConfigFile.Value.String(),
		}

		container.Volumes = append(container.Volumes, k8sutil.NewPodVolumeFromConfigMap("receive-limits-config", br.Options.ReceiveLimitsConfigFile.Name))
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "receive-limits-config",
			MountPath: br.Options.ReceiveLimitsConfigFile.MountPath(),
		})
	}

	br.container = container
}

func (br *baseReceive) withIngestorConfig() {
	br.Options.TsdbPath = "/var/thanos/receive"
	br.Options.Label = []Label{{Key: "replica", Value: "\"$(POD_NAME)\""}}
	br.Options.ObjstoreConfig = "$(OBJSTORE_CONFIG)"

	br.Env = append(br.Env, k8sutil.NewEnvFromField("OBJSTORE_CONFIG", "objectStore-secret"))
}

func (br *baseReceive) withIngestorContainer(volumeType string, volumeSize string) {
	if br.Options.TsdbPath == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	container := br.makeContainer()
	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: br.Options.TsdbPath,
		},
	}
	container.VolumeClaims = []k8sutil.VolumeClaim{
		k8sutil.NewVolumeClaimProvider(dataVolumeName, volumeType, volumeSize),
	}
	container.Env = append(container.Env, k8sutil.NewEnvFromField("POD_NAME", "metadata.name"))
	br.container = container
}

func (br *baseReceive) manifests() k8sutil.ObjectMap {
	container := br.container
	if container == nil {
		panic(`container is not initialized`)
	}

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      br.Name,
		Labels:    br.CommonLabels,
		Namespace: br.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &br.TerminationGracePeriodSeconds,
		Affinity:                      br.Affinity,
		SecurityContext:               br.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container}, br.Sidecars...),
	}

	ret := k8sutil.ObjectMap{}
	// Create the statefulset or deployment based on the presence of the TSDB path.
	if br.Options.TsdbPath != "" {
		statefulset := &k8sutil.StatefulSet{
			MetaConfig: commonObjectMeta.Clone(),
			Replicas:   br.Replicas,
			Pod:        pod,
		}

		ret["receive-statefulSet"] = statefulset.MakeManifest()
	} else {
		deployment := &k8sutil.Deployment{
			MetaConfig: commonObjectMeta.Clone(),
			Replicas:   br.Replicas,
			Pod:        pod,
		}

		ret["receive-deployment"] = deployment.MakeManifest()
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["receive-service"] = service.MakeManifest()

	if br.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["receive-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["receive-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["receive-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["receive-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (br *baseReceive) makeContainer() *k8sutil.Container {
	if br.Options == nil {
		br.Options = &ReceiveOptions{}
	}

	httpPort := defaultHTTPPort
	if br.Options.HttpAddress != nil && br.Options.HttpAddress.Port != 0 {
		httpPort = br.Options.HttpAddress.Port
	}

	grpcPort := defaultGRPCPort
	if br.Options.GrpcAddress != nil && br.Options.GrpcAddress.Port != 0 {
		grpcPort = br.Options.GrpcAddress.Port
	}

	livenessPort := br.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(httpPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, httpPort))
	}

	readinessPort := br.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(httpPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, httpPort))
	}

	ret := br.ToContainer()
	ret.Name = "thanos"
	ret.Args = append([]string{"receive"}, cmdopt.GetOpts(br.Options)...)
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
			ContainerPort: int32(br.Options.RemoteWriteAddress.Port),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("http", httpPort, httpPort),
		k8sutil.NewServicePort("grpc", grpcPort, grpcPort),
		k8sutil.NewServicePort("remote-write", br.Options.RemoteWriteAddress.Port, br.Options.RemoteWriteAddress.Port),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	return ret
}
