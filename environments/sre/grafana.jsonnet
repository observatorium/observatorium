local thanos = (import 'thanos-mixin/mixin.libsonnet');
local thanosReceiveController = (import 'thanos-receive-controller-mixin/mixin.libsonnet');
local jaeger = (import 'jaeger-mixin/mixin.libsonnet');
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

local dashboards = thanos {
  _config+:: {
    thanosQuerierJobPrefix: 'thanos-querier',
    thanosStoreJobPrefix: 'thanos-store',
    thanosReceiveJobPrefix: 'thanos-receive',
    thanosRuleJobPrefix: 'thanos-rule',
    thanosCompactJobPrefix: 'thanos-compactor',
    thanosSidecarJobPrefix: 'thanos-sidecar',

    thanosQuerierSelector: 'job=~"%s.*"' % self.thanosQuerierJobPrefix,
    thanosStoreSelector: 'job=~"%s.*"' % self.thanosStoreJobPrefix,
    thanosReceiveSelector: 'job=~"%s.*"' % self.thanosReceiveJobPrefix,
    thanosRuleSelector: 'job=~"%s.*"' % self.thanosRuleJobPrefix,
    thanosCompactSelector: 'job=~"%s.*"' % self.thanosCompactJobPrefix,
    thanosSidecarSelector: 'job=~"%s.*"' % self.thanosSidecarJobPrefix,
  },
}.grafanaDashboards;

local thanosReceiveDashboards = thanosReceiveController {
  _config+:: {
    thanosReceiveJobPrefix: 'thanos-receive',
    thanosReceiveControllerJobPrefix: 'thanos-receive-controller',

    thanosReceiveSelector: 'job=~"%s.*"' % self.thanosReceiveJobPrefix,
    thanosReceiveControllerSelector: 'job=~"%s.*"' % self.thanosReceiveControllerJobPrefix,
  },
}.grafanaDashboards;

{
  ['grafana-dashboard-observatorium-thanos-%s.configmap' % std.split(name, '.')[0]]:
    local configmap = k.core.v1.configMap;
    configmap.new() +
    configmap.mixin.metadata.withName('grafana-dashboard-observatorium-thanos-%s' % std.split(name, '.')[0]) +
    configmap.withData({
      [name]: std.manifestJsonEx(dashboards[name], '  '),
    })
  for name in std.objectFields(dashboards)
} + {
  ['grafana-dashboard-observatorium-thanos-%s.configmap' % std.split(name, '.')[0]]:
    local configmap = k.core.v1.configMap;
    configmap.new() +
    configmap.mixin.metadata.withName('grafana-dashboard-observatorium-thanos-%s' % std.split(name, '.')[0]) +
    configmap.withData({
      [name]: std.manifestJsonEx(thanosReceiveDashboards[name], '  '),
    })
  for name in std.objectFields(thanosReceiveDashboards)
} + {
  ['grafana-dashboard-observatorium-jaeger-%s.configmap' % std.split(name, '.')[0]]:
    local configmap = k.core.v1.configMap;
    configmap.new() +
    configmap.mixin.metadata.withName('grafana-dashboard-observatorium-jaeger-%s' % std.split(name, '.')[0]) +
    configmap.withData({
      [name]: std.manifestJsonEx(jaeger.grafanaDashboards[name], '  '),
    })
  for name in std.objectFields(jaeger.grafanaDashboards)
}
