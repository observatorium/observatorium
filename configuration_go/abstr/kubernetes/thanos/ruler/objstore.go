package ruler

import (
	"net"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/objstore"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultInternalPort = 8081
	defaultPublicPort   = 8080
)

type rulesObjstoreConfigFile = k8sutil.ConfigFile

// NewAlertRelabelConfigFile returns a new alertRelabelConfigFile option
func NewRulesObjstoreConfigFile(value *objstore.BucketConfig) *rulesObjstoreConfigFile {
	ret := k8sutil.NewConfigFile("/etc/rules-objstore/objstore", "config.yaml", "objstore", "observatorium-rules-objstore")
	if value != nil {
		ret.WithValue(value.String())
	}
	return ret
}

type RulesObjstoreOptions struct {
	DebugName          string                   `opt:"debug.name,single-hyphen"`
	LogFormat          string                   `opt:"log.format,single-hyphen"`
	LogLevel           string                   `opt:"log.level,single-hyphen"`
	ObjstoreConfigFile *rulesObjstoreConfigFile `opt:"objstore.config-file,single-hyphen"`
	WebHealthchecksURL string                   `opt:"web.healthchecks.url,single-hyphen"`
	WebInternalListen  *net.TCPAddr             `opt:"web.internal.listen,single-hyphen"`
	WebListen          *net.TCPAddr             `opt:"web.listen,single-hyphen"`
}

type RulesObjstoreDeployment struct {
	options *RulesObjstoreOptions

	k8sutil.DeploymentGenericConfig
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
		k8sutil.NameLabel:      "rules-objstore",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "rules-storage",
		k8sutil.VersionLabel:   imageTag,
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	probePort := k8sutil.GetPortOrDefault(defaultHTTPPort, opts.WebInternalListen)

	return &RulesObjstoreDeployment{
		options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/observatorium/rules-objstore",
			ImageTag:             imageTag,
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-rules-objstore",
			Namespace:            namespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			ContainerResources:   k8sutil.NewResourcesRequirements("50m", "1", "200Mi", "400Mi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/live", probePort, k8sutil.ProbeConfig{
				FailureThreshold: 10,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
				SuccessThreshold: 1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/ready", probePort, k8sutil.ProbeConfig{
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
}

func (r *RulesObjstoreDeployment) Manifests() k8sutil.ObjectMap {
	container := r.makeContainer()

	ret := k8sutil.ObjectMap{}
	ret.AddAll(r.GenerateObjects(container))

	return ret
}

func (r *RulesObjstoreDeployment) makeContainer() *k8sutil.Container {
	internalPort := k8sutil.GetPortOrDefault(defaultInternalPort, r.options.WebInternalListen)
	k8sutil.CheckProbePort(internalPort, r.LivenessProbe)
	k8sutil.CheckProbePort(internalPort, r.ReadinessProbe)

	publicPort := k8sutil.GetPortOrDefault(defaultPublicPort, r.options.WebListen)

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
		k8sutil.NewServicePort("internal", internalPort, internalPort),
		k8sutil.NewServicePort("public", publicPort, publicPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "internal",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
		},
	}

	if r.options.ObjstoreConfigFile != nil {
		r.options.ObjstoreConfigFile.AddToContainer(ret)
	}

	return ret
}
