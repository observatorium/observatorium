local t = (import 'kube-thanos/thanos.libsonnet');
local rc = (import 'thanos-receive-controller/thanos-receive-controller.libsonnet');
local memcached = (import 'memcached.libsonnet');

// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  name: 'observatorum-xyz',
  namespace: 'observatorium',
  version: 'v0.17.1',
  image: 'quay.io/thanos/thanos:' + defaults.version,
  objectStorageConfig: {
    name: 'thanos-objectstorage',
    key: 'thanos.yaml',
  },
  hashrings: [{
    hashring: 'default',
    tenants: [],
  }],
  stores+: {
    shards: 1,
    volumeClaimTemplate: {
      spec: {
        accessModes: ['ReadWriteOnce'],
        resources: {
          requests: {
            storage: '50Gi',
          },
        },
      },
    },
    serviceMonitor: false,
    ignoreDeletionMarksDelay: '24h',
  },
  replicaLabels: ['prometheus_replica', 'rule_replica', 'replica'],
  deduplicationReplicaLabels: ['replica'],

  receiveController: {
    local rc = self,
    namespace: defaults.namespace,
    commonLabels+:: defaults.commonLabels,
    replicas: 1,
    version: 'master-2020-02-06-b66e0c8',
    image: 'quay.io/observatorium/thanos-receive-controller:' + rc.version,
    hashrings: defaults.hashrings,
  },

  memcached: {
    local memcached = self,
    namespace: defaults.namespace,
    commonLabels+:: defaults.commonLabels,
    version: '1.6.3-alpine',
    image: 'docker.io/memcached:' + memcached.version,
    exporterVersion: 'v0.6.0',
    exporterImage: 'prom/memcached-exporter:' + memcached.exporterVersion,
  },

  compact: {
    replicas: 1,
    volumeClaimTemplate: {
      spec: {
        accessModes: ['ReadWriteOnce'],
        resources: {
          requests: {
            storage: '50Gi',
          },
        },
      },
    },
    disableDownsampling: true,
    deleteDelay: '48h',
    retentionResolutionRaw: '14d',
    retentionResolution5m: '1s',
    retentionResolution1h: '1s',
  },

  receivers: {
    replicas: 1,
    replicationFactor: 1,
    retention: '4d',
    storage: '50Gi',
    serviceMonitor: false,
    volumeClaimTemplate: {
      spec: {
        accessModes: ['ReadWriteOnce'],
        resources: {
          requests: {
            storage: '50Gi',
          },
        },
      },
    },
  },

  rule: {
    replicas: 1,
    volumeClaimTemplate: {
      spec: {
        accessModes: ['ReadWriteOnce'],
        resources: {
          requests: {
            storage: '50Gi',
          },
        },
      },
    },
  },

  storeCache: defaults.memcached {
    replicas: 1,
    cpuRequest:: '50m',
    cpuLimit:: '50m',
    memoryLimitMb: 1024,
    memoryRequestBytes: 128 * 1024 * 1024,
    memoryLimitBytes: 128 * 1024 * 1024,
  },

  query: {
    replicas: 1,
    queryTimeout: '15m',
  },

  queryFrontend: {
    replicas: 1,
  },

  queryFrontendCache: defaults.memcached {
    replicas: 1,
    cpuRequest:: '50m',
    cpuLimit:: '50m',
    memoryLimitMb: 1024,
    memoryRequestBytes: 128 * 1024 * 1024,
    memoryLimitBytes: 128 * 1024 * 1024,
  },

  commonLabels: {
    'app.kubernetes.io/instance': defaults.name,
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local thanos = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,

  // Safety checks for combined config of defaults and params.
  assert std.isObject(thanos.config.objectStorageConfig) : 'objectStorageConfig replicas has to be an object',
  assert std.isArray(thanos.config.hashrings),
  assert std.isObject(thanos.config.stores),

  compact:: t.compact(thanos.config.compact {
    name: '%s-thanos-compact' % thanos.config.name,
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    deduplicationReplicaLabels: thanos.config.deduplicationReplicaLabels,
    objectStorageConfig: thanos.config.objectStorageConfig,
    logLevel: 'info',
  }),

  receiveController:: rc(thanos.config.receiveController {
    local cfg = self,
    name: thanos.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
  }),

  receiversService:: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: '%s-thanos-receive' % thanos.config.name,
      namespace: thanos.config.namespace,
      labels: thanos.config.commonLabels { 'app.kubernetes.io/name': 'thanos-receive' },
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

  receivers:: t.receiveHashrings(thanos.config.receivers {
    hashrings: thanos.config.hashrings,
    name: thanos.config.name + '-thanos-receive',
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    replicaLabels: thanos.config.replicaLabels,
    objectStorageConfig: thanos.config.objectStorageConfig,
    hashringConfigMapName: '%s-generated' % thanos.receiveController.configmap.metadata.name,
    logLevel: 'info',
  }),

  rule:: t.rule(thanos.config.rule {
    name: thanos.config.name + '-' + 'thanos-rule',
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    objectStorageConfig: thanos.config.objectStorageConfig,
    queriers: ['dnssrv+_http._tcp.%s.%s.svc.cluster.local' % [thanos.query.service.metadata.name, thanos.query.service.metadata.namespace]],
  }),

  stores:: t.storeShards(thanos.config.stores {
    name: thanos.config.name + '-thanos-store-shard',
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    objectStorageConfig: thanos.config.objectStorageConfig,
    replicas: 1,
    logLevel: 'info',
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
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [thanos.storeCache.service.metadata.name, thanos.storeCache.service.metadata.namespace]],
      },
    },
    bucketCache: {
      type: 'memcached',
      config+: memcachedDefaults {
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [thanos.storeCache.service.metadata.name, thanos.storeCache.service.metadata.namespace]],
      },
    },
  }),

  storeCache:: memcached(thanos.config.storeCache {
    local cfg = self,
    name: thanos.config.name + '-thanos-store-' + cfg.commonLabels['app.kubernetes.io/name'],
    component: 'store-cache',
  }),

  query:: t.query(thanos.config.query {
    name: '%s-thanos-query' % thanos.config.name,
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    stores: [
      'dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [service.metadata.name, service.metadata.namespace]
      for service in
        [thanos.rule.service] +
        [thanos.stores.shards[shard].service for shard in std.objectFields(thanos.stores.shards)] +
        [thanos.receivers.hashrings[hashring].service for hashring in std.objectFields(thanos.receivers.hashrings)]
    ],
    replicaLabels: thanos.config.replicaLabels,
  }),

  queryFrontend:: t.queryFrontend(thanos.config.queryFrontend {
    name: '%s-thanos-query-frontend' % thanos.config.name,
    namespace: thanos.config.namespace,
    commonLabels+:: thanos.config.commonLabels,
    image: thanos.config.image,
    version: thanos.config.version,
    downstreamURL: 'http://%s.%s.svc.cluster.local.:%d' % [
      thanos.query.service.metadata.name,
      thanos.query.service.metadata.namespace,
      thanos.query.service.spec.ports[1].port,
    ],
    splitInterval: '24h',
    maxRetries: 0,
    logQueriesLongerThan: '5s',
    queryRangeCache: {
      type: 'memcached',
      config+: {
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [thanos.queryFrontendCache.service.metadata.name, thanos.queryFrontendCache.service.metadata.namespace]],
      },
    },
    labelsCache: {
      type: 'memcached',
      config+: {
        addresses: ['dnssrv+_client._tcp.%s.%s.svc' % [thanos.queryFrontendCache.service.metadata.name, thanos.queryFrontendCache.service.metadata.namespace]],
      },
    },
  }),

  queryFrontendCache:: memcached(thanos.config.queryFrontendCache {
    local cfg = self,
    name: thanos.config.name + '-thanos-query-frontend-' + cfg.commonLabels['app.kubernetes.io/name'],
    component: 'query-frontend-cache',
  }),

  manifests:: {
    ['query-' + name]: thanos.query[name]
    for name in std.objectFields(thanos.query)
    if thanos.query[name] != null
  } + {
    ['query-frontend-' + name]: thanos.queryFrontend[name]
    for name in std.objectFields(thanos.queryFrontend)
    if thanos.queryFrontend[name] != null
  } + {
    ['query-frontend-cache-' + name]: thanos.queryFrontendCache[name]
    for name in std.objectFields(thanos.queryFrontendCache)
    if thanos.queryFrontendCache[name] != null
  } + {
    ['receive-' + hashring + '-' + name]: thanos.receivers.hashrings[hashring][name]
    for hashring in std.objectFields(thanos.receivers.hashrings)
    for name in std.objectFields(thanos.receivers.hashrings[hashring])
    if thanos.receivers.hashrings[hashring][name] != null
  } + {
    [if thanos.config.receivers.serviceMonitor == true && thanos.receivers.serviceMonitor != null then 'receive-service-monitor']: thanos.receivers.serviceMonitor,
    'receive-service-account': thanos.receivers.serviceAccount,
    'receive-service': thanos.receiversService,
  } + {
    ['compact-' + name]: thanos.compact[name]
    for name in std.objectFields(thanos.compact)
    if thanos.compact[name] != null
  } + {
    ['store-' + shard + '-' + name]: thanos.stores.shards[shard][name]
    for shard in std.objectFields(thanos.stores.shards)
    for name in std.objectFields(thanos.stores.shards[shard])
    if thanos.stores.shards[shard][name] != null
  } + {
    [if thanos.config.stores.serviceMonitor == true && thanos.stores.serviceMonitor != null then 'store-service-monitor']: thanos.stores.serviceMonitor,
    'store-service-account': thanos.stores.serviceAccount,
  } + {
    ['store-cache-' + name]: thanos.storeCache[name]
    for name in std.objectFields(thanos.storeCache)
    if thanos.storeCache[name] != null
  } + {
    ['rule-' + name]: thanos.rule[name]
    for name in std.objectFields(thanos.rule)
    if thanos.rule[name] != null
  } + {
    ['receive-controller-' + name]: thanos.receiveController[name]
    for name in std.objectFields(thanos.receiveController)
    if thanos.receiveController[name] != null
  },
}
