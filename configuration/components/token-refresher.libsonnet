local defaults = {
  namespace: error 'must provide namespace',
  url: error 'must provide url',
  issuerUrl: error 'must provide issuerUrl',
  clientId: error 'must provide clientId',
  clientSecret: error 'must provide clientSecret',
  audience: error 'must provide audience',
  serviceMonitor: true,
  secretName: 'token-refresher-oidc',
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
    clientIDKey: 'clientId',
    issuerURLKey: 'issuerUrl',
    serviceMonitor: config.serviceMonitor,
  });

  {
    config:: config {
      ports: tr.config.ports,
      name: tr.config.name,
    },
  }
  { ['token-refresher/token-refresher-' + name]: tr[name] for name in std.objectFields(tr) }
  {
    'token-refresher/token-refrehser-oidc-secret': {
      kind: 'Secret',
      apiVersion: 'v1',
      metadata: {
        name: config.secretName,
        namespace: config.namespace,
      },
      type: 'Opaque',
      stringData: {
        audience: config.audience,
        clientId: config.clientId,
        clientSecret: config.clientSecret,
        issuerUrl: config.issuerUrl,
      },
    },
  }
