local cr = import 'generic-operator/config';
local objectStorageConfig = cr.spec.objectStorageConfig;
local hashrings = cr.spec.hashrings;
cr.spec {
  name: cr.metadata.name,
  namespace: cr.metadata.namespace,
  uid: cr.metadata.uid,
  kind: cr.kind,
  apiVersion: cr.apiVersion,
  compact+:: {
    image: cr.spec.compact.image,
    version: cr.spec.compact.version,
    objectStorageConfig: objectStorageConfig,
  },
  thanosReceiveController+:: {
    image: cr.spec.thanosReceiveController.image,
    hashrings: hashrings,
  },
  receivers+:: {
    image: cr.spec.receivers.image,
    version: cr.spec.receivers.version,
    hashrings: hashrings,
    objectStorageConfig: objectStorageConfig,
  },
  rule+:: {
    image: cr.spec.rule.image,
    version: cr.spec.rule.version,
    objectStorageConfig: objectStorageConfig,
  },
  store+:: {
    image: cr.spec.store.image,
    version: cr.spec.store.version,
    objectStorageConfig: objectStorageConfig,
  },
  storeCache+:: {
    image: cr.spec.store.cache.image,
    version: cr.spec.store.cache.version,
    exporterImage: cr.spec.store.cache.exporterImage,
    exporterVersion: cr.spec.store.cache.exporterVersion,
    replicas: cr.spec.store.cache.replicas,
    memoryLimitMb: cr.spec.store.cache.memoryLimitMb,
  },
  query+:: {
    image: cr.spec.query.image,
    version: cr.spec.query.version,
  },
  queryCache+:: {
    image: cr.spec.queryCache.image,
  },
  apiQuery+:: {
    image: cr.spec.apiQuery.image,
    version: cr.spec.apiQuery.version,
  },
  api+:: {
    image: cr.spec.api.image,
  },
}
