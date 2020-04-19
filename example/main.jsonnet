local obs = (import '../jsonnet/environments/base/observatorium.jsonnet');

{
  local cr = self,
  name:: 'observatorium-cr',

  apiVersion: 'core.observatorium.io/v1alpha1',
  kind: 'Observatorium',
  metadata: obs.config.commonLabels {
    name: obs.config.name,
    'app.kubernetes.io/name': cr.name,
    'app.kubernetes.io/component': cr.name,
  },
  spec: {
    thanosImage: obs.config.thanosImage,
    thanosVersion: obs.config.thanosVersion,
    objectStorageConfig: obs.config.objectStorageConfig,
    hashrings: obs.config.hashrings,

    queryCache: {
      image: obs.config.queryCache.image,
      replicas: obs.config.queryCache.replicas,
      version: obs.config.queryCache.version,
    },
    store: {
      volumeClaimTemplate: obs.config.store.volumeClaimTemplate,
      shards: obs.config.store.shards,
      cache: {
        image: obs.config.storeCache.image,
        version: obs.config.storeCache.version,
        exporterImage: obs.config.storeCache.exporterImage,
        exporterVersion: obs.config.storeCache.exporterVersion,
        replicas: obs.config.storeCache.replicas,
        memoryLimitMb: obs.config.storeCache.memoryLimitMb,
      },
    },
    compact: {
      volumeClaimTemplate: obs.config.compact.volumeClaimTemplate,
      retentionResolutionRaw: obs.config.compact.retentionResolutionRaw,
      retentionResolution5m: obs.config.compact.retentionResolution5m,
      retentionResolution1h: obs.config.compact.retentionResolution1h,
    },
    rule: {
      volumeClaimTemplate: obs.config.rule.volumeClaimTemplate,
    },
    receivers: {
      volumeClaimTemplate: obs.config.receivers.volumeClaimTemplate,
    },
    thanosReceiveController: {
      image: obs.config.thanosReceiveController.image,
      version: obs.config.thanosReceiveController.version,
    },
    apiGateway: {
      image: obs.config.apiGateway.image,
      version: obs.config.apiGateway.version,
    },
  },
}
