(import 'observatorium/observatorium-api.libsonnet') {
  observatorium+:: {
    local namespace = 'observatorium',
    namespace:: namespace,

    local kt = (import 'kube-thanos.libsonnet') + {
      thanos+:: {
        namespace:: $.observatorium.namespace,
      },
    },

    api+: {
      replicas:: 3,

      deployment+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                super.containers[0]
                {
                  args: [
                    '--web.listen=0.0.0.0:8080',
                    '--metrics.ui.endpoint=http://%s.%s.svc.cluster.local:%d' % [
                      kt.thanos.querier.service.metadata.name,
                      kt.thanos.querier.service.metadata.namespace,
                      kt.thanos.querier.service.spec.ports[1].port,
                    ],
                    '--metrics.query.endpoint=http://%s.%s.svc.cluster.local:%d/api/v1/query' % [
                      kt.thanos.querierCache.service.metadata.name,
                      kt.thanos.querierCache.service.metadata.namespace,
                      kt.thanos.querierCache.service.spec.ports[0].port,
                    ],
                    '--metrics.write.endpoint=http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
                      kt.thanos.receive.service.metadata.name,
                      kt.thanos.receive.service.metadata.namespace,
                      kt.thanos.receive.service.spec.ports[2].port,
                    ],
                    '--log.level=warn',
                  ],
                },
              ],
            },
          },
        },
      },
    },
  },
}
