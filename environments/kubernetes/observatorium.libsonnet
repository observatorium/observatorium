local tenants = import '../../tenants.libsonnet';

(import 'observatorium/observatorium-api.libsonnet') {
  observatorium+:: {
    namespace:: 'observatorium',

    local kt = (import 'kube-thanos.libsonnet') + {
      thanos+:: {
        namespace:: $.observatorium.namespace,

        querier+:: {
          name:: 'observatorium-api-thanos-querier',
        },
      },
    },

    querier+: kt.thanos.querier {
      replicas:: 2,

      deployment+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                super.containers[0]
                {
                  args: [
                    'query',
                    '--query.replica-label=replica',
                    '--query.replica-label=ruler_replica',
                    '--query.replica-label=prometheus_replica',
                    '--web.external-prefix=%s/ui/v1/metrics' % $.observatorium.api.externalURL,
                    '--grpc-address=0.0.0.0:%d' % kt.thanos.querier.service.spec.ports[0].port,
                    '--http-address=0.0.0.0:%d' % kt.thanos.querier.service.spec.ports[1].port,
                    '--store=dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [
                      kt.thanos.store.service.metadata.name,
                      kt.thanos.store.service.metadata.namespace,
                    ],
                    '--store=dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [
                      kt.thanos.ruler.service.metadata.name,
                      kt.thanos.ruler.service.metadata.namespace,
                    ],
                  ] + [
                    '--store=dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [
                      kt.thanos.receive['service-' + tenant.hashring].metadata.name,
                      kt.thanos.receive['service-' + tenant.hashring].metadata.namespace,
                    ]
                    for tenant in tenants
                  ],
                },
              ],
            },
          },
        },
      },
    },

    api+: {
      replicas:: 3,
      externalURL:: 'https://observatorium.api',

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
                      $.observatorium.querier.service.metadata.name,
                      $.observatorium.querier.service.metadata.namespace,
                      $.observatorium.querier.service.spec.ports[1].port,
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
