package ruler

import (
	"net"
	"path/filepath"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	"github.com/observatorium/observatorium/configuration_go/kubegen/containeropts"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultSyncerInternalPort = 8083
)

type TenantsConfig struct {
	Tenants []TenantConfig `yaml:"tenants"`
}

type TenantConfig struct {
	ID string `yaml:"id"`
}

func NewTenantsFileOption(value *TenantsConfig) *containeropts.ConfigResourceAsFile {
	ret := containeropts.NewConfigResourceAsFile("/etc/thanos-rules-syncer/tenants", "config.yaml", "objstore", "observatorium-rules-syncer-tenants")
	if value != nil {
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			panic(err)
		}
		ret.WithValue(string(valueBytes))
	}
	return ret
}

type RulesSyncerOptions struct {
	File                string                         `opt:"file,single-hyphen"`
	Interval            int                            `opt:"interval,single-hyphen"`
	ObservatoriumApiUrl string                         `opt:"observatorium-api-url,single-hyphen"`
	ObservatoriumCa     string                         `opt:"observatorium-ca,single-hyphen"`
	OidcAudience        string                         `opt:"oidc.audience,single-hyphen"`
	OidcClientId        string                         `opt:"oidc.client-id,single-hyphen"`
	OidcClientSecret    string                         `opt:"oidc.client-secret,single-hyphen"`
	OidcIssuerUrl       string                         `opt:"oidc.issuer-url,single-hyphen"`
	RulesBackendUrl     string                         `opt:"rules-backend-url,single-hyphen"`
	Tenant              string                         `opt:"tenant,single-hyphen"`
	TenantsFile         containeropts.ContainerUpdater `opt:"tenants-file,single-hyphen"`
	ThanosRuleUrl       *net.TCPAddr                   `opt:"thanos-rule-url,single-hyphen"`
	WebInternalListen   *net.TCPAddr                   `opt:"web.internal.listen,single-hyphen"`
}

type RulesSyncerContainer struct {
	Options *RulesSyncerOptions

	workload.Container
}

func NewRulesSyncerContainer(opts *RulesSyncerOptions) *workload.Container {
	if opts == nil {
		opts = &RulesSyncerOptions{}
	}

	internalPort := kghelpers.GetPortOrDefault(defaultSyncerInternalPort, opts.WebInternalListen)

	ret := &workload.Container{}
	ret.Image = "quay.io/observatorium/observatorium-rules-syncer"
	ret.Name = "observatorium-rules-syncer"
	ret.Args = cmdopt.GetOpts(opts)
	ret.Resources = kghelpers.NewResourcesRequirements("32m", "128m", "64Mi", "128Mi")
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "internal",
			ContainerPort: int32(internalPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		kghelpers.NewServicePort("internal", internalPort, internalPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "internal",
			RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
		},
	}
	ret.Volumes = []corev1.Volume{
		{
			Name: "rule-syncer",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	ret.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "rule-syncer",
			MountPath: filepath.Dir(opts.File),
		},
	}

	if opts.TenantsFile != nil {
		opts.TenantsFile.Update(ret)
	}

	return ret

}
