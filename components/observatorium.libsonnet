local t = (import 'kube-thanos/thanos.libsonnet');
local l = (import './loki.libsonnet');
local thanosReceiveController = (import 'thanos-receive-controller/thanos-receive-controller.libsonnet');
local memcached = (import 'memcached.libsonnet');
local observatoriumAPI = (import 'observatorium/observatorium-api.libsonnet');

{
  local obs = self,

  config:: {
    commonLabels:: {
      'app.kubernetes.io/part-of': 'observatorium',
      'app.kubernetes.io/instance': obs.config.name,
    },
    replicaLabels:: ['prometheus_replica', 'rule_replica', 'replica'],
    deduplicationReplicaLabels:: ['replica'],

    podLabelSelector:: {
      [labelName]: obs.config.commonLabels[labelName]
      for labelName in std.objectFields(obs.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  compact:: t.compact({
    name: '%s-thanos-compact' % obs.config.name,
    namespace: obs.config.namespace,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
    deduplicationReplicaLabels: obs.config.deduplicationReplicaLabels,
    deleteDelay: '48h',
    disableDownsampling: true,
  }),

  thanosReceiveController:: thanosReceiveController {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
      namespace: obs.config.namespace,
      replicas: 1,
      commonLabels+:: obs.config.commonLabels,
    },
  },

  receiversService:: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: '%s-thanos-receive' % obs.config.name,
      namespace: obs.config.namespace,
      labels: obs.config.commonLabels { 'app.kubernetes.io/name': 'thanos-receive' },
    },
    spec: {
      selector: { 'app.kubernetes.io/name': 'thanos-receive' },
      ports: [
        { name: 'grpc', port: 10901, targetPort: 10901 },
        { name: 'http', port: 10902, targetPort: 10902 },
        { name: 'remote-write', port: 19291, targetPort: 19291 },
      ],
    },
  },

  receivers:: {
    [hashring.hashring]:
      t.receive({
        name: '%s-thanos-receive-%s' % [obs.config.name, hashring.hashring],
        namespace: obs.config.namespace,
        replicas: 1,
        replicaLabels: obs.config.replicaLabels,
        replicationFactor: 1,
        retention: '4d',
        hashringConfigMapName: '%s-generated' % obs.thanosReceiveController.configmap.metadata.name,
        commonLabels+::
          obs.config.commonLabels {
            'controller.receive.thanos.io/hashring': hashring.hashring,
          },
      }) +
      {
        podDisruptionBudget:: {},  // hide this object, we don't want it
        statefulSet+: {
          metadata+: {
            labels+: {
              'controller.receive.thanos.io': 'thanos-receive-controller',
            },
          },
          spec+: {
            template+: {
              spec+: {
                containers: [
                  if c.name == 'thanos-receive' then c {
                    env+: [
                      {
                        name: 'DEBUG',
                        value: obs.config.receivers.debug,
                      },
                    ],
                  }
                  else c
                  for c in super.containers
                ],
              },
            },
          },
        },
      }
    for hashring in obs.config.hashrings
  },

  rule:: t.rule({
    name: obs.config.name + '-' + 'thanos-rule',
    namespace: obs.config.namespace,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
    queriers: ['dnssrv+_http._tcp.%s.%s.svc.cluster.local' % [obs.query.service.metadata.name, obs.query.service.metadata.namespace]],
  }),

  store:: {
    ['shard' + i]: t.store({
      name: '%s-thanos-store-shard-%d' % [obs.config.name, i],
      namespace: obs.config.namespace,
      commonLabels+:: obs.config.commonLabels {
        'store.observatorium.io/shard': 'shard-' + i,
      },
      replicas: 1,
      ignoreDeletionMarksDelay: '24h',

      local memcachedDefaults = {
        timeout: '2s',
        max_idle_connections: 1000,
        max_async_concurrency: 100,
        max_async_buffer_size: 100000,
        max_get_multi_concurrency: 900,
        max_get_multi_batch_size: 1000,
      },
      indexCache: {
        type: 'memcached',
        config+: memcachedDefaults {
          addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [obs.storeCache.service.metadata.name, obs.storeCache.service.metadata.namespace]],
        },
      },
      bucketCache: {
        type: 'memcached',
        config+: memcachedDefaults {
          addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [obs.storeCache.service.metadata.name, obs.storeCache.service.metadata.namespace]],
        },
      },
    }) + {
      statefulSet+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                if c.name == 'thanos-store' then c {
                  args+: [
                    |||
                      --selector.relabel-config=
                        - action: hashmod
                          source_labels: ["__block_id"]
                          target_label: shard
                          modulus: %d
                        - action: keep
                          source_labels: ["shard"]
                          regex: %d
                    ||| % [obs.config.store.shards, i],
                  ],
                } else c
                for c in super.containers
              ],
            },
          },
        },
      },
    }
    for i in std.range(0, obs.config.store.shards - 1)
  },

  storeCache:: memcached({
    local cfg = self,
    name: obs.config.name + '-thanos-store-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    commonLabels+:: obs.config.commonLabels,
    cpuRequest:: '100m',
    cpuLimit:: '200m',
    memoryRequestBytes: 128 * 1024 * 1024,
    memoryLimitBytes: 256 * 1024 * 1024,
  }),

  query:: t.query({
    name: '%s-thanos-query' % obs.config.name,
    namespace: obs.config.namespace,
    commonLabels+:: obs.config.commonLabels,
    replicas: 1,
    queryTimeout: '15m',
    stores: [
      'dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [service.metadata.name, service.metadata.namespace]
      for service in
        [obs.rule.service] +
        [obs.store[shard].service for shard in std.objectFields(obs.store)] +
        [obs.receivers[hashring].service for hashring in std.objectFields(obs.receivers)]
    ],
    replicaLabels: obs.config.replicaLabels,
  }),

  queryFrontend:: t.queryFrontend({
    name: '%s-thanos-query-frontend' % obs.config.name,
    namespace: obs.config.namespace,
    commonLabels+:: obs.config.commonLabels,
    replicas: 1,
    downstreamURL: 'http://%s.%s.svc.cluster.local.:%d' % [
      obs.query.service.metadata.name,
      obs.query.service.metadata.namespace,
      obs.query.service.spec.ports[1].port,
    ],
    splitInterval: '24h',
    maxRetries: 0,
    logQueriesLongerThan: '5s',
    serviceMonitor: true,
    queryRangeCache: {
      type: 'memcached',
      config+: {
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [obs.queryFrontendCache.service.metadata.name, obs.queryFrontendCache.service.metadata.namespace]],
      },
    },
    labelsCache: {
      type: 'memcached',
      config+: {
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [obs.queryFrontendCache.service.metadata.name, obs.queryFrontendCache.service.metadata.namespace]],
      },
    },
  }),

  queryFrontendCache:: memcached({
    local cfg = self,
    name: obs.config.name + '-thanos-query-frontend-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    component: 'query-frontend-cache',
    commonLabels+:: obs.config.commonLabels,
    cpuRequest:: '100m',
    cpuLimit:: '200m',
    memoryRequestBytes: 128 * 1024 * 1024,
    memoryLimitBytes: 256 * 1024 * 1024,
  }),

  gubernator:: (import 'gubernator.libsonnet')({
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    replicas: 2,
    commonLabels+:: obs.config.commonLabels,
  }),

  api:: observatoriumAPI + observatoriumAPI.withRateLimiter {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
      namespace: obs.config.namespace,
      replicas: 1,
      commonLabels+:: obs.config.commonLabels,
      metrics: {
        readEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
          obs.queryFrontend.service.metadata.name,
          obs.queryFrontend.service.metadata.namespace,
          obs.queryFrontend.service.spec.ports[0].port,
        ],
        writeEndpoint: 'http://%s.%s.svc.cluster.local:%d' % [
          obs.receiversService.metadata.name,
          obs.receiversService.metadata.namespace,
          obs.receiversService.spec.ports[2].port,
        ],
      },
      rateLimiter: {
        grpcAddress: '%s.%s.svc.cluster.local:%d' % [
          obs.gubernator.service.metadata.name,
          obs.gubernator.service.metadata.namespace,
          obs.gubernator.config.ports.grpc,
        ],
      },
    } + if std.length(obs.config.loki) != 0 then {
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
    } else {},
  },

  loki::
    l +
    l.withMemberList +
    l.withReplicas +
    l.withVolumeClaimTemplate {
      config+:: {
        local cfg = self,
        name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
        namespace: obs.config.namespace,
        commonLabels+:: obs.config.commonLabels,
        replicas: obs.config.replicas,
      },
    },
} + {
  local obs = self,

  manifests+:: {
    ['thanos-query-' + name]: obs.query[name]
    for name in std.objectFields(obs.query)
    if obs.query[name] != null
  } + {
    ['thanos-query-frontend-' + name]: obs.queryFrontend[name]
    for name in std.objectFields(obs.queryFrontend)
    if obs.queryFrontend[name] != null
  } + {
    ['query-frontend-cache-' + name]: obs.queryFrontendCache[name]
    for name in std.objectFields(obs.queryFrontendCache)
    if obs.queryFrontendCache[name] != null
  } + {
    ['thanos-receive-' + hashring + '-' + name]: obs.receivers[hashring][name]
    for hashring in std.objectFields(obs.receivers)
    for name in std.objectFields(obs.receivers[hashring])
    if obs.receivers[hashring][name] != null
  } + {
    'thanos-receive-service': obs.receiversService,
  } + {
    ['thanos-compact-' + name]: obs.compact[name]
    for name in std.objectFields(obs.compact)
    if obs.compact[name] != null
  } + {
    ['thanos-store-' + shard + '-' + name]: obs.store[shard][name]
    for shard in std.objectFields(obs.store)
    for name in std.objectFields(obs.store[shard])
    if obs.store[shard][name] != null
  } + {
    ['store-cache-' + name]: obs.storeCache[name]
    for name in std.objectFields(obs.storeCache)
    if obs.storeCache[name] != null
  } + {
    ['thanos-rule-' + name]: obs.rule[name]
    for name in std.objectFields(obs.rule)
    if obs.rule[name] != null
  } + {
    ['thanos-receive-controller-' + name]: obs.thanosReceiveController[name]
    for name in std.objectFields(obs.thanosReceiveController)
  } + {
    ['api-' + name]: obs.api[name]
    for name in std.objectFields(obs.api)
    if obs.api[name] != null
  } {
    ['gubernator-' + name]: obs.gubernator[name]
    for name in std.objectFields(obs.gubernator)
    if obs.gubernator[name] != null
  } + {
    ['api-thanos-query-' + name]: obs.apiQuery[name]
    for name in std.objectFields(obs.apiQuery)
  } + if std.length(obs.config.loki) != 0 then {
    ['loki-' + name]: obs.loki.manifests[name]
    for name in std.objectFields(obs.loki.manifests)
  } else {},
}
