package up

import (
	"fmt"
	"net"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	upoptions "github.com/observatorium/up/pkg/options"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

type EndpointType string

const (
	EndpointTypeLogs    EndpointType = "logs"
	EndpointTypeMetrics EndpointType = "metrics"
)

type QueriesFile struct {
	Queries []upoptions.QuerySpec  `yaml:"queries"`
	Labels  []upoptions.LabelSpec  `yaml:"labels"`
	Series  []upoptions.SeriesSpec `yaml:"series"`
}

func (q QueriesFile) String() string {
	res, err := yaml.Marshal(q)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal queries file: %v", err))
	}
	return string(res)
}

type queriesFileOption = k8sutil.ConfigFile

// NewQueriesFileOption creates a new queries file option with the given value.
// If the related configmap already exists, you can pass nil to val and call WithExistingCM.
func NewQueriesFileOption(value *QueriesFile) *queriesFileOption {
	ret := k8sutil.NewConfigFile("/etc/up/config/queries", "queries.yaml", "queries-file", "observatorium-up-queries")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type tokenFileOption = k8sutil.ConfigFile

func NewTokenFileOption(value *string) *tokenFileOption {
	ret := k8sutil.NewConfigFile("/etc/up/config/token", "token", "token-file", "observatorium-up-token")
	if value != nil {
		ret.WithValue(*value)
	}
	return ret
}

type UpOptions struct {
	Duration                *model.Duration    `opt:"duration"`
	EndpointRead            string             `opt:"endpoint-read"`
	EndpointType            EndpointType       `opt:"endpoint-type"`
	EndpointWrite           string             `opt:"endpoint-write"`
	InitialQueryDelay       *model.Duration    `opt:"initial-query-delay"`
	Labels                  []string           `opt:"labels"`
	Latency                 model.Duration     `opt:"latency"`
	Listen                  *net.TCPAddr       `opt:"listen"`
	LogLevel                log.LogLevel       `opt:"log.level"`
	Logs                    []string           `opt:"logs"`
	LogsFile                string             `opt:"logs-file"` // TODO: support this
	Name                    string             `opt:"name"`
	Period                  model.Duration     `opt:"period"`
	QueriesFile             *queriesFileOption `opt:"queries-file"`
	Step                    model.Duration     `opt:"step"`
	Tenant                  string             `opt:"tenant"`
	TenantHeader            string             `opt:"tenant-header"`
	Threshold               float64            `opt:"threshold"`
	TLSClientCertFile       string             `opt:"tls-client-cert-file"`
	TLSClientPrivateKeyFile string             `opt:"tls-client-private-key-file"`
	TLSCAFile               string             `opt:"tls-ca-file"`
	Token                   string             `opt:"token"`
	TokenFile               *tokenFileOption   `opt:"token-file"`
}

type UpDeployment struct {
	options *UpOptions
	k8sutil.DeploymentGenericConfig
}

// NewUp creates a new UpDeployment with standard defaults.
func NewUp(opts *UpOptions, namespace, imageTag string) *UpDeployment {
	if opts == nil {
		opts = &UpOptions{}
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "observatorium-up",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "blackbox-prober",
		k8sutil.VersionLabel:   imageTag,
	}

	return &UpDeployment{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Name:                          "observatorium-up",
			Image:                         "quay.io/observatorium/up",
			ImageTag:                      imageTag,
			ImagePullPolicy:               corev1.PullIfNotPresent,
			Namespace:                     namespace,
			CommonLabels:                  commonLabels,
			Replicas:                      1,
			ContainerResources:            k8sutil.NewResourcesRequirements("100m", "500m", "1Gi", "2Gi"),
			EnableServiceMonitor:          true,
			TerminationGracePeriodSeconds: 30,
			Env:                           []corev1.EnvVar{},
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}
}

func (u *UpDeployment) Manifests() k8sutil.ObjectMap {
	container := u.makeContainer()
	ret := k8sutil.ObjectMap{}
	ret.AddAll(u.GenerateObjects(container))

	return ret
}

func (u *UpDeployment) makeContainer() *k8sutil.Container {
	ret := u.ToContainer()
	ret.Name = "observatorium-up"
	serverPort := getPort(8080, u.options.Listen)
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
	if u.options.QueriesFile != nil {
		u.options.QueriesFile.AddToContainer(ret)
	}
	if u.options.TokenFile != nil {
		u.options.TokenFile.AddToContainer(ret)
	}

	ret.Args = cmdopt.GetOpts(u.options)
	return ret
}

func getPort(defaultValue int, addr *net.TCPAddr) int {
	if addr != nil {
		return addr.Port
	}
	return defaultValue
}
