local deployment = (import './hydra/deployment.libsonnet');
local configmap = (import './hydra/configmap.libsonnet');
local service = (import './hydra/service.libsonnet');
local usercreator = (import './hydra/usercreator.libsonnet');

local defaults = {
  local defaults = self,
  namespace: error 'must provide namespace',
  url: 'http://hydra.hydra.svc.cluster.local:4444/',
  audience: 'observatorium',
  clientId: 'user',
  clientSecret: 'secret',
  hydra_config: {
    dsn: 'sqlite:///var/lib/sqlite/hydra.sqlite?_fk=true',
    strategies: {
      access_token: 'jwt',
    },
    urls: {
      'self': {
        issuer: defaults.url,
      },
    },
  },
};

function(params)
  local config = defaults + params;
  local namespacePatch = {
    metadata+: {
      namespace: config.namespace,
    },
  };
  {
    config:: config,
    'hydra/service': service + namespacePatch,
    'hydra/deployment': deployment + namespacePatch,
    'hydra/usercreator': usercreator + namespacePatch + {
      spec+: {
        template+: {
          spec+: {
            containers: [
              super.containers[0]
              {
                args: [
                  '-v',
                  '--header',
                  'Content-Type: application/json',
                  '--data',
                  std.manifestJsonMinified({
                    audience: [config.namespace],
                    client_id: config.clientId,
                    client_secret: config.clientSecret,
                    grant_types: ['client_credentials'],
                    token_endpiont_auth_method: 'client_secret_basic',
                  }),
                  std.format('%s:4445/clients', config.url),
                ],
              },
            ],
          },
        },
      },
    },
    'hydra/configmap': configmap + namespacePatch + {
      data: {
        'config.yaml': std.manifestYamlDoc(config.hydra_config),
      },
    },
  }
