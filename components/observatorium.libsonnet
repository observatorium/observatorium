local t = (import 'kube-thanos/thanos.libsonnet');
local l = (import './loki.libsonnet');
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local obs = self,

  config:: {
    commonLabels:: {
      'app.kubernetes.io/part-of': 'observatorium',
      'app.kubernetes.io/instance': obs.config.name,
    },
    replicaLabels:: ['prometheus_replica', 'rule_replica', 'replica'],
    deduplicationReplicaLabels:: ['replica'],
  },

  compact::
    t.compact +
    t.compact.withRetention +
    t.compact.withDeleteDelay +
    t.compact.withDeduplication + {
      config+:: {
        local cfg = self,
        name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
        namespace: obs.config.namespace,
        replicas: 1,
        commonLabels+:: obs.config.commonLabels,
        deduplicationReplicaLabels: obs.config.deduplicationReplicaLabels,
        deleteDelay: '48h',
      },
    } + (
      if std.objectHas(obs.config.compact, 'enableDownsampling') && obs.config.compact.enableDownsampling == true then
        {}
      else t.compact.withDownsamplingDisabled
    ),

  thanosReceiveController:: (import 'thanos-receive-controller/thanos-receive-controller.libsonnet') + {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
      namespace: obs.config.namespace,
      replicas: 1,
      commonLabels+:: obs.config.commonLabels,
    },
  },

  receivers:: {
    [hashring.hashring]:
      t.receive +
      t.receive.withRetention +
      t.receive.withHashringConfigMap + {
        config+:: {
          local cfg = self,
          name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'] + '-' + hashring.hashring,
          namespace: obs.config.namespace,
          replicas: 1,
          replicationFactor: 1,
          retention: '4d',
          hashringConfigMapName: '%s-generated' % obs.thanosReceiveController.configmap.metadata.name,
          commonLabels+::
            obs.config.commonLabels {
              'controller.receive.thanos.io/hashring': hashring.hashring,
            },
        },
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

  rule:: t.rule {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
      namespace: obs.config.namespace,
      replicas: 1,
      commonLabels+:: obs.config.commonLabels,
    },
  },

  store:: {
    ['shard' + i]:
      t.store +
      t.store.withIndexCacheMemcached +
      t.store.withCachingBucketMemcached +
      t.store.withIgnoreDeletionMarksDelay {
        config+:: {
          local cfg = self,
          name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'] + '-shard-' + i,
          namespace: obs.config.namespace,
          commonLabels+:: obs.config.commonLabels {
            'store.observatorium.io/shard': 'shard-' + i,
          },
          replicas: 1,
          ignoreDeletionMarksDelay: '24h',
          memcached+: {
            addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [obs.storeCache.service.metadata.name, obs.storeCache.service.metadata.namespace]],
            timeout: '2s',
            maxIdleConnections: 1000,
            maxAsyncConcurrency: 100,
            maxAsyncBufferSize: 100000,
            maxGetMultiConcurrency: 900,
            maxGetMultiBatchSize: 1000,
          },
        },
      } + {
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

  storeCache:: (import 'memcached.libsonnet') + {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-thanos-store-' + cfg.commonLabels['app.kubernetes.io/name'],
      namespace: obs.config.namespace,
      commonLabels+:: obs.config.commonLabels,
      cpuRequest:: '100m',
      cpuLimit:: '200m',
      memoryRequestBytes: 128 * 1024 * 1024,
      memoryLimitBytes: 256 * 1024 * 1024,
    },
  },

  query:: t.query + t.query.withQueryTimeout {
    config+:: {
      local cfg = self,
      name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
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
    },
  },

  queryFrontend::
    t.queryFrontend +
    t.queryFrontend.withServiceMonitor +
    t.queryFrontend.withSplitInterval +
    t.queryFrontend.withMaxRetries +
    t.queryFrontend.withLogQueriesLongerThan +
    t.queryFrontend.withInMemoryResponseCache + {
      config+:: {
        local cfg = self,
        name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
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
      },
    },

  receiveService::
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      obs.config.name + '-thanos-receive',
      { 'app.kubernetes.io/name': 'thanos-receive' },
      [
        ports.newNamed('grpc', 10901, 10901),
        ports.newNamed('http', 10902, 10902),
        ports.newNamed('remote-write', 19291, 19291),
      ]
    ) +
    service.mixin.metadata.withNamespace(obs.config.namespace),

  api:: (import 'observatorium/observatorium-api.libsonnet') + {
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
          obs.receiveService.metadata.name,
          obs.receiveService.metadata.namespace,
          obs.receiveService.spec.ports[2].port,
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

  rule+:: {
    config+:: {
      queriers: ['dnssrv+_http._tcp.%s.%s.svc.cluster.local' % [obs.query.service.metadata.name, obs.query.service.metadata.namespace]],
    },
  },

  manifests+:: {
    ['thanos-query-' + name]: obs.query[name]
    for name in std.objectFields(obs.query)
  } + {
    ['thanos-query-frontend-' + name]: obs.queryFrontend[name]
    for name in std.objectFields(obs.queryFrontend)
  } + {
    ['thanos-receive-' + hashring + '-' + name]: obs.receivers[hashring][name]
    for hashring in std.objectFields(obs.receivers)
    for name in std.objectFields(obs.receivers[hashring])
  } + {
    'thanos-receive-service': obs.receiveService,
  } + {
    ['thanos-compact-' + name]: obs.compact[name]
    for name in std.objectFields(obs.compact)
  } + {
    ['thanos-store-' + shard + '-' + name]: obs.store[shard][name]
    for shard in std.objectFields(obs.store)
    for name in std.objectFields(obs.store[shard])
  } + {
    ['store-cache-' + name]: obs.storeCache[name]
    for name in std.objectFields(obs.storeCache)
  } + {
    ['thanos-rule-' + name]: obs.rule[name]
    for name in std.objectFields(obs.rule)
  } + {
    ['thanos-receive-controller-' + name]: obs.thanosReceiveController[name]
    for name in std.objectFields(obs.thanosReceiveController)
  } + {
    ['api-' + name]: obs.api[name]
    for name in std.objectFields(obs.api)
    if obs.api[name] != null
  } + {
    ['api-thanos-query-' + name]: obs.apiQuery[name]
    for name in std.objectFields(obs.apiQuery)
  } + {
    ['loki-' + name]: obs.loki.manifests[name]
    for name in std.objectFields(obs.loki.manifests)
    if std.length(obs.config.loki) != 0
  },
}
