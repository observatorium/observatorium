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
  local m = super.memcached,
  local tsList = list.asList('telemeter', ts, [])
                 + list.withAuthorizeURL($._config)
                 + list.withNamespace($._config)
                 + list.withServerImage($._config)
                 + list.withResourceRequestsAndLimits('telemeter-server', $._config.telemeterServer.resourceRequests, $._config.telemeterServer.resourceLimits),
  local mList = list.asList('memcached', m, [
                  {
                    name: 'MEMCACHED_IMAGE',
                    value: m.images.memcached,
                  },
                  {
                    name: 'MEMCACHED_IMAGE_TAG',
                    value: m.tags.memcached,
                  },
                  {
                    name: 'MEMCACHED_EXPORTER_IMAGE',
                    value: m.images.exporter,
                  },
                  {
                    name: 'MEMCACHED_EXPORTER_IMAGE_TAG',
                    value: m.tags.exporter,
                  },
                ])
                + list.withResourceRequestsAndLimits('memcached', $.memcached.resourceRequests, $.memcached.resourceLimits)
                + list.withNamespace($._config),

  telemeterServer+:: {
    list: list.asList('telemeter', {}, []) + {
      objects:
        tsList.objects +
        mList.objects,

      parameters:
        tsList.parameters +
        mList.parameters,
    },
  },

  _config+:: {
    telemeterServer+: {
      whitelist+: (import 'telemeter/metrics.jsonnet'),
      elideLabels+: [
        'prometheus_replica',
      ],
    },
  },
  memcached+:: {
    images:: {
      memcached: '${MEMCACHED_IMAGE}',
      exporter: '${MEMCACHED_EXPORTER_IMAGE}',
    },
    tags:: {
      memcached: '${MEMCACHED_IMAGE_TAG}',
      exporter: '${MEMCACHED_EXPORTER_IMAGE_TAG}',
    },
  },
  apiVersion: 'v1',
  kind: 'Template',
  metadata: {
    name: 'observatorium-telemeter',
  },
  objects: $.telemeterServer.list.objects,
  parameters: $.telemeterServer.list.parameters,
}
