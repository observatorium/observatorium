local thanos = (import './thanos.libsonnet');
local loki = (import './loki.libsonnet');
local tracing = (import './tracing.libsonnet');
local api = (import 'observatorium-api/observatorium-api.libsonnet');

{
  local obs = self,

  config:: {
    name: 'observatorium-xyz',
    namespace: 'observatorium',

    commonLabels:: {
      'app.kubernetes.io/part-of': 'observatorium',
      'app.kubernetes.io/instance': obs.config.name,
    },
  },

  thanos:: thanos({
    name: obs.config.name,
    namespace: obs.config.namespace,
    commonLabels+: obs.config.commonLabels,
    hashrings: [{
      hashring: 'default',
      tenants: [],
    }],
    stores+: {
      shards: 1,
    },
  }),

  gubernator:: (import 'gubernator.libsonnet')({
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    version: 'v2.0.0-rc.36',
    image: 'ghcr.io/mailgun/gubernator:' + cfg.version,
    imagePullPolicy: 'IfNotPresent',
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
  }),

  api:: api({
    local cfg = self,
    version: 'main-2022-03-03-v0.1.2-139-gb3a1918',
    image: 'quay.io/observatorium/api:' + cfg.version,
    imagePullPolicy: 'IfNotPresent',
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
    metrics: {
      readEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
        obs.thanos.queryFrontend.service.metadata.name,
        obs.thanos.queryFrontend.service.metadata.namespace,
        obs.thanos.queryFrontend.service.spec.ports[0].port,
      ],
      writeEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
        obs.thanos.receiversService.metadata.name,
        obs.thanos.receiversService.metadata.namespace,
        obs.thanos.receiversService.spec.ports[2].port,
      ],
    },
    logs: {
      readEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
        obs.loki.manifests['query-frontend-http-service'].metadata.name,
        obs.loki.manifests['query-frontend-http-service'].metadata.namespace,
        obs.loki.manifests['query-frontend-http-service'].spec.ports[0].port,
      ],
      tailEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
        obs.loki.manifests['querier-http-service'].metadata.name,
        obs.loki.manifests['querier-http-service'].metadata.namespace,
        obs.loki.manifests['querier-http-service'].spec.ports[0].port,
      ],
      writeEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
        obs.loki.manifests['distributor-http-service'].metadata.name,
        obs.loki.manifests['distributor-http-service'].metadata.namespace,
        obs.loki.manifests['distributor-http-service'].spec.ports[0].port,
      ],
    },
    traces: {
      writeEndpoint: obs.tracing.manifests.otelcollector.metadata.name + '-collector:4317',
    },
    rateLimiter: {
      grpcAddress: '%s.%s.svc.cluster.local:%d' % [
        obs.gubernator.service.metadata.name,
        obs.gubernator.service.metadata.namespace,
        obs.gubernator.config.ports.grpc,
      ],
    },
  }),

  loki:: loki({
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    commonLabels+:: obs.config.commonLabels,
    version: '2.7.2',
    image: 'docker.io/grafana/loki:' + cfg.version,
    imagePullPolicy: 'IfNotPresent',
    replicationFactor: 1,
    query+: {
      concurrency: 2,
    },
    replicas: {
      compactor: 1,
      distributor: 1,
      ingester: 1,
      index_gateway: 1,
      querier: 1,
      query_frontend: 1,
      query_scheduler: 1,
      ruler: 1,
    },
    memberlist: {
      ringName: 'gossip-ring',
    },
    wal: {
      replayMemoryCeiling: '100MB',
    },
    volumeClaimTemplate: {
      spec: {
        accessModes: ['ReadWriteOnce'],
        resources: {
          requests: {
            storage: '250Mi',
          },
        },
      },
    },
    objectStorageConfig: {
      secretName: 'loki-objectstorage',
      endpointKey: 'endpoint',
    },
    rulesStorageConfig: {
      type: 's3',
      secretName: 'obs-rules-objectstorage',
      endpointKey: 'endpoint',
    },
  }),

  tracing:: tracing({
    local cfg = self,
    namespace: 'observatorium',
    commonLabels+:: obs.config.commonLabels,
  }),
} + {
  local obs = self,

  manifests+::
    {
      ['observatorium/gubernator-' + name]: obs.gubernator[name]
      for name in std.objectFields(obs.gubernator)
      if obs.gubernator[name] != null
    } + {
      ['observatorium/api-' + name]: obs.api[name]
      for name in std.objectFields(obs.api)
      if obs.api[name] != null
    } + {
      ['observatorium/thanos-' + name]: obs.thanos.manifests[name]
      for name in std.objectFields(obs.thanos.manifests)
    } + (if std.objectHas(obs.loki, 'manifests') then {
           ['observatorium/loki-' + name]: obs.loki.manifests[name]
           for name in std.objectFields(obs.loki.manifests)
         } else {})
    + (if obs.tracing.config.enabled then {
         ['observatorium/tracing-' + name]: obs.tracing.manifests[name]
         for name in std.objectFields(obs.tracing.manifests)
       } else {}),
}
