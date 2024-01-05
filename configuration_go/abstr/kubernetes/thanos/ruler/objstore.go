package ruler

import (
	"fmt"
	"net"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/objstore"
	"github.com/observatorium/observatorium/configuration_go/schemas/thanos/option"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultInternalPort = 8081
	defaultPublicPort   = 8080
)

type objstoreConfigFile = option.ConfigFile[objstore.BucketConfig]

// NewAlertRelabelConfigFile returns a new alertRelabelConfigFile option
func NewObjstoreConfigFile(name string, value objstore.BucketConfig) *objstoreConfigFile {
	return option.NewConfigFile("/etc/rules-objstore/objstore", "config.yaml", name, value)
}

type RulesObjstoreOptions struct {
	DebugName          string              `opt:"debug.name,single-hyphen"`
	LogFormat          string              `opt:"log.format,single-hyphen"`
	LogLevel           string              `opt:"log.level,single-hyphen"`
	ObjstoreConfigFile *objstoreConfigFile `opt:"objstore.config-file,single-hyphen"`
	WebHealthchecksURL string              `opt:"web.healthchecks.url,single-hyphen"`
	WebInternalListen  *net.TCPAddr        `opt:"web.internal.listen,single-hyphen"`
	WebListen          *net.TCPAddr        `opt:"web.listen,single-hyphen"`
}

type RulesObjstoreDeployment struct {
	Options *RulesObjstoreOptions

	k8sutil.DeploymentGenericConfig
}

func NewRulesObjstore() *RulesObjstoreDeployment {
	opts := &RulesObjstoreOptions{
		LogLevel:  "warn",
		LogFormat: "logfmt",
	}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "rules-objstore",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "rules-storage",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	return &RulesObjstoreDeployment{
		Options: opts,
		DeploymentGenericConfig: k8sutil.DeploymentGenericConfig{
			Image:                "quay.io/observatorium/rules-objstore",
			ImagePullPolicy:      corev1.PullIfNotPresent,
			Name:                 "observatorium-rules-objstore",
			Namespace:            defaultNamespace,
			CommonLabels:         commonLabels,
			Replicas:             1,
			PodResources:         k8sutil.NewResourcesRequirements("50m", "1", "200Mi", "400Mi"),
			Affinity:             k8sutil.NewAntiAffinity(nil, labelSelectors),
			EnableServiceMonitor: true,

			LivenessProbe: k8sutil.NewProbe("/live", defaultInternalPort, k8sutil.ProbeConfig{
				FailureThreshold: 10,
				PeriodSeconds:    30,
				TimeoutSeconds:   1,
				SuccessThreshold: 1,
			}),
			ReadinessProbe: k8sutil.NewProbe("/ready", defaultInternalPort, k8sutil.ProbeConfig{
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

	commonObjectMeta := k8sutil.MetaConfig{
		Name:      r.Name,
		Labels:    r.CommonLabels,
		Namespace: r.Namespace,
	}
	commonObjectMeta.Labels[k8sutil.VersionLabel] = container.ImageTag

	pod := &k8sutil.Pod{
		TerminationGracePeriodSeconds: &r.TerminationGracePeriodSeconds,
		Affinity:                      r.Affinity,
		SecurityContext:               r.SecurityContext,
		ServiceAccountName:            commonObjectMeta.Name,
		ContainerProviders:            append([]k8sutil.ContainerProvider{container}, r.Sidecars...),
	}

	deployment := &k8sutil.Deployment{
		MetaConfig: commonObjectMeta.Clone(),
		Replicas:   r.Replicas,
		Pod:        pod,
	}

	ret := k8sutil.ObjectMap{
		"rules-objstore-deployment": deployment.MakeManifest(),
	}

	service := &k8sutil.Service{
		MetaConfig:   commonObjectMeta.Clone(),
		ServicePorts: pod,
	}
	ret["rules-objstore-service"] = service.MakeManifest()

	if r.EnableServiceMonitor {
		serviceMonitor := &k8sutil.ServiceMonitor{
			MetaConfig:              commonObjectMeta.Clone(),
			ServiceMonitorEndpoints: pod,
		}
		ret["rules-objstore-serviceMonitor"] = serviceMonitor.MakeManifest()
	}

	serviceAccount := &k8sutil.ServiceAccount{
		MetaConfig: commonObjectMeta.Clone(),
		Name:       pod.ServiceAccountName,
	}
	ret["rules-objstore-serviceAccount"] = serviceAccount.MakeManifest()

	// Create configMaps required by the containers
	for name, config := range pod.GetConfigMaps() {
		configMap := &k8sutil.ConfigMap{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       config,
		}
		configMap.MetaConfig.Name = name
		ret["rules-objstore-configMap-"+name] = configMap.MakeManifest()
	}

	// Create secrets required by the containers
	for name, secret := range pod.GetSecrets() {
		secret := &k8sutil.Secret{
			MetaConfig: commonObjectMeta.Clone(),
			Data:       secret,
		}
		secret.MetaConfig.Name = name
		ret["rules-objstore-secret-"+name] = secret.MakeManifest()
	}

	return ret
}

func (r *RulesObjstoreDeployment) makeContainer() *k8sutil.Container {
	if r.Options == nil {
		r.Options = &RulesObjstoreOptions{}
	}

	internalPort := defaultInternalPort
	if r.Options.WebInternalListen != nil && r.Options.WebInternalListen.Port != 0 {
		internalPort = r.Options.WebInternalListen.Port
	}

	publicPort := defaultPublicPort
	if r.Options.WebListen != nil && r.Options.WebListen.Port != 0 {
		internalPort = r.Options.WebListen.Port
	}

	livenessPort := r.LivenessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if livenessPort != int32(internalPort) {
		panic(fmt.Sprintf(`liveness probe port %d does not match http port %d`, livenessPort, internalPort))
	}

	readinessPort := r.ReadinessProbe.ProbeHandler.HTTPGet.Port.IntVal
	if readinessPort != int32(internalPort) {
		panic(fmt.Sprintf(`readiness probe port %d does not match http port %d`, readinessPort, internalPort))
	}

	ret := r.ToContainer()
	ret.Name = "thanos"
	ret.Args = cmdopt.GetOpts(r.Options)
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

	if r.Options.ObjstoreConfigFile != nil {
		ret.ConfigMaps[r.Options.ObjstoreConfigFile.Name] = map[string]string{
			r.Options.ObjstoreConfigFile.FileName(): r.Options.ObjstoreConfigFile.Value.String(),
		}

		ret.Volumes = append(ret.Volumes, k8sutil.NewPodVolumeFromConfigMap("objstore", r.Options.ObjstoreConfigFile.Name))
		ret.VolumeMounts = append(ret.VolumeMounts, corev1.VolumeMount{
			Name:      "objstore",
			MountPath: r.Options.ObjstoreConfigFile.MountPath(),
		})
	}

	return ret
}
