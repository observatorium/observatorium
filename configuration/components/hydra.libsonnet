local deployment = (import './hydra/deployment.libsonnet');
local configmap = (import './hydra/configmap.libsonnet');
local service = (import './hydra/service.libsonnet');
local usercreator = (import './hydra/usercreator.libsonnet');

local defaults = {
  local defaults = self,
  namespace: 'hydra',
  audience: 'observatorium',
  clientId: 'user',
  clientSecret: 'secret',
  image: 'oryd/hydra:v1.11.7',
};

function(params)
  local config = defaults + params;
  local baseUrl = std.format('http://hydra.%s.svc.cluster.local', config.namespace);
  local hydraConfig = {
    dsn: 'sqlite:///var/lib/sqlite/hydra.sqlite?_fk=true',
    strategies: {
      access_token: 'jwt',
    },
    urls: {
      'self': {
        issuer: std.format('%s:4444/', baseUrl),
      },
    },
  };
  local namespacePatch = {
    metadata+: {
      namespace: config.namespace,
    },
  };
  {
    config:: config { issuerUrl: hydraConfig.urls['self'].issuer },
    'hydra/service': service + namespacePatch,
    'hydra/deployment': deployment + namespacePatch + {
      spec+: {
        template+: {
          spec+: {
            initContainers: [
              super.initContainers[0]
              {
                image: config.image,
              },
            ],
            containers: [
              super.containers[0]
              {
                image: config.image,
              },
            ],
          },
        },
      },
    },
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
                    audience: [config.audience],
                    client_id: config.clientId,
                    client_secret: config.clientSecret,
                    grant_types: ['client_credentials'],
                    token_endpoint_auth_method: 'client_secret_basic',
                  }),
                  std.format('%s:4445/clients', baseUrl),
                ],
              },
            ],
          },
        },
      },
    },
    'hydra/configmap': configmap + namespacePatch + {
      data: {
        'config.yaml': std.manifestYamlDoc(hydraConfig),
      },
    },
  }
