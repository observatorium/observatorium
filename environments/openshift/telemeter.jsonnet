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
  memcached+:: {
    service+: {
      metadata+: {
        namespace: '${NAMESPACE}',
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
                    value: 'docker.io/memcached',
                  },
                  {
                    name: 'MEMCACHED_IMAGE_TAG',
                    value: '1.5.20-alpine',
                  },
                ])
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
    image:: '${MEMCACHED_IMAGE}:${MEMCACHED_IMAGE_TAG}',
  },


  apiVersion: 'v1',
  kind: 'Template',
  metadata: {
    name: 'observatorium-telemeter',
  },
  objects: $.telemeterServer.list.objects,
  parameters: $.telemeterServer.list.parameters,
}
