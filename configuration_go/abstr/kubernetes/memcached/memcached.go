package memcached

import (
	"fmt"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultPort         = 11211
	dataVolumeName      = "data"
	exporterDefaultPort = 9150
)

// MemcachedOptions is the options for the memcached container.
type MemcachedOptions struct {
	ConnLimit       int    `opt:"conn-limit"`
	ListenBacklog   int    `opt:"listen-backlog"`
	MaxItemSize     string `opt:"max-item-size"`
	MaxReqsPerEvent int    `opt:"max-reqs-per-event"`
	MemoryLimit     int    `opt:"memory-limit"`
	Port            int    `opt:"port"`
	Threads         int    `opt:"threads"`
	Verbose         bool   `opt:"verbose"`
	VeryVerbose     bool   `opt:"vv,single-hyphen"`

	// Extra options not included above.
	cmdopt.ExtraOpts
}

// MemcachedDeployment is the memcached deployment.
type MemcachedDeployment struct {
	Options *MemcachedOptions
	k8sutil.DeploymentGenericConfig
	ExporterImage    string
	ExporterImageTag string
}

// NewMemcachedStatefulSet returns a new memcached deployment.
func NewMemcachedStatefulSet() *MemcachedDeployment {
	options := MemcachedOptions{}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "memcached",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "memcached",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	genericDeployment := k8sutil.DeploymentGenericConfig{
		Name:                          "memcached",
		Image:                         "docker.io/memcached",
		ImageTag:                      "latest",
		ImagePullPolicy:               corev1.PullIfNotPresent,
		CommonLabels:                  commonLabels,
		Replicas:                      1,
		Env:                           []corev1.EnvVar{},
		PodResources:                  k8sutil.NewResourcesRequirements("500m", "3", "2Gi", "3Gi"),
		Affinity:                      *k8sutil.NewAntiAffinity(nil, labelSelectors),
		EnableServiceMonitor:          true,
		TerminationGracePeriodSeconds: 120,
		SecurityContext:               k8sutil.GetDefaultSecurityContext(),
		ConfigMaps:                    map[string]map[string]string{},
		Secrets:                       map[string]map[string][]byte{},
	}

	return &MemcachedDeployment{
		Options:                 &options,
		DeploymentGenericConfig: genericDeployment,
		ExporterImage:           "quay.io/prometheus/memcached-exporter",
		ExporterImageTag:        "latest",
	}
}

// Manifests returns the manifests for the memcached deployment.
func (s *MemcachedDeployment) Manifests() k8sutil.ObjectMap {
	container := s.makeContainer()

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      s.Name,
		Labels:    s.CommonLabels,
		Namespace: s.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &s.TerminationGracePeriodSeconds,
		Affinity:                      &s.Affinity,
		SecurityContext:               s.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container, s.makeExporterContainer()}, s.Sidecars...),
	}

	deployment := &k8sutil.Deployment{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   s.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"memcached-deployment": deployment.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
		// As memcached is deployed as a deployment, we use a headless service.
		// Combined with DNS discovery, this ensures direct access to the pods.
		ClusterIP: "None",
	}
	ret["memcached-service"] = service.MakeManifest()

	if s.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["memcached-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["memcached-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["memcached-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["memcached-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (s *MemcachedDeployment) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &MemcachedOptions{}
	}

	httpPort := defaultPort
	if s.Options.Port != 0 {
		httpPort = s.Options.Port
	}

	ret := s.ToContainer()
	ret.Args = cmdopt.GetOpts(s.Options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "client",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("client", httpPort, httpPort),
	}

	return ret
}

func (s *MemcachedDeployment) makeExporterContainer() *k8sutil.Container {
	return &k8sutil.Container{
		Name:            "memcached-exporter",
		Image:           s.ExporterImage,
		ImageTag:        s.ExporterImageTag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       k8sutil.NewResourcesRequirements("50m", "200m", "50Mi", "200Mi"),
		Args: []string{
			fmt.Sprintf("--memcached.address=localhost:%d", s.Options.Port),
			fmt.Sprintf("--web.listen-address=:%d", exporterDefaultPort),
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: exporterDefaultPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ServicePorts: []corev1.ServicePort{
			k8sutil.NewServicePort("metrics", exporterDefaultPort, exporterDefaultPort),
		},
		MonitorPorts: []monv1.Endpoint{
			{
				Port:           "metrics",
				RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
			},
		},
	}
}
