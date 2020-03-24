{
  local defaultConfig = self,

  name: 'observatorium-xyz',
  namespace: 'observatorium',
  thanosVersion: 'master-2020-03-24-4bd19b16',
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

  apiGateway: {
    local gwConfig = self,
    version: 'master-2020-03-19-62d4b8e',
    image: 'quay.io/observatorium/observatorium:' + gwConfig.version,
  },

  apiGatewayQuery: {
    image: defaultConfig.thanosImage,
    version: defaultConfig.thanosVersion,
  },
}
