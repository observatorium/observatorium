local thanos = import 'thanos-mixin/mixin.libsonnet';
local thanosReceiveController = import 'thanos-receive-controller-mixin/mixin.libsonnet';

{
  rules+: {
    prometheusrule+: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'PrometheusRule',
      metadata: {
        name: 'observatorium-thanos',
        labels: {
          prometheus: 'app-sre',
          role: 'alert-rules',
        },
      },

      local alerts = thanos + thanosReceiveController {
        _config+:: {
          thanosQuerierJobPrefix: 'thanos-query',
          thanosStoreJobPrefix: 'thanos-store',
          thanosReceiveJobPrefix: 'thanos-receive-.*',
          thanosCompactJobPrefix: 'thanos-compact',
          thanosReceiveControllerJobPrefix: 'thanos-receive-controller',

          thanosQuerierSelector: 'job="%s"' % self.thanosQuerierJobPrefix,
          thanosStoreSelector: 'job="%s"' % self.thanosStoreJobPrefix,
          thanosReceiveSelector: 'job=~"%s"' % self.thanosReceiveJobPrefix,
          thanosCompactSelector: 'job="%s"' % self.thanosCompactJobPrefix,
          thanosReceiveControllerSelector: 'job="%s"' % self.thanosReceiveControllerJobPrefix,

          local config = self,
          // We build alerts for the presence of all these jobs.
          jobs: {
            ThanosQuerier: config.thanosQuerierSelector,
            ThanosStore: config.thanosStoreSelector,
            ThanosCompact: config.thanosCompactSelector,
          } + {
            ['ThanosReceive' + capitalize(tenant.hashring)]: 'job="thanos-receive-%s"' % tenant.hashring
            for tenant in tenants
          },
        },

        // Filter rule groups that we don't care about, like the sidecar
        prometheusAlerts+:: {
          groups: std.filter(function(g) (g.name != 'thanos-sidecar.rules'), super.groups),
        },
      },
      spec: alerts.prometheusAlerts,
    },
  },
}
