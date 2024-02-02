package memcached

import (
	"fmt"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	workload.DeploymentWorkload
	ExporterImage    string
	ExporterImageTag string
}

// NewMemcachedStatefulSet returns a new memcached deployment.
func NewMemcached() *MemcachedDeployment {
	options := MemcachedOptions{}

	commonLabels := map[string]string{
		workload.NameLabel:      "memcached",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "memcached",
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Name:                          "memcached",
			Image:                         "docker.io/memcached",
			ImageTag:                      "latest",
			ImagePullPolicy:               corev1.PullIfNotPresent,
			CommonLabels:                  commonLabels,
			Env:                           []corev1.EnvVar{},
			ContainerResources:            kghelpers.NewResourcesRequirements("500m", "3", "2Gi", "3Gi"),
			Affinity:                      kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor:          true,
			TerminationGracePeriodSeconds: 120,
			ConfigMaps:                    map[string]map[string]string{},
			Secrets:                       map[string]map[string][]byte{},
		},
	}

	return &MemcachedDeployment{
		Options:            &options,
		DeploymentWorkload: depWorkload,
		ExporterImage:      "quay.io/prometheus/memcached-exporter",
		ExporterImageTag:   "latest",
	}
}

// Manifests returns the manifests for the memcached deployment.
func (m *MemcachedDeployment) Objects() []runtime.Object {
	if m.EnableServiceMonitor {
		m.Sidecars = append(m.Sidecars, m.makeExporterContainer())
	}

	container := m.makeContainer()
	ret := m.DeploymentWorkload.Objects(container)

	// Set headless service to get stable network ID.
	service := kghelpers.GetObject[*corev1.Service](ret, "")
	service.Spec.ClusterIP = corev1.ClusterIPNone

	return ret
}

func (m *MemcachedDeployment) makeContainer() *workload.Container {
	if m.Options == nil {
		m.Options = &MemcachedOptions{}
	}

	httpPort := defaultPort
	if m.Options.Port != 0 {
		httpPort = m.Options.Port
	}

	ret := m.ToContainer()
	ret.Name = "memcached"
	ret.Args = cmdopt.GetOpts(m.Options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "client",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("client", httpPort, httpPort),
	}

	return ret
}

func (m *MemcachedDeployment) makeExporterContainer() *workload.Container {
	return &workload.Container{
		Name:            "memcached-exporter",
		Image:           m.ExporterImage,
		ImageTag:        m.ExporterImageTag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       kghelpers.NewResourcesRequirements("50m", "200m", "50Mi", "200Mi"),
		Args: []string{
			fmt.Sprintf("--memcached.address=localhost:%d", m.Options.Port),
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
			kghelpers.NewServicePort("metrics", exporterDefaultPort, exporterDefaultPort),
		},
		MonitorPorts: []monv1.Endpoint{
			{
				Port:           "metrics",
				RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
			},
		},
	}
}
