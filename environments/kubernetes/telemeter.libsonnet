(import 'telemeter/server/kubernetes.libsonnet') +
(import 'telemeter/prometheus/kubernetes.libsonnet') +
(import 'memcached.libsonnet') +
{
  _config+:: {
    namespace: 'observatorium',
  },

  telemeterServer+:: {
    local image = 'quay.io/app-sre/telemeter:c205c41',
    local memcachedReplicas = std.range(1, $.memcached.replicas),
    statefulSet+: {
      spec+: {
        replicas: 3,
        template+: {
          spec+: {
            containers: [
              super.containers[0] {
                image: image,
                command+: [
                  '--token-expire-seconds=3600',
                  '--forward-url=http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
                    $.thanos.receive.service.metadata.name,
                    $.thanos.receive.service.metadata.namespace,
                    $.thanos.receive.service.spec.ports[2].port,
                  ],
                ] + [
                  '--memcached=%s-%d.%s.%s.svc.cluster.local:%d' % [
                    $.memcached.statefulSet.metadata.name,
                    i,
                    $.memcached.service.metadata.name,
                    $.memcached.service.metadata.namespace,
                    $.memcached.service.spec.ports[0].port,
                  ]
                  for i in memcachedReplicas
                ],
              },
            ],
          },
        },
      },
    },
  },
}
