package avalanche

import (
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
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
	k8sutil.DeploymentGenericConfig
}

func NewAvalanche(opts *AvalancheOptions, namespace, imageTag string) *AvalancheDeployment {
	commonLabels := map[string]string{
		k8sutil.NameLabel:      "avalanche",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "avalanche",
		k8sutil.VersionLabel:   imageTag,
	}

	return &AvalancheDeployment{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Name:                          "avalanche",
			Image:                         "quay.io/prometheuscommunity/avalanche",
			ImageTag:                      imageTag,
			ImagePullPolicy:               corev1.PullIfNotPresent,
			Namespace:                     namespace,
			CommonLabels:                  commonLabels,
			Replicas:                      1,
			PodResources:                  k8sutil.NewResourcesRequirements("100m", "500m", "1Gi", "2Gi"),
			EnableServiceMonitor:          true,
			TerminationGracePeriodSeconds: 30,
			Env:                           []corev1.EnvVar{},
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}
}

func (u *AvalancheDeployment) Manifests() k8sutil.ObjectMap {
	container := u.makeContainer()
	ret := k8sutil.ObjectMap{}
	ret.AddAll(u.GenerateObjects(container))

	return ret
}

func (u *AvalancheDeployment) makeContainer() *k8sutil.Container {
	ret := u.ToContainer()
	ret.Name = "avalanche"
	serverPort := 9001
	if u.options.Port != 0 {
		serverPort = u.options.Port
	}

	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(serverPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("http", serverPort, serverPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port: "http",
		},
	}

	ret.Args = cmdopt.GetOpts(u.options)
	return ret
}
