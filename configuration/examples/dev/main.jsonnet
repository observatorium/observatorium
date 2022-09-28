local dex = (import '../../components/dex.libsonnet')({
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
        issuerCAPath: '/var/run/tls/test/service-ca.crt',
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
});

local minio = (import '../../components/minio.libsonnet')({
  namespace: 'observatorium-minio',
  buckets: ['thanos', 'loki', 'rules'],
  accessKey: 'minio',
  secretKey: 'minio123',
});

local api = (import 'observatorium-api/observatorium-api.libsonnet');
local loki = (import '../../components/loki.libsonnet');
local obs = (import '../../components/observatorium.libsonnet');
local dev = obs {
  loki: loki(
    obs.loki.config {
      config+: {
        querier+: {
          engine+: {
            max_look_back_period: '5m',
          },
        },
      },
    },
  ),


  api: api(
    obs.api.config {
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
              issuerURL: 'https://%s.%s.svc.cluster.local:%d/dex' % [
                dex.service.metadata.name,
                dex.service.metadata.namespace,
                dex.service.spec.ports[0].port,
              ],
              issuerCAPath: dex.config.config.staticClients[0].issuerCAPath,
              usernameClaim: 'email',
              configMapName:: '%s-ca-tls' % [dex.config.config.staticClients[0].id],
              caKey:: 'service-ca.crt',
            },
            rateLimits: [
              {
                endpoint: '/api/metrics/v1/.+/api/v1/receive',
                limit: 1000,
                window: '1s',
              },
              {
                endpoint: '/api/logs/v1/.*',
                limit: 1000,
                window: '1s',
              },
            ],
          },
        ],
      },
    },
  ),
};

local up = (import 'up/up.libsonnet')({
  local cfg = self,
  name: dev.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
  namespace: dev.config.namespace,
  replicas: 1,
  commonLabels+:: dev.config.commonLabels,
  version: 'master-2020-11-04-0c6ece8',
  image: 'quay.io/observatorium/up:' + cfg.version,
  endpointType: 'metrics',
  writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/receive' % [
    dev.api.service.metadata.name,
    dev.api.service.metadata.namespace,
    dev.api.service.spec.ports[1].port,
  ],
  readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/query' % [
    dev.api.service.metadata.name,
    dev.api.service.metadata.namespace,
    dev.api.service.spec.ports[1].port,
  ],
});

dev.manifests +
{ ['up-' + name]: up[name] for name in std.objectFields(up) if up[name] != null } +
{
  'minio-deployment': minio.deployment,
  'minio-pvc': minio.pvc,
  'minio-secret-thanos': {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'thanos-objectstorage',
      namespace: dev.config.namespace,
    },
    stringData: {
      'thanos.yaml': |||
        type: s3
        config:
          bucket: thanos
          endpoint: %s.%s.svc.cluster.local:9000
          insecure: true
          access_key: minio
          secret_key: minio123
      ||| % [minio.service.metadata.name, minio.config.namespace],
    },
    type: 'Opaque',
  },
  'minio-secret-loki': {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'loki-objectstorage',
      namespace: dev.config.namespace,
    },
    stringData: {
      endpoint: 'http://minio:minio123@%s.%s.svc.cluster.local.:9000/loki' % [minio.service.metadata.name, minio.config.namespace],
    },
    type: 'Opaque',
  },
  'minio-secret-observatorium-rules': {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'obs-rules-objectstorage',
      namespace: dev.config.namespace,
    },
    stringData: {
      endpoint: 'http://minio:minio123@%s.%s.svc.cluster.local.:9000/rules' % [minio.service.metadata.name, minio.config.namespace],
    },
    type: 'Opaque',
  },
  'minio-service': minio.service,
} +
{ ['dex-' + name]: dex[name] for name in std.objectFields(dex) if dex[name] != null }
