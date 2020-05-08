local obs = (import '../base/observatorium.jsonnet');
local minio = (import '../../components/minio.libsonnet') + {
  config:: {
    namespace: 'minio',
    bucketSecretNamespace: obs.config.namespace,
  },
};

local up = (import '../../components/up.libsonnet') + {
  config+:: {
    local cfg = self,
    name: obs.config.name + '-' + cfg.commonLabels['app.kubernetes.io/name'],
    namespace: obs.config.namespace,
    replicas: 1,
    commonLabels+:: obs.config.commonLabels,
    version: obs.config.up.version,
    image: obs.config.up.image,
  },
};

obs.manifests +
minio.manifests +
(up +
 up.withReadEndpoint {
   readEndpoint: 'http://%s.%s.svc:9090/api/v1/query' % [obs.queryCache.service.metadata.name, obs.queryCache.service.metadata.namespace],
 } +
 up.withWriteEndpoint {
   writeEndpoint: 'http://%s.%s.svc:9090/api/v1/query' % [obs.queryCache.service.metadata.name, obs.queryCache.service.metadata.namespace],
 }).manifests
