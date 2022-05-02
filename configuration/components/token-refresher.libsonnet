local defaults = {
  namespace: error 'must provide namespace',
  serviceMonitor: true,
  secretName: 'token-refresher-oidc',
  issuerUrl: error 'must provide issuerUrl',
  clientId: error 'must provide clientId',
  clientSecret: error 'must provide clientSecret',
  audience: error 'must provide audience',
};

local refresher = import 'token-refresher/token-refresher.libsonnet';

function(params)
  local config = defaults + params;
  local tr = refresher({
    name: 'token-refresher',
    namespace: config.namespace,
    version: 'master-2021-03-05-b34376b',
    url: config.issuerUrl,
    secretName: config.secretName,
    clientIDKey: 'clientId',
    issuerURLKey: 'issuerUrl',
    serviceMonitor: config.serviceMonitor,
  });

  { ['token-refresher/token-refresher-' + name]: tr[name] for name in std.objectFields(tr) } +
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
