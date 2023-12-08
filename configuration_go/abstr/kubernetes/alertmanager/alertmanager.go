package alertmanager

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	thanoslog "github.com/observatorium/observatorium/configuration_go/schemas/thanos/log"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultClusterPort = 9094
	defaultWebPort     = 9093
	dataVolumeName     = "alertmanager-data"
	defaultNamespace   = "observatorium"
)

type AlertmanagersConfigFile struct {
	SecretName    string
	ConfigMapName string
	FileName      string
}

func (a AlertmanagersConfigFile) String() string {
	return filepath.Join("/etc/alertmanager/config", a.FileName)
}

func (a AlertmanagersConfigFile) MountPath() string {
	return "/etc/alertmanager/config"
}

type AlertManagerOptions struct {
	ConfigFile               *AlertmanagersConfigFile `opt:"config.file"`
	StoragePath              string                   `opt:"storage.path"`
	DataRetention            model.Duration           `opt:"data.retention"`
	DataMaintenanceInterval  model.Duration           `opt:"data.maintenance-interval"`
	AlertsGCInterval         model.Duration           `opt:"alerts.gc-interval"`
	WebListenAddress         *net.TCPAddr             `opt:"web.listen-address"`
	WebExternalURL           string                   `opt:"web.external-url"`
	WebRoutePrefix           string                   `opt:"web.route-prefix"`
	WebGetConcurrency        int                      `opt:"web.get-concurrency"`
	WebTimeout               model.Duration           `opt:"web.timeout"`
	ClusterListenAddress     string                   `opt:"cluster.listen-address"`
	ClusterPeer              []string                 `opt:"cluster.peer"`
	ClusterPeerTimeout       model.Duration           `opt:"cluster.peer-timeout"`
	ClusterGossipInterval    model.Duration           `opt:"cluster.gossip-interval"`
	ClusterPushPullInterval  model.Duration           `opt:"cluster.pushpull-interval"`
	ClusterTCPTimeout        model.Duration           `opt:"cluster.tcp-timeout"`
	ClusterProbeTimeout      model.Duration           `opt:"cluster.probe-timeout"`
	ClusterProbeInterval     model.Duration           `opt:"cluster.probe-interval"`
	ClusterSettleTimeout     model.Duration           `opt:"cluster.settle-timeout"`
	ClusterReconnectInterval model.Duration           `opt:"cluster.reconnect-interval"`
	ClusterReconnectTimeout  model.Duration           `opt:"cluster.reconnect-timeout"`
	LogLevel                 thanoslog.LogLevel       `opt:"log.level"`
	LogFormat                thanoslog.LogFormat      `opt:"log.format"`
}

type AlertManagerStatefulSet struct {
	Options    *AlertManagerOptions
	VolumeType string
	VolumeSize string

	k8sutil.DeploymentGenericConfig
}

func NewAlertManager() *AlertManagerStatefulSet {
	opts := &AlertManagerOptions{
		LogLevel:    "warn",
		LogFormat:   "logfmt",
		StoragePath: "/data",
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "alertmanager",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "alertmanager",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &AlertManagerStatefulSet{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/prometheus/alertmanager",
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-alertmanager",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("500m", "1", "500Mi", "4Gi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/-/healthy", defaultWebPort, k8sutil.ProbeConfig{
				FailureThreshold: 8,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/-/ready", defaultWebPort, k8sutil.ProbeConfig{
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
		"alertmanager-statefulSet": statefulset.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: k8sutil.ServiceProviderFunc(func() []corev1.ServicePort { return container.ServicePorts[:1] }),
	}
	ret["alertmanager-service"] = service.MakeManifest()

	if len(s.Options.ClusterPeer) > 0 {
		headlessService := &k8sutil.Service{
			MetaConfig:   commonObjectMeta.Clone(),
			ServicePorts: k8sutil.ServiceProviderFunc(func() []corev1.ServicePort { return container.ServicePorts[1:2] }),
			ClusterIP:    corev1.ClusterIPNone,
		}
		headlessService.MetaConfig.Name = fmt.Sprintf("%s-cluster", headlessService.MetaConfig.Name)
		ret["alertmanager-headless-service"] = headlessService.MakeManifest()
	}

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["alertmanager-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["alertmanager-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["alertmanager-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["alertmanager-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *AlertManagerStatefulSet) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &AlertManagerOptions{}
	}

	webPort := defaultWebPort
	if s.Options.WebListenAddress != nil && s.Options.WebListenAddress.Port != 0 {
		webPort = s.Options.WebListenAddress.Port
	}

	clusterPort := defaultClusterPort
	if s.Options.ClusterListenAddress != "" {
		var err error
		clusterPort, err = strconv.Atoi(strings.Split(s.Options.ClusterListenAddress, ":")[1])
		if err != nil {
			panic(fmt.Sprintf(`failed to parse cluster listen address %s`, s.Options.ClusterListenAddress))
		}
	}

	livenessPort := s.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(webPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, webPort))
	}

	readinessPort := s.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(webPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, webPort))
	}

	if s.Options.StoragePath == "" {
		panic(`data directory is not specified for the statefulset.`)
	}

	ret := s.ToContainer()
	ret.Name = "alertmanager"
	ret.Args = cmdopt.GetOpts(s.Options)
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
		// k8sutil.NewServicePort("cluster-tcp", clusterPort, clusterPort),
	}
	if len(s.Options.ClusterPeer) > 0 {
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
			MountPath: s.Options.StoragePath,
		},
	}

	if s.Options.ConfigFile != nil {
		ret.Volumes = []corev1.Volume{
			{
				Name: s.Options.ConfigFile.SecretName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: s.Options.ConfigFile.SecretName,
					},
				},
			},
		}

		ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
			Name:      s.Options.ConfigFile.SecretName,
			MountPath: s.Options.ConfigFile.MountPath(),
			ReadOnly:  true,
		})
	}

	// TODO: add headless service for cluster communication
	// TODO: PDB

	return ret
}
