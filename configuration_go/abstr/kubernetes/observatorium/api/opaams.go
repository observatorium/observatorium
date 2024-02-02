package api

import (
	"net"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
	kghelpers "github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

type OpaAmsOptions struct {
	AmsMappings             []string     `opt:"ams.mappings"`
	AmsMappingsPath         string       `opt:"ams.mappings-path"`
	AmsURL                  string       `opt:"ams.url"`
	DebugName               string       `opt:"debug.name"`
	InternalTracingEndpoint *net.TCPAddr `opt:"internal.tracing.endpoint"`
	LogFormat               string       `opt:"log.format"`
	LogLevel                string       `opt:"log.level"`
	Memcached               string       `opt:"memcached"`
	MemcachedExpire         int          `opt:"memcached.expire"`
	MemcachedInterval       int          `opt:"memcached.interval"`
	OidcAudience            string       `opt:"oidc.audience"`
	OidcClientID            string       `opt:"oidc.client-id"`
	OidcClientSecret        string       `opt:"oidc.client-secret"`
	OidcIssuerURL           string       `opt:"oidc.issuer-url"`
	OpaPackage              string       `opt:"opa.package"`
	OpaRule                 string       `opt:"opa.rule"`
	ResourceTypePrefix      string       `opt:"resource-type-prefix"`
	WebHealthchecksURL      string       `opt:"web.healthchecks.url"`
	WebInternalListen       *net.TCPAddr `opt:"web.internal.listen"`
	WebListen               *net.TCPAddr `opt:"web.listen"`
}

func MakeOpaAms(opts *OpaAmsOptions, enableMonitor bool) *workload.Container {
	webInternalListen, _ := net.ResolveTCPAddr("tcp", ":8081")
	if opts.WebInternalListen != nil {
		webInternalListen = opts.WebInternalListen
	}

	webListen, _ := net.ResolveTCPAddr("tcp", ":8080")
	if opts.WebListen != nil {
		webListen = opts.WebListen
	}

	ret := &workload.Container{
		Name:  "opa-ams",
		Image: "quay.io/observatorium/opa-ams",
		Args:  cmdopt.GetOpts(opts),
		LivenessProbe: kghelpers.NewProbe("/live", webInternalListen.Port, kghelpers.ProbeConfig{
			PeriodSeconds:    30,
			SuccessThreshold: 1,
			FailureThreshold: 10,
			TimeoutSeconds:   1,
		}),
		ReadinessProbe: kghelpers.NewProbe("/ready", webInternalListen.Port, kghelpers.ProbeConfig{
			PeriodSeconds:    5,
			SuccessThreshold: 1,
			FailureThreshold: 12,
			TimeoutSeconds:   1,
		}),
		Resources: kghelpers.NewResourcesRequirements("100m", "200m", "100Mi", "200Mi"),
		Ports: []corev1.ContainerPort{
			{
				Name:          "opa-ams-api",
				ContainerPort: int32(webListen.Port),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ServicePorts: []corev1.ServicePort{
			kghelpers.NewServicePort("opa-ams-api", webListen.Port, webListen.Port),
		},
	}

	if enableMonitor {
		ret.Ports = append(ret.Ports, corev1.ContainerPort{
			Name:          "opa-ams-metrics",
			ContainerPort: int32(webInternalListen.Port),
			Protocol:      corev1.ProtocolTCP,
		})
		ret.ServicePorts = append(ret.ServicePorts, kghelpers.NewServicePort("opa-ams-metrics", webInternalListen.Port, webInternalListen.Port))
		ret.MonitorPorts = []monv1.Endpoint{
			{
				Port:           "opa-ams-metrics",
				RelabelConfigs: kghelpers.GetDefaultServiceMonitorRelabelConfig(),
			},
		}
	}

	return ret
}
