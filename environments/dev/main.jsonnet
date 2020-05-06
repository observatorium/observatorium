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
    writeEndpoint: 'http://%s.%s.svc:9090/api/v1/query' % [obs.queryCache.service.metadata.name, obs.queryCache.service.metadata.namespace],
    readEndpoint: 'http://%s.%s.svc:9090/api/v1/query' % [obs.queryCache.service.metadata.name, obs.queryCache.service.metadata.namespace],
    queryConfig: {
      name: 'Clusters',
      query: 'avg_over_time(sum(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    version: obs.config.up.version,
    image: obs.config.up.image,
  },
};

obs.manifests +
minio.manifests +
(up + up.withReadEndpoint + up.withWriteEndpoint).manifests
