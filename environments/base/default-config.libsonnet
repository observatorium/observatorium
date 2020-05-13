{
  local defaultConfig = self,

  name: 'observatorium-xyz',
  namespace: 'observatorium',
  thanosVersion: 'v0.12.0-rc.1',
  thanosImage: 'quay.io/thanos/thanos:' + defaultConfig.thanosVersion,
  objectStorageConfig: {
    name: 'thanos-objectstorage',
    key: 'thanos.yaml',
  },

  hashrings: [
    {
      hashring: 'default',
      tenants: [
        // Match all for now
        // 'foo',
        // 'bar',
      ],
    },
  ],

  compact: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
    objectStorageConfig: defaultConfig.objectStorageConfig,
    retentionResolutionRaw: '14d',
    retentionResolution5m: '1s',
    retentionResolution1h: '1s',
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

  thanosReceiveController: {
    local trcConfig = self,
    version: 'master-2020-02-06-b66e0c8',
    image: 'quay.io/observatorium/thanos-receive-controller:' + trcConfig.version,
    hashrings: defaultConfig.hashrings,
  },

  receivers: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
    hashrings: defaultConfig.hashrings,
    objectStorageConfig: defaultConfig.objectStorageConfig,
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
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
    objectStorageConfig: defaultConfig.objectStorageConfig,
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

  store: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
    objectStorageConfig: defaultConfig.objectStorageConfig,
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
  },

  storeCache: {
    local scConfig = self,
    replicas: 1,
    version: '1.6.3-alpine',
    image: 'docker.io/memcached:' + scConfig.version,
    exporterVersion: 'v0.6.0',
    exporterImage: 'prom/memcached-exporter:' + scConfig.exporterVersion,
    memoryLimitMb: 1024,
  },

  query: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
  },

  queryCache: {
    local qcConfig = self,
    replicas: 1,
    version: 'master-fdcd992f',
    image: 'quay.io/cortexproject/cortex:' + qcConfig.version,
  },

  api: {
    local gwConfig = self,
    version: 'master-2020-05-11-v0.1.1-35-g3c24d9b',
    image: 'quay.io/observatorium/observatorium:' + gwConfig.version,
  },

  apiQuery: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
  },

  up: {
    local upConfig = self,
    version: 'master-2020-03-31-6e67351',
    image: 'quay.io/observatorium/up:' + upConfig.version,
  },
}
