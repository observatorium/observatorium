local list = import 'telemeter/lib/list.libsonnet';

(import '../kubernetes/telemeter.libsonnet') +
{
  telemeterServer+:: {
    statefulSet+: {
      spec+: {
        replicas: 10,

        template+: {
          spec+: {
            containers: [
              c {
                command: [
                  if std.startsWith(c, '--forward-url=') then '--forward-url=${TELEMETER_FORWARD_URL}' else c
                  for c in super.command
                ],
              }
              for c in super.containers
            ],
          },
        },
      },
    },
  },
} + {
  local ts = super.telemeterServer,
  telemeterServer+:: {
    list: list.asList('telemeter', ts, [])
          + list.withAuthorizeURL($._config)
          + list.withNamespace($._config)
          + list.withServerImage($._config)
          + list.withResourceRequestsAndLimits('telemeter-server', $._config.telemeterServer.resourceRequests, $._config.telemeterServer.resourceLimits),
  },

  _config+:: {
    telemeterServer+: {
      whitelist+: (import 'telemeter/metrics.jsonnet'),
      elideLabels+: [
        'prometheus_replica',
      ],
    },
  },
}
