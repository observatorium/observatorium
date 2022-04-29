local kp = (import 'kube-prometheus/main.libsonnet') +
           (import 'kube-prometheus/addons/all-namespaces.libsonnet') + {
  values+:: {
    common+: {
      namespace: 'monitoring',
    },
    prometheus+: {
      namespaces: [],
    },
  },
};

function(params) {
  'kube-prometheus/0prometheus-operator': [
    kp.prometheusOperator[name]
    for name in std.filter(
      (function(name) name != 'serviceMonitor' && name != 'prometheusRule'),
      std.objectFields(kp.prometheusOperator)
    )
  ],
  'kube-prometheus/00namespace': [kp.kubePrometheus[name] for name in std.objectFields(kp.kubePrometheus)],
  'kube-prometheus/grafana': [kp.grafana[name] for name in std.objectFields(kp.grafana)],
  'kube-prometheus/node-exporter': [kp.nodeExporter[name] for name in std.objectFields(kp.nodeExporter)],
  'kube-prometheus/blackbox-exporter': [kp.blackboxExporter[name] for name in std.objectFields(kp.blackboxExporter)],
  'kube-prometheus/kube-state-metrics': [kp.kubeStateMetrics[name] for name in std.objectFields(kp.kubeStateMetrics)],
  'kube-prometheus/alertmanager': [kp.alertmanager[name] for name in std.objectFields(kp.alertmanager)],
  'kube-prometheus/prometheus': [kp.prometheus[name] for name in std.objectFields(kp.prometheus)],
}
