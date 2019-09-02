local thanos = (import 'thanos-mixin/mixin.libsonnet');
local thanosReceiveController = (import 'thanos-receive-controller-mixin/mixin.libsonnet');
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

local dashboards = thanos {
  _config+:: {
    thanosQuerier: 'thanos-querier',
    thanosStore: 'thanos-store',
    thanosReceive: 'thanos-receive',
    thanosRule: 'thanos-rule',
    thanosCompact: 'thanos-compact',
    thanosSidecar: 'thanos-sidecar',

    thanosQuerierSelector: 'job="%s"' % self.thanosQuerier,
    thanosStoreSelector: 'job="%s"' % self.thanosStore,
    thanosReceiveSelector: 'job="%s"' % self.thanosReceive,
    thanosRuleSelector: 'job="%s"' % self.thanosRule,
    thanosCompactSelector: 'job="%s"' % self.thanosCompact,
    thanosSidecarSelector: 'job="%s"' % self.thanosSidecar,
  },
}.grafanaDashboards;

local thanosReceiveDashboards = thanosReceiveController {
  _config+:: {
    thanosReceive: 'thanos-receive',
    thanosReceiveControllerJobPrefix: 'thanos-receive-controller',

    thanosReceiveSelector: 'job="%s"' % self.thanosReceive,
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
}
