local dex = (import '../../components/dex.libsonnet') + {
  config+:: {
    local cfg = self,
    name: 'dex',
    namespace: 'dex',
    config+: {
      oauth2: {
        passwordConnector: 'local',
      },
      staticClients: [
        {
          id: 'test',
          name: 'test',
          secret: 'ZXhhbXBsZS1hcHAtc2VjcmV0',
        },
      ],
      enablePasswordDB: true,
      staticPasswords: [
        {
          email: 'admin@example.com',
          // bcrypt hash of the string "password"
          hash: '$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W',
          username: 'admin',
          userID: '08a8684b-db88-4b73-90a9-3cd1661f5466',
        },
      ],
    },
    version: 'v2.24.0',
    image: 'quay.io/dexidp/dex:v2.24.0',
    commonLabels+:: {
      'app.kubernetes.io/instance': 'e2e-test',
    },
  },
};

local obs = (import '../base/observatorium.jsonnet') + {
  config+:: {
    receivers+:: {
      logLevel: 'debug',
      debug: '1',
    },
  },


  api+: {
    config+: {
      rbac: {
        roles: [
          {
            name: 'read-write',
            resources: [
              'logs',
              'metrics',
            ],
            tenants: [
              'test',
            ],
            permissions: [
              'read',
              'write',
            ],
          },
        ],
        roleBindings: [
          {
            name: 'test',
            roles: [
              'read-write',
            ],
            subjects: [
              {
                name: dex.config.config.staticPasswords[0].email,
                kind: 'user',
              },
            ],
          },
        ],
      },
      tenants: {
        tenants: [
          {
            name: dex.config.config.staticClients[0].name,
            id: '1610b0c3-c509-4592-a256-a1871353dbfa',
            oidc: {
              clientID: dex.config.config.staticClients[0].id,
              clientSecret: dex.config.config.staticClients[0].secret,
              issuerURL: 'http://%s.%s.svc.cluster.local:%d/dex' % [
                dex.service.metadata.name,
                dex.service.metadata.namespace,
                dex.service.spec.ports[0].port,
              ],
              usernameClaim: 'email',
            },
          },
        ],
      },
    },
  },
};

local minio = (import '../../components/minio.libsonnet') + {
  config:: {
    namespace: 'observatorium-minio',
    bucketSecretNamespace: obs.config.namespace,
  },
};

local up = (import '../../components/up.libsonnet') + {
  config+:: {
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
    version: obs.config.up.version,
    image: obs.config.up.image,
    endpointType: 'metrics',
    writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/receive' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/query' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
  },
};

obs.manifests +
minio.manifests +
up.manifests +
dex.manifests
