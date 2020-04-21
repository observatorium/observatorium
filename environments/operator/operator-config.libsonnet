local cr = import 'generic-operator/config';
local thanosImage = cr.spec.thanosImage;
local thanosVersion = cr.spec.thanosVersion;
local objectStorageConfig = cr.spec.objectStorageConfig;
local hashrings = cr.spec.hashrings;
cr.spec {
  name: cr.metadata.name,
  namespace: cr.metadata.namespace,
  compact+:: {
    image: thanosImage,
    version: thanosVersion,
    objectStorageConfig: objectStorageConfig,
  },
  thanosReceiveController+:: {
    image: cr.spec.thanosReceiveController.image,
    hashrings: hashrings,
  },
  receivers+:: {
    image: thanosImage,
    version: thanosVersion,
    hashrings: hashrings,
    objectStorageConfig: objectStorageConfig,
  },
  rule+:: {
    image: thanosImage,
    version: thanosVersion,
    objectStorageConfig: objectStorageConfig,
  },
  store+:: {
    image: thanosImage,
    version: thanosVersion,
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
    image: thanosImage,
    version: thanosVersion,
  },
  queryCache+:: {
    image: cr.spec.queryCache.image,
  },
  apiGateway+:: {
    image: cr.spec.apiGateway.image,
  },
}
