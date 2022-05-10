local base_kp = (import 'kube-prometheus/main.libsonnet') +
                (import 'kube-prometheus/addons/all-namespaces.libsonnet') +
                (import 'kube-prometheus/addons/strip-limits.libsonnet') +
                (import 'kube-prometheus/addons/networkpolicies-disabled.libsonnet')
                + {
                  values+:: {
                    common+: {
                      namespace: 'monitoring',
                    },
                    prometheus+: {
                      namespaces: [],
                    },
                  },
                };

local defaults = {
  observatoriumRemoteWriteUrl: error 'must provide observatoriumRemoteWriteUrl',
  observatoriumDatasourceUrl: error 'must provide observatoriumDatasourceUrl',
};

function(params)
  local config = defaults + params;
  local kp = base_kp {
    prometheus+: {
      prometheus+: {
        spec+: {
          remoteWrite: [{
            url: config.observatoriumRemoteWriteUrl,
          }],
        },
      },
    },
    values+:: {
      grafana+:: {
        datasources+: [
          {
            name: 'Observatorium API',
            type: 'prometheus',
            access: 'proxy',
            url: config.observatoriumDatasourceUrl,
            version: 1,
            editable: false,
          },
        ],
      },
    },
  };
  {
    ['kube-prometheus/setup/0prometheus-operator-' + name]:
      kp.prometheusOperator[name]
    for name in std.filter(
      (function(name) name != 'serviceMonitor' && name != 'prometheusRule'),
      std.objectFields(kp.prometheusOperator)
    )
  } +
  { ['kube-prometheus/setup/00namespace-' + name]: kp.kubePrometheus[name] for name in std.filter((function(name) name == 'namespace'), std.objectFields(kp.kubePrometheus)) } +
  { ['kube-prometheus/grafana-' + name]: kp.grafana[name] for name in std.objectFields(kp.grafana) } +
  { ['kube-prometheus/node-exporter' + name]: kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter) } +
  { ['kube-prometheus/alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
  { ['kube-prometheus/prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) }
