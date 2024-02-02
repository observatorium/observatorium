package up

import (
	"fmt"
	"net"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/log"
	upoptions "github.com/observatorium/up/pkg/options"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// NewQueriesFileOption creates a new queries file option with the given value.
func NewQueriesFileOption(value *QueriesFile) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/up/config/queries", "queries.yaml", "queries-file", "observatorium-up-queries")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

// NewTokenFileOption creates a new token file option with the given value.
func NewTokenFileOption(value *string) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/up/config/token", "token", "token-file", "observatorium-up-token")
	if value != nil {
		ret.WithValue(*value)
	}
	return ret
}

type UpOptions struct {
	Duration                *time.Duration                 `opt:"duration"`
	EndpointRead            string                         `opt:"endpoint-read"`
	EndpointType            EndpointType                   `opt:"endpoint-type"`
	EndpointWrite           string                         `opt:"endpoint-write"`
	InitialQueryDelay       *time.Duration                 `opt:"initial-query-delay"`
	Labels                  []string                       `opt:"labels"`
	Latency                 time.Duration                  `opt:"latency"`
	Listen                  *net.TCPAddr                   `opt:"listen"`
	LogLevel                log.Level                      `opt:"log.level"`
	Logs                    []string                       `opt:"logs"`
	LogsFile                string                         `opt:"logs-file"` // TODO: support this
	Name                    string                         `opt:"name"`
	Period                  time.Duration                  `opt:"period"`
	QueriesFile             containeropts.ContainerUpdater `opt:"queries-file"`
	Step                    time.Duration                  `opt:"step"`
	Tenant                  string                         `opt:"tenant"`
	TenantHeader            string                         `opt:"tenant-header"`
	Threshold               float64                        `opt:"threshold"`
	TLSClientCertFile       string                         `opt:"tls-client-cert-file"`
	TLSClientPrivateKeyFile string                         `opt:"tls-client-private-key-file"`
	TLSCAFile               string                         `opt:"tls-ca-file"`
	Token                   string                         `opt:"token"`
	TokenFile               containeropts.ContainerUpdater `opt:"token-file"`
}

type UpDeployment struct {
	options *UpOptions
	workload.DeploymentWorkload
}

// NewUp creates a new UpDeployment with standard defaults.
func NewUp(opts *UpOptions, namespace, imageTag string) *UpDeployment {
	if opts == nil {
		opts = &UpOptions{}
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "observatorium-up",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "blackbox-prober",
		workload.VersionLabel:   imageTag,
	}

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Name:                          "observatorium-up",
			Image:                         "quay.io/observatorium/up",
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

	return &UpDeployment{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (u *UpDeployment) Objects() []runtime.Object {
	container := u.makeContainer()
	return u.DeploymentWorkload.Objects(container)
}

func (u *UpDeployment) makeContainer() *workload.Container {
	ret := u.ToContainer()
	ret.Name = "observatorium-up"
	serverPort := kghelpers.GetPortOrDefault(8080, u.options.Listen)
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
	if u.options.QueriesFile != nil {
		u.options.QueriesFile.Update(ret)
	}
	if u.options.TokenFile != nil {
		u.options.TokenFile.Update(ret)
	}

	ret.Args = cmdopt.GetOpts(u.options)
	return ret
}
