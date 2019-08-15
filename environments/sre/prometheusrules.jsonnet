local thanos = (import 'thanos-mixin/mixin.libsonnet');

{
  'observatorium-thanos-stage.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-stage',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos {
      _config+:: {
        thanosQuerierSelector: 'job="thanos-querier", namespace="telemeter-stage"',
        thanosStoreSelector: 'job="thanos-store", namespace="telemeter-stage"',
        thanosReceiveSelector: 'job="thanos-receive", namespace="telemeter-stage"',
      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },
    spec: alerts.prometheusAlerts,
  },
  'observatorium-thanos-production.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-production',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos {
      _config+:: {
        thanosQuerierSelector: 'job="thanos-querier", namespace="telemeter-production"',
        thanosStoreSelector: 'job="thanos-store", namespace="telemeter-production"',
        thanosReceiveSelector: 'job="thanos-receive", namespace="telemeter-production"',
      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },
    spec: alerts.prometheusAlerts,
  },
}
