(import 'observatorium/observatorium-api.libsonnet') {
  observatorium+:: {
    namespace:: 'observatorium',

    local obs = (import 'observatorium.jsonnet') + {
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
                      'thanos-query',
                      $.observatorium.namespace,
                      9090,
                    ],
                    '--metrics.query.endpoint=http://%s.%s.svc.cluster.local:%d/api/v1/query' % [
                      'cortex-query-cache',
                      $.observatorium.namespace,
                      9090,
                    ],
                    '--metrics.write.endpoint=http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
                      'thanos-receive',
                      $.observatorium.namespace,
                      19291,
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
