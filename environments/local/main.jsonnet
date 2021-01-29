local tenant = {
  name: 'test-oidc',
  id: '1610b0c3-c509-4592-a256-a1871353dbfa',
  clientID: 'observatorium',
  issuerURL: 'http://172.17.0.1:4444/',
  user: 'user',
};

local api = (import 'observatorium/observatorium-api.libsonnet');
local obs = (import '../../components/observatorium.libsonnet');
local dev = obs {
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

local minio = (import '../../components/minio.libsonnet')({
  namespace: 'observatorium-minio',
  bucketSecretNamespace: dev.config.namespace,
});

dev.manifests
{
  'minio-deployment': minio.deployment,
  'minio-pvc': minio.pvc,
  'minio-secret-thanos': minio.secretThanos,
  'minio-secret-loki': minio.secretLoki,
  'minio-service': minio.service,
}
