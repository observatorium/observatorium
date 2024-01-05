package ruler

import (
	"net"
	"path/filepath"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultSyncerInternalPort = 8083
)

type RulesSyncerOptions struct {
	File                string       `opt:"file,single-hyphen"`
	Interval            int          `opt:"interval,single-hyphen"`
	ObservatoriumApiUrl string       `opt:"observatorium-api-url,single-hyphen"`
	ObservatoriumCa     string       `opt:"observatorium-ca,single-hyphen"`
	OidcAudience        string       `opt:"oidc.audience,single-hyphen"`
	OidcClientId        string       `opt:"oidc.client-id,single-hyphen"`
	OidcClientSecret    string       `opt:"oidc.client-secret,single-hyphen"`
	OidcIssuerUrl       string       `opt:"oidc.issuer-url,single-hyphen"`
	RulesBackendUrl     string       `opt:"rules-backend-url,single-hyphen"`
	Tenant              string       `opt:"tenant,single-hyphen"`
	ThanosRuleUrl       *net.TCPAddr `opt:"thanos-rule-url,single-hyphen"`
	WebInternalListen   *net.TCPAddr `opt:"web.internal.listen,single-hyphen"`
}

type RulesSyncerContainer struct {
	Options *RulesSyncerOptions

	k8sutil.Container
}

func NewRulesSyncerContainer(opts *RulesSyncerOptions) *k8sutil.Container {
	if opts == nil {
		opts = &RulesSyncerOptions{}
	}

	internalPort := defaultSyncerInternalPort
	if opts.WebInternalListen != nil && opts.WebInternalListen.Port != 0 {
		internalPort = opts.WebInternalListen.Port
	}

	ret := &k8sutil.Container{}
	ret.Image = "quay.io/observatorium/observatorium-rules-syncer"
	ret.Name = "observatorium-rules-syncer"
	ret.Args = cmdopt.GetOpts(opts)
	ret.Resources = k8sutil.NewResourcesRequirements("32m", "128m", "64Mi", "128Mi")
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "internal",
			ContainerPort: int32(internalPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("internal", internalPort, internalPort),
	}
	ret.MonitorPorts = []monv1.Endpoint{
		{
			Port:           "internal",
			RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
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

	return ret

}
