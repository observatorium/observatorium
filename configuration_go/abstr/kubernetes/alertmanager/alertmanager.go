package alertmanager

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultClusterPort = 9094
	defaultWebPort     = 9093
	dataVolumeName     = "alertmanager-data"
)

// NewConfigFile returns a new config file option.
func NewConfigFile(value *string) *k8sutil.ConfigFile {
	ret := k8sutil.NewConfigFile("/etc/alertmanager/config", "config.yaml", "config-file", "alertmanager-config")
	if value != nil {
		ret.WithValue(*value)
	}
	return ret
}

type AlertManagerOptions struct {
	ConfigFile               k8sutil.ContainerUpdater `opt:"config.file"`
	StoragePath              string                   `opt:"storage.path"`
	DataRetention            time.Duration            `opt:"data.retention"`
	DataMaintenanceInterval  time.Duration            `opt:"data.maintenance-interval"`
	AlertsGCInterval         time.Duration            `opt:"alerts.gc-interval"`
	WebListenAddress         *net.TCPAddr             `opt:"web.listen-address"`
	WebExternalURL           string                   `opt:"web.external-url"`
	WebRoutePrefix           string                   `opt:"web.route-prefix"`
	WebGetConcurrency        int                      `opt:"web.get-concurrency"`
	WebTimeout               time.Duration            `opt:"web.timeout"`
	ClusterListenAddress     string                   `opt:"cluster.listen-address"`
	ClusterPeer              []string                 `opt:"cluster.peer"`
	ClusterPeerTimeout       time.Duration            `opt:"cluster.peer-timeout"`
	ClusterGossipInterval    time.Duration            `opt:"cluster.gossip-interval"`
	ClusterPushPullInterval  time.Duration            `opt:"cluster.pushpull-interval"`
	ClusterTCPTimeout        time.Duration            `opt:"cluster.tcp-timeout"`
	ClusterProbeTimeout      time.Duration            `opt:"cluster.probe-timeout"`
	ClusterProbeInterval     time.Duration            `opt:"cluster.probe-interval"`
	ClusterSettleTimeout     time.Duration            `opt:"cluster.settle-timeout"`
	ClusterReconnectInterval time.Duration            `opt:"cluster.reconnect-interval"`
	ClusterReconnectTimeout  time.Duration            `opt:"cluster.reconnect-timeout"`
	LogLevel                 log.LogLevel             `opt:"log.level"`
	LogFormat                log.LogFormat            `opt:"log.format"`
}

type AlertManagerStatefulSet struct {
	options    *AlertManagerOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewDefaultOptions() *AlertManagerOptions {
	return &AlertManagerOptions{
		LogLevel:    log.LogLevelWarn,
		LogFormat:   log.LogFormatLogfmt,
		StoragePath: "/data",
	}
}

func NewAlertManager(opts *AlertManagerOptions, namespace, imageTag string) *AlertManagerStatefulSet {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "alertmanager",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "alertmanager",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultWebPort, opts.WebListenAddress)

	return &AlertManagerStatefulSet{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/prometheus/alertmanager",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-alertmanager",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("500m", "1", "500Mi", "4Gi"),
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
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
		VolumeSize: "10Gi",
	}
}

func (s *AlertManagerStatefulSet) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()

	ret := k8sutil.ObjectMap{}

	ret.AddAll(s.GenerateObjectsStatefulSet(container))
	service := k8sutil.GetObject[*corev1.Service](ret, s.Name)
	// remove cluster port
	service.Spec.Ports = service.Spec.Ports[:1]

	// Add headless service for cluster port
	if len(s.options.ClusterPeer) > 0 {
		headlessService := &k8sutil.Service{
			MetaConfig:   *s.ObjectMeta(),
			ServicePorts: k8sutil.ServiceProviderFunc(func() []corev1.ServicePort { return container.ServicePorts[1:2] }),
			ClusterIP:    corev1.ClusterIPNone,
		}
		headlessService.MetaConfig.Name = fmt.Sprintf("%s-cluster", headlessService.MetaConfig.Name)
		ret.Add(headlessService.MakeManifest())

		ss := k8sutil.GetObject[*appsv1.StatefulSet](ret, s.Name)
		ss.Spec.ServiceName = headlessService.MetaConfig.Name
	}

	return ret
}

func (s *AlertManagerStatefulSet) makeContainer() *k8sutil.Container {
	webPort := k8sutil.GetPortOrDefault(defaultWebPort, s.options.WebListenAddress)

	clusterPort := defaultClusterPort
	if s.options.ClusterListenAddress != "" {
		var err error
		clusterPort, err = strconv.Atoi(strings.Split(s.options.ClusterListenAddress, ":")[1])
		if err != nil {
			panic(fmt.Sprintf(`failed to parse cluster listen address %s`, s.options.ClusterListenAddress))
		}
	}

	k8sutil.CheckProbePort(webPort, s.LivenessProbe)
	k8sutil.CheckProbePort(webPort, s.ReadinessProbe)

	if s.options.StoragePath == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "alertmanager"
	ret.Args = cmdopt.GetOpts(s.options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(webPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "cluster-tcp",
			ContainerPort: int32(clusterPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("http", webPort, webPort),
	}
	if len(s.options.ClusterPeer) > 0 {
		ret.ServicePorts = append(ret.ServicePorts, k8sutil.NewServicePort("cluster-tcp", clusterPort, clusterPort))
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
			MountPath: s.options.StoragePath,
		},
	}

	if s.options.ConfigFile != nil {
		s.options.ConfigFile.Update(ret)
	}

	return ret
}
