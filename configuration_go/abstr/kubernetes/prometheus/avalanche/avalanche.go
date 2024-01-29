package avalanche

import (
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AvalancheOptions struct {
	MetricCount         int           `opt:"metric-count"`
	LabelCount          int           `opt:"label-count"`
	SeriesCount         int           `opt:"series-count"`
	MetricNameLength    int           `opt:"metricname-length"`
	LabelNameLength     int           `opt:"labelname-length"`
	ConstLabels         []string      `opt:"const-label"`
	ValueInterval       int           `opt:"value-interval"`
	SeriesInterval      int           `opt:"series-interval"`
	MetricInterval      int           `opt:"metric-interval"`
	Port                int           `opt:"port"`
	RemoteURL           string        `opt:"remote-url"`
	RemotePprofURLs     []string      `opt:"remote-pprof-urls"`
	RemotePprofInterval time.Duration `opt:"remote-pprof-interval"`
	RemoteBatchSize     int           `opt:"remote-batch-size"`
	RemoteRequestsCount int           `opt:"remote-requests-count"`
	RemoteWriteInterval time.Duration `opt:"remote-write-interval"`
	RemoteTenant        string        `opt:"remote-tenant"`
	TLSClientInsecure   bool          `opt:"tls-client-insecure,noval"`
	RemoteTenantHeader  string        `opt:"remote-tenant-header"`
}

type AvalancheDeployment struct {
	options *AvalancheOptions
	workload.DeploymentWorkload
}

func NewAvalanche(opts *AvalancheOptions, namespace, imageTag string) *AvalancheDeployment {
	commonLabels := map[string]string{
		workload.NameLabel:      "avalanche",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "avalanche",
		workload.VersionLabel:   imageTag,
	}

	dedWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Name:                          "avalanche",
			Image:                         "quay.io/prometheuscommunity/avalanche",
			ImageTag:                      imageTag,
			ImagePullPolicy:               corev1.PullIfNotPresent,
			Namespace:                     namespace,
			CommonLabels:                  commonLabels,
			ContainerResources:            kghelpers.NewResourcesRequirements("100m", "500m", "1Gi", "2Gi"),
			EnableServiceMonitor:          true,
			TerminationGracePeriodSeconds: 30,
			Env:                           []corev1.EnvVar{},
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}

	return &AvalancheDeployment{
		options:            opts,
		DeploymentWorkload: dedWorkload,
	}
}

func (a *AvalancheDeployment) Objects() []runtime.Object {
	container := a.makeContainer()
	return a.DeploymentWorkload.Objects(container)
}

func (a *AvalancheDeployment) makeContainer() *workload.Container {
	ret := a.ToContainer()
	ret.Name = "avalanche"
	serverPort := 9001
	if a.options.Port != 0 {
		serverPort = a.options.Port
	}

	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(serverPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("http", serverPort, serverPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port: "http",
		},
	}

	ret.Args = cmdopt.GetOpts(a.options)
	return ret
}
