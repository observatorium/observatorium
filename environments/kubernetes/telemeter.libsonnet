(import 'telemeter/server/kubernetes.libsonnet') +
(import 'telemeter/prometheus/kubernetes.libsonnet') +
{
  _config+:: {
    namespace: 'observatorium',
  },

  telemeterServer+:: {
    local image = 'quay.io/app-sre/telemeter:c205c41',

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
                    'thanos-receive',
                    '${NAMESPACE}',
                    19291,
                  ],
                ],
              },
            ],
          },
        },
      },
    },
  },
}
