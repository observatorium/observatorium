(import 'telemeter/benchmark/kubernetes.libsonnet') + {
  _config+:: {
    namespace: 'observatorium',
  },

  telemeterServer+:: {
    statefulSet+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              super.containers[0] {
                command+: [
                  '--forward-url=http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
                    $.thanos.receive.service.metadata.name,
                    $.thanos.receive.service.metadata.namespace,
                    $.thanos.receive.service.spec.ports[1].port,
                  ],
                ],
              },
            ] + [
              super.containers[1],
            ],
          },
        },
      },
    },
  },
}
