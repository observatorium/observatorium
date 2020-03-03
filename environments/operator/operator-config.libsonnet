local cr = import 'generic-operator/config';
local thanosImage = cr.spec.thanosImage + cr.spec.thanosVersion;
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
    image: cr.spec.thanosReceiveController.image + cr.spec.thanosReceiveController.version,
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
  query: {
    image: thanosImage,
    version: thanosVersion,
  },
  queryCache+:: {
    image: cr.spec.queryCache.image + cr.spec.queryCache.version,
  },
  apiGatewayQuery: {
    image: thanosImage,
    version: thanosVersion,
  },
  apiGateway+:: {
    image: cr.spec.apiGateway.image + cr.spec.apiGateway.version,
  },
}
