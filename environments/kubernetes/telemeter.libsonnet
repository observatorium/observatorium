(import 'telemeter/benchmark/kubernetes.libsonnet') +
(import 'telemeter/prometheus/kubernetes.libsonnet') +
{
  _config+:: {
    namespace: 'observatorium',
  },

  telemeterServer+:: {
    local image = 'quay.io/metalmatze/telemeter:4799e07',

    statefulSet+: {
      spec+: {
        replicas: 3,
        template+: {
          spec+: {
            containers: [
              super.containers[0] {
                image: image,
                command+: [
                  '--forward-url=http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
                    $.thanos.receive.service.metadata.name,
                    $.thanos.receive.service.metadata.namespace,
                    $.thanos.receive.service.spec.ports[1].port,
                  ],
                ],
              },
            ] + [
              super.containers[1] {
                image: image,
              },
            ],
          },
        },
      },
    },
  },
}
