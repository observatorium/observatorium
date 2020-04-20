local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local up = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    readEndpoint: error 'must provide readEndpoint',
    queryConfig: error 'must provide queryConfig',

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

  configmap:
    local configmap = k.core.v1.configMap;

    configmap.new() +
    configmap.mixin.metadata.withName(up.config.name) +
    configmap.mixin.metadata.withNamespace(up.config.namespace) +
    configmap.mixin.metadata.withLabels(up.config.commonLabels) +
    configmap.withData({
      'queries.yaml': std.manifestYamlDoc(up.config.queryConfig),
    }),

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
        ],
        +if std.objectHas(up.config, 'readEndpoint') then
          ['--endpoint-read=' + up.config.readEndpoint] else [],
        +if std.objectHas(up.config, 'writeEndpoint') then
          ['--endpoint-write=' + up.config.writeEndpoint] else []
      ) +
      container.withPorts([
        containerPort.newNamed(8080, 'http'),
      ]) +
      container.withVolumeMounts([
        containerVolumeMount.new('query-config', '/etc/up/'),
      ]);

    d.new() +
    d.mixin.metadata.withName(up.config.name) +
    d.mixin.metadata.withNamespace(up.config.namespace) +
    d.mixin.spec.selector.withMatchLabels(up.config.podLabelSelector) +
    d.mixin.spec.template.metadata.withLabels(up.config.commonLabels) +
    d.mixin.spec.template.spec.withContainers([c]) +
    d.mixin.spec.template.spec.withVolumes([
      { name: 'query-config', configMap: { name: up.configmap.metadata.name } },
    ]),

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
}
