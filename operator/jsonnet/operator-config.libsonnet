local default = import 'default-config.libsonnet';
local cr = import 'generic-operator/config';
local objectStorageConfig = cr.spec.objectStorageConfig;
local thanosObjectStorageConfig = objectStorageConfig.thanos;
local lokiObjectStorageConfig = objectStorageConfig.loki;
local hashrings = cr.spec.hashrings;
cr.spec {
  name: cr.metadata.name,
  namespace: cr.metadata.namespace,
  uid: cr.metadata.uid,
  kind: cr.kind,
  apiVersion: cr.apiVersion,
  compact+:: {
    image: if std.objectHas(cr.spec.compact, 'image') then cr.spec.compact.image else default.compact.image,
    version: if std.objectHas(cr.spec.compact, 'version') then cr.spec.compact.version else default.compact.version,
    objectStorageConfig: thanosObjectStorageConfig,
    logLevel: 'info',
  },
  thanosReceiveController+:: {
    image: if std.objectHas(cr.spec, 'thanosReceiveController') && std.objectHas(cr.spec.thanosReceiveController, 'image') then cr.spec.thanosReceiveController.image else default.thanosReceiveController.image,
    version: if std.objectHas(cr.spec, 'thanosReceiveController') && std.objectHas(cr.spec.thanosReceiveController, 'version') then cr.spec.thanosReceiveController.version else default.thanosReceiveController.version,
    hashrings: hashrings,
  },
  receivers+:: {
    image: if std.objectHas(cr.spec.receivers, 'image') then cr.spec.receivers.image else default.receivers.image,
    version: if std.objectHas(cr.spec.receivers, 'version') then cr.spec.receivers.version else default.receivers.version,
    hashrings: hashrings,
    objectStorageConfig: thanosObjectStorageConfig,
    logLevel: 'info',
    debug: '',
  },
  rule+:: {
    image: if std.objectHas(cr.spec.rule, 'image') then cr.spec.rule.image else default.rule.image,
    version: if std.objectHas(cr.spec.rule, 'version') then cr.spec.rule.version else default.rule.version,
    objectStorageConfig: thanosObjectStorageConfig,
  },
  store+:: {
    image: if std.objectHas(cr.spec.store, 'image') then cr.spec.store.image else default.store.image,
    version: if std.objectHas(cr.spec.store, 'version') then cr.spec.store.version else default.store.version,
    objectStorageConfig: thanosObjectStorageConfig,
    logLevel: 'info',
  },
  storeCache+:: {
    image: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'image') then cr.spec.store.cache.image else default.storeCache.image,
    version: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'version') then cr.spec.store.cache.version else default.storeCache.version,
    exporterImage: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'exporterImage') then cr.spec.store.cache.exporterImage else default.storeCache.exporterImage,
    exporterVersion: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'exporterVersion') then cr.spec.store.cache.exporterVersion else default.storeCache.exporterVersion,
    replicas: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'replicas') then cr.spec.store.cache.replicas else default.storeCache.replicas,
    memoryLimitMb: if std.objectHas(cr.spec.store, 'cache') && std.objectHas(cr.spec.store.cache, 'memoryLimitMb') then cr.spec.store.cache.memoryLimitMb else default.storeCache.memoryLimitMb,
  },
  query+:: {
    image: if std.objectHas(cr.spec, 'query') && std.objectHas(cr.spec.query, 'image') then cr.spec.query.image else default.query.image,
    version: if std.objectHas(cr.spec, 'query') && std.objectHas(cr.spec.query, 'version') then cr.spec.query.version else default.query.version,
  },
  queryFrontend+:: {
    image: if std.objectHas(cr.spec, 'queryFrontend') && std.objectHas(cr.spec.queryFrontend, 'image') then cr.spec.queryFrontend.image else default.queryFrontend.image,
    version: if std.objectHas(cr.spec, 'queryFrontend') && std.objectHas(cr.spec.queryFrontend, 'version') then cr.spec.queryFrontend.version else default.queryFrontend.version,
  },
  apiQuery+:: {
    image: if std.objectHas(cr.spec, 'apiQuery') && std.objectHas(cr.spec.apiQuery, 'image') then cr.spec.apiQuery.image else default.apiQuery.image,
    version: if std.objectHas(cr.spec, 'apiQuery') && std.objectHas(cr.spec.apiQuery, 'version') then cr.spec.apiQuery.version else default.apiQuery.version,
  },
  api+:: {
    image: if std.objectHas(cr.spec, 'api') && std.objectHas(cr.spec.api, 'image') then cr.spec.api.image else default.api.image,
    version: if std.objectHas(cr.spec, 'api') && std.objectHas(cr.spec.api, 'version') then cr.spec.api.version else default.api.version,
    rbac: if std.objectHas(cr.spec, 'api') && std.objectHas(cr.spec.api, 'rbac') then cr.spec.api.rbac else default.api.rbac,
    tenants: if std.objectHas(cr.spec, 'api') && std.objectHas(cr.spec.api, 'tenants') then { tenants: cr.spec.api.tenants } else default.api.tenants,
    tls: if std.objectHas(cr.spec, 'api') && std.objectHas(cr.spec.api, 'tls') then cr.spec.api.tls else {},
  },
  loki+:: if std.objectHas(cr.spec, 'loki') then {
    image: if std.objectHas(cr.spec.loki, 'image') then cr.spec.loki.image else default.loki.image,
    replicas: if std.objectHas(cr.spec.loki, 'replicas') then cr.spec.loki.replicas else default.loki.replicas,
    version: if std.objectHas(cr.spec.loki, 'version') then cr.spec.loki.version else default.loki.version,
    objectStorageConfig: if lokiObjectStorageConfig != null then lokiObjectStorageConfig else default.loki.objectStorageConfig,
  } else {},
}
