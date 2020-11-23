local thanos = (import './thanos.libsonnet');
local loki = (import './loki.libsonnet');
local api = (import 'observatorium/observatorium-api.libsonnet');

{
  local obs = self,

  config:: {
    name: 'observatorium-xyz',
    namespace: 'observatorium',

    loki: {},

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
    stores: {
      shards: 1,
    },
  }),

  loki::
    loki +
    loki.withMemberList +
    loki.withReplicas +
    loki.withDataReplication +
    loki.withVolumeClaimTemplate {
      config+:: {
        local cfg = self,
        name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
        namespace: obs.config.namespace,
        commonLabels+:: obs.config.commonLabels,
        replicas: obs.config.replicas,
        replicationFactor: 1,
      },
    },

  gubernator:: (import 'gubernator.libsonnet')({
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    version: '1.0.0-rc.1',
    image: 'thrawn01/gubernator:' + cfg.version,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
  }),

  api:: api({
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    commonLabels+:: obs.config.commonLabels,
    version: 'master-2020-11-02-v0.1.1-192-ge324057',
    image: 'quay.io/observatorium/observatorium:' + cfg.version,
    replicas: 1,

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
    rateLimiter: {
      grpcAddress: '%s.%s.svc.cluster.local:%d' % [
        obs.gubernator.service.metadata.name,
        obs.gubernator.service.metadata.namespace,
        obs.gubernator.config.ports.grpc,
      ],
    },
  } + if std.objectHas(obs.config, 'loki') && std.length(obs.config.loki) != 0 then {
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
  } else {}),
} + {
  local obs = self,

  manifests+::
    {
      ['gubernator-' + name]: obs.gubernator[name]
      for name in std.objectFields(obs.gubernator)
      if obs.gubernator[name] != null
    } + {
      ['api-' + name]: obs.api[name]
      for name in std.objectFields(obs.api)
      if obs.api[name] != null
    } + {
      ['thanos-' + name]: obs.thanos.manifests[name]
      for name in std.objectFields(obs.thanos.manifests)
    } + if std.length(obs.config.loki) != 0 then {
      ['loki-' + name]: obs.loki.manifests[name]
      for name in std.objectFields(obs.loki.manifests)
    }
    else {},
}
