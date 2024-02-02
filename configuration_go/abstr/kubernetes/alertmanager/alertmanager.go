package alertmanager

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"

	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultClusterPort = 9094
	defaultWebPort     = 9093
	dataVolumeName     = "alertmanager-data"
)

// NewConfigFile returns a new config file option.
func NewConfigFile(value *string) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/alertmanager/config", "config.yaml", "config-file", "alertmanager-config")
	if value != nil {
		ret.WithValue(*value)
	}
	return ret
}

type AlertManagerOptions struct {
	ConfigFile               containeropts.ContainerUpdater `opt:"config.file"`
	StoragePath              string                         `opt:"storage.path"`
	DataRetention            time.Duration                  `opt:"data.retention"`
	DataMaintenanceInterval  time.Duration                  `opt:"data.maintenance-interval"`
	AlertsGCInterval         time.Duration                  `opt:"alerts.gc-interval"`
	WebListenAddress         *net.TCPAddr                   `opt:"web.listen-address"`
	WebExternalURL           string                         `opt:"web.external-url"`
	WebRoutePrefix           string                         `opt:"web.route-prefix"`
	WebGetConcurrency        int                            `opt:"web.get-concurrency"`
	WebTimeout               time.Duration                  `opt:"web.timeout"`
	ClusterListenAddress     string                         `opt:"cluster.listen-address"`
	ClusterPeer              []string                       `opt:"cluster.peer"`
	ClusterPeerTimeout       time.Duration                  `opt:"cluster.peer-timeout"`
	ClusterGossipInterval    time.Duration                  `opt:"cluster.gossip-interval"`
	ClusterPushPullInterval  time.Duration                  `opt:"cluster.pushpull-interval"`
	ClusterTCPTimeout        time.Duration                  `opt:"cluster.tcp-timeout"`
	ClusterProbeTimeout      time.Duration                  `opt:"cluster.probe-timeout"`
	ClusterProbeInterval     time.Duration                  `opt:"cluster.probe-interval"`
	ClusterSettleTimeout     time.Duration                  `opt:"cluster.settle-timeout"`
	ClusterReconnectInterval time.Duration                  `opt:"cluster.reconnect-interval"`
	ClusterReconnectTimeout  time.Duration                  `opt:"cluster.reconnect-timeout"`
	LogLevel                 log.Level                      `opt:"log.level"`
	LogFormat                log.Format                     `opt:"log.format"`
}

type AlertManagerStatefulSet struct {
	options *AlertManagerOptions
	workload.StatefulSetWorkload
}

func NewDefaultOptions() *AlertManagerOptions {
	return &AlertManagerOptions{
		LogLevel:    log.LevelWarn,
		LogFormat:   log.FormatLogfmt,
		StoragePath: "/data",
	}
}

func NewAlertManager(opts *AlertManagerOptions, namespace, imageTag string) *AlertManagerStatefulSet {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "alertmanager",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "alertmanager",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultWebPort, opts.WebListenAddress)

	ssWorkload := workload.StatefulSetWorkload{
		Replicas:   1,
		VolumeSize: "10Gi",
		PodConfig: workload.PodConfig{
			Image:                "quay.io/prometheus/alertmanager",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-alertmanager",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("500m", "1", "500Mi", "4Gi"),
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
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}

	return &AlertManagerStatefulSet{
		options:             opts,
		StatefulSetWorkload: ssWorkload,
	}

}

func (a *AlertManagerStatefulSet) Objects() []runtime.Object {
	container := a.makeContainer()
	ret := a.StatefulSetWorkload.Objects(container)

	service := kghelpers.GetObject[*corev1.Service](ret, a.Name)
	// remove cluster port
	service.Spec.Ports = service.Spec.Ports[:1]

	// Add headless service for cluster port
	if len(a.options.ClusterPeer) > 0 {
		headlessService := &workload.Service{
			MetaConfig:   *a.ObjectMeta(),
			ServicePorts: workload.ServiceProviderFunc(func() []corev1.ServicePort { return container.ServicePorts[1:2] }),
			ClusterIP:    corev1.ClusterIPNone,
		}
		headlessService.MetaConfig.Name = fmt.Sprintf("%s-cluster", headlessService.MetaConfig.Name)
		ret = append(ret, headlessService.Object())

		ss := kghelpers.GetObject[*appsv1.StatefulSet](ret, a.Name)
		ss.Spec.ServiceName = headlessService.MetaConfig.Name
	}

	return ret
}

func (a *AlertManagerStatefulSet) makeContainer() *workload.Container {
	webPort := kghelpers.GetPortOrDefault(defaultWebPort, a.options.WebListenAddress)

	clusterPort := defaultClusterPort
	if a.options.ClusterListenAddress != "" {
		var err error
		clusterPort, err = strconv.Atoi(strings.Split(a.options.ClusterListenAddress, ":")[1])
		if err != nil {
			panic(fmt.Sprintf(`failed to parse cluster listen address %s`, a.options.ClusterListenAddress))
		}
	}

	kghelpers.CheckProbePort(webPort, a.LivenessProbe)
	kghelpers.CheckProbePort(webPort, a.ReadinessProbe)

	if a.options.StoragePath == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := a.ToContainer()
	ret.Name = "alertmanager"
	ret.Args = cmdopt.GetOpts(a.options)
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
		kghelpers.NewServicePort("http", webPort, webPort),
	}
	if len(a.options.ClusterPeer) > 0 {
		ret.ServicePorts = append(ret.ServicePorts, kghelpers.NewServicePort("cluster-tcp", clusterPort, clusterPort))
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "http",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}
	ret.VolumeClaims = append(ret.VolumeClaims, workload.PersistentVolumeClaim{
		Name:  dataVolumeName,
		Size:  a.VolumeSize,
		Class: a.VolumeType,
	})
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      dataVolumeName,
			MountPath: a.options.StoragePath,
		},
	}

	if a.options.ConfigFile != nil {
		a.options.ConfigFile.Update(ret)
	}

	return ret
}
