{
  local defaultConfig = self,

  name: 'observatorium-xyz',
  namespace: 'observatorium',
  objectStorageConfig: {
    thanos: {
      name: 'thanos-objectstorage',
      key: 'thanos.yaml',
    },
    loki: {
      secretName: 'loki-objectstorage',
      endpointKey: 'endpoint',
      bucketsKey: 'buckets',
      regionKey: 'region',
      accessKeyIdKey: 'aws_access_key_id',
      secretAccessKeyKey: 'aws_secret_access_key',
    },
  },

  lokiRingStore: {
    local lokiRingStoreConfig = self,
    version: 'v3.4.9',
    image: 'quay.io/coreos/etcd:' + lokiRingStoreConfig.version,
    replicas: 1,
  },

  lokiCaches: {
    local scConfig = self,
    version: '1.6.3-alpine',
    image: 'docker.io/memcached:' + scConfig.version,
    exporterVersion: 'v0.6.0',
    exporterImage: 'prom/memcached-exporter:' + scConfig.exporterVersion,
    replicas: {
      chunk_cache: 1,
      index_query_cache: 1,
      results_cache: 1,
    },
  },

  loki+: {
    local lokiConfig = self,
    version: '2.0.0',
    image: 'docker.io/grafana/loki:' + lokiConfig.version,
    objectStorageConfig: defaultConfig.objectStorageConfig.loki,
    replicas: {
      compactor: 1,
      distributor: 1,
      ingester: 1,
      querier: 1,
      query_frontend: 1,
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
  },
}
