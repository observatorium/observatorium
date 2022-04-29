local defaults = {
  namespace: error 'must provide namespace',
  url: error 'must provide url',
  secretName: error 'must provide secretName',
  serviceMonitor: true,
};

local refresher = import 'token-refresher/token-refresher.libsonnet';

function(params)
  local config = defaults + params;
  local tr = refresher({
    name: 'token-refresher',
    namespace: config.namespace,
    version: 'master-2021-03-05-b34376b',
    url: config.url,
    secretName: config.secretName,
    serviceMonitor: config.serviceMonitor,
  });

  { ['token-refresher/token-refresher-' + name]: tr[name] for name in std.objectFields(tr) }
