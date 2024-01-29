package ruler

import (
	"net"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/objstore"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultInternalPort = 8081
	defaultPublicPort   = 8080
)

// NewRulesObjstoreConfigFile creates a new ConfigFile option for the rules-objstore configuration file.
func NewRulesObjstoreConfigFile(value *objstore.BucketConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/rules-objstore/objstore", "config.yaml", "objstore", "observatorium-rules-objstore")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type RulesObjstoreOptions struct {
	DebugName          string                         `opt:"debug.name,single-hyphen"`
	LogFormat          string                         `opt:"log.format,single-hyphen"`
	LogLevel           string                         `opt:"log.level,single-hyphen"`
	ObjstoreConfigFile containeropts.ContainerUpdater `opt:"objstore.config-file,single-hyphen"`
	WebHealthchecksURL string                         `opt:"web.healthchecks.url,single-hyphen"`
	WebInternalListen  *net.TCPAddr                   `opt:"web.internal.listen,single-hyphen"`
	WebListen          *net.TCPAddr                   `opt:"web.listen,single-hyphen"`
}

type RulesObjstoreDeployment struct {
	options *RulesObjstoreOptions
	workload.DeploymentWorkload
}

func NewRulesObjstoreDefaultOptions() *RulesObjstoreOptions {
	return &RulesObjstoreOptions{
		LogLevel:  "warn",
		LogFormat: "logfmt",
	}
}

func NewRulesObjstore(opts *RulesObjstoreOptions, namespace, imageTag string) *RulesObjstoreDeployment {
	if opts == nil {
		opts = NewRulesObjstoreDefaultOptions()
	}

	commonLabels := map[string]string{
		workload.NameLabel:      "rules-objstore",
		workload.InstanceLabel:  "observatorium",
		workload.PartOfLabel:    "observatorium",
		workload.ComponentLabel: "rules-storage",
		workload.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		workload.NameLabel:     commonLabels[workload.NameLabel],
		workload.InstanceLabel: commonLabels[workload.InstanceLabel],
	}

	probePort := kghelpers.GetPortOrDefault(defaultInternalPort, opts.WebInternalListen)

	depWorkload := workload.DeploymentWorkload{
		Replicas: 1,
		PodConfig: workload.PodConfig{
			Image:                "quay.io/observatorium/rules-objstore",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-rules-objstore",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			ContainerResources:   kghelpers.NewResourcesRequirements("50m", "1", "200Mi", "400Mi"),
			Affinity:             kghelpers.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: kghelpers.NewProbe("/live", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 10,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
				SuccessThreshold: 1,
			}),
			ReadinessProbe: kghelpers.NewProbe("/ready", probePort, kghelpers.ProbeConfig{
				FailureThreshold: 12,
				PeriodSeconds:    5,
				TimeoutSeconds:   1,
				SuccessThreshold: 1,
			}),

			TerminationGracePeriodSeconds: 120,
			ConfigMaps:                    make(map[string]map[string]string),
			Secrets:                       make(map[string]map[string][]byte),
		},
	}

	return &RulesObjstoreDeployment{
		options:            opts,
		DeploymentWorkload: depWorkload,
	}
}

func (r *RulesObjstoreDeployment) Objects() []runtime.Object {
	container := r.makeContainer()
	return r.DeploymentWorkload.Objects(container)
}

func (r *RulesObjstoreDeployment) makeContainer() *workload.Container {
	internalPort := kghelpers.GetPortOrDefault(defaultInternalPort, r.options.WebInternalListen)
	kghelpers.CheckProbePort(internalPort, r.LivenessProbe)
	kghelpers.CheckProbePort(internalPort, r.ReadinessProbe)

	publicPort := kghelpers.GetPortOrDefault(defaultPublicPort, r.options.WebListen)

	ret := r.ToContainer()
	ret.Name = "thanos"
	ret.Args = cmdopt.GetOpts(r.options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "internal",
			ContainerPort: int32(internalPort),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "public",
			ContainerPort: int32(publicPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("internal", internalPort, internalPort),
		kghelpers.NewServicePort("public", publicPort, publicPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "internal",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if r.options.ObjstoreConfigFile != nil {
		r.options.ObjstoreConfigFile.Update(ret)
	}

	return ret
}
