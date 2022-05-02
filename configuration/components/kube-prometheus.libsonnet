local base_kp = (import 'kube-prometheus/main.libsonnet') +
                (import 'kube-prometheus/addons/all-namespaces.libsonnet') +
                (import 'kube-prometheus/addons/strip-limits.libsonnet')
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

function(params)
  local kp = base_kp {
    prometheus+: {
      prometheus+: {
        spec+: {
          remoteWrite: [{
            url: 'FIX ME',
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
            url: 'FIX ME',
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
  { ['kube-prometheus/blackbox-exporter-' + name]: kp.blackboxExporter[name] for name in std.objectFields(kp.blackboxExporter) } +
  { ['kube-prometheus/kube-state-metrics-' + name]: kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics) } +
  { ['kube-prometheus/alertmanager-' + name]: kp.alertmanager[name] for name in std.objectFields(kp.alertmanager) } +
  { ['kube-prometheus/prometheus-' + name]: kp.prometheus[name] for name in std.objectFields(kp.prometheus) }
