local hydra = (import '../../components/hydra.libsonnet')({});

local tenant = {
  name: 'test-oidc',
  id: '1610b0c3-c509-4592-a256-a1871353dbfa',
  clientID: 'observatorium',
  issuerURL: hydra.config.issuerUrl,
  user: 'user',
};

local minio = (import '../../components/minio.libsonnet')({
  namespace: 'observatorium-minio',
  buckets: ['thanos', 'loki', 'rules'],
  accessKey: 'minio',
  secretKey: 'minio123',
});

local api = (import 'observatorium-api/observatorium-api.libsonnet');
local obs = (import '../../components/observatorium.libsonnet');
local tracing = (import '../../components/tracing.libsonnet');
local dev = obs {
  tracing: tracing(
    obs.tracing.config {
      tenants: [tenant.name],
      enabled: true,
      monitoring: true,
      jaegerSpec: {
        strategy: 'allinone',
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
              'traces',
            ],
            tenants: [
              tenant.name,
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
                name: tenant.user,
                kind: 'user',
              },
            ],
          },
        ],
      },
      tenants: {
        tenants: [
          {
            name: tenant.name,
            id: tenant.id,
            oidc: {
              clientID: tenant.clientID,
              issuerURL: tenant.issuerURL,
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

local token_refresher = (import '../../components/token-refresher.libsonnet')({
  namespace: dev.api.config.namespace,
  version: 'master-2021-03-05-b34376b',
  url: std.format('http://%s.%s.svc.cluster.local:%d', [dev.api.service.metadata.name, dev.api.service.metadata.namespace, dev.api.service.spec.ports[2].port]),
  issuerUrl: hydra.config.issuerUrl,
  clientId: hydra.config.clientId,
  clientSecret: hydra.config.clientSecret,
  audience: hydra.config.audience,
});

local kube_prometheus = (import '../../components/kube-prometheus.libsonnet')({
  observatoriumDatasourceUrl: std.format('http://%s.%s.svc.cluster.local:%d/api/metrics/v1/%s/',
                                         [
                                           token_refresher.config.name,
                                           token_refresher.config.namespace,
                                           token_refresher.config.ports.web,
                                           tenant.name,
                                         ]),
  observatoriumRemoteWriteUrl: std.format('http://%s.%s.svc.cluster.local:%d/api/metrics/v1/%s/api/v1/receive',
                                          [
                                            token_refresher.config.name,
                                            token_refresher.config.namespace,
                                            token_refresher.config.ports.web,
                                            tenant.name,
                                          ]),
});

dev.manifests
{
  'minio/minio-deployment': minio.deployment,
  'minio/minio-pvc': minio.pvc,
  'minio/minio-secret-thanos': {
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
  'minio/minio-secret-loki': {
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
  'minio/minio-service': minio.service,
} +
kube_prometheus +
token_refresher +
hydra
