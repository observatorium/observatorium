local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local up = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',

    validated:: if
      !std.objectHas(self, 'queryConfig') &&
      !std.objectHas(self, 'writeEndpoint') &&
      !std.objectHas(self, 'readEndpoint')
    then error 'should set one of queryConfig, writeEndpoint or readEndpoint',

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-up',
      'app.kubernetes.io/instance': up.config.name,
      'app.kubernetes.io/version': up.config.version,
      'app.kubernetes.io/component': 'blackbox-prober',
    },

    podLabelSelector:: {
      [labelName]: up.config.commonLabels[labelName]
      for labelName in std.objectFields(up.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      up.config.name,
      up.config.podLabelSelector,
      [
        ports.newNamed('http', 8080, 8080),
      ],
    ) +
    service.mixin.metadata.withNamespace(up.config.namespace) +
    service.mixin.metadata.withLabels(up.config.commonLabels),

  deployment:
    local d = k.apps.v1.deployment;
    local container = d.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;
    local containerVolumeMount = container.volumeMountsType;

    local c =
      container.new('observatorium-up', up.config.image) +
      container.withArgs(
        [
          '--duration=0',
          '--queries-file=/etc/up/queries.yaml',
          '--log.level=debug',
        ]
      ) +
      container.withPorts([
        containerPort.newNamed(8080, 'http'),
      ]);

    d.new() +
    d.mixin.metadata.withName(up.config.name) +
    d.mixin.metadata.withNamespace(up.config.namespace) +
    d.mixin.spec.selector.withMatchLabels(up.config.podLabelSelector) +
    d.mixin.spec.template.metadata.withLabels(up.config.commonLabels) +
    d.mixin.spec.template.spec.withContainers([c]),

  withResources:: {
    local u = self,
    config+:: {
      resources: error 'must provide resources',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-up' then c {
                resources: u.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },

  withReadEndpoint:: {
    local u = self,
    config+:: {
      readEndpoint: error 'must provide read endpoint',
    },
    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-up' then c {
                args+: ['--endpoint-read=' + up.config.readEndpoint],
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },

  withWriteEndpoint:: {
    local u = self,
    config+:: {
      writeEndpoint: error 'must provide write endpoint',
    },
    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-up' then c {
                args+: ['--endpoint-write=' + up.config.writeEndpoint],
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },

  withQuery:: {
    local u = self,
    config+:: {
      queryConfig: error 'must provide query config endpoint',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              c {
                volumeMounts+: [
                  {
                    mountPath: '/etc/up/',
                    name: 'query-config',
                    readOnly: false,
                  },
                ],
              }
              for c in super.containers

            ],
            volumes+: [
              {
                configMap: {
                  name: up.config.name,
                },
                name: 'query-config',
              },
            ],
          },
        },
      },
    },

    configmap+: {
      apiVersion: 'v1',
      data: {
        'queries.yaml': std.manifestYamlDoc(up.config.queryConfig),
      },
      kind: 'ConfigMap',
      metadata: {
        labels: up.config.commonLabels,
        name: up.config.name,
        namespace: up.config.namespace,
      },
    },
  },

  withServiceMonitor:: {
    local u = self,
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: u.config.name,
        namespace: u.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: u.config.podLabelSelector,
        },
        endpoints: [
          { port: 'http' },
        ],
      },
    },
  },

  manifests+:: {
    ['up-' + name]: up[name]
    for name in std.objectFields(up)
  },
}
