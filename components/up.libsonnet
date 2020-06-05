local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local up = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    endpointType: error 'must provide endpoint type',
    queryConfig: {},
    readEndpoint: '',
    writeEndpoint: '',
    logs: '',
    resources: {
      requests: {},
      limits: {},
    },
    serviceMonitor: false,

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
          '--log.level=debug',
          '--endpoint-type=' + up.config.endpointType,
        ] + (
          if up.config.queryConfig != {} then
            ['--queries-file=/etc/up/queries.yaml']
          else []
        ) + (
          if up.config.readEndpoint != '' then
            ['--endpoint-read=' + up.config.readEndpoint]
          else []
        ) + (
          if up.config.writeEndpoint != '' then
            ['--endpoint-write=' + up.config.writeEndpoint]
          else []
        ) + (
          if up.config.logs != '' then
            ['--logs=' + up.config.logs]
          else []
        )
      ) +
      container.withPorts([
        containerPort.newNamed(8080, 'http'),
      ]) +
      container.withVolumeMounts(
        if up.config.queryConfig != {} then
          [
            {
              mountPath: '/etc/up/',
              name: 'query-config',
              readOnly: false,
            },
          ] else [],
      ) +
      container.mixin.resources.withLimits(up.config.resources.limits) +
      container.mixin.resources.withRequests(up.config.resources.requests);

    d.new() +
    d.mixin.metadata.withName(up.config.name) +
    d.mixin.metadata.withNamespace(up.config.namespace) +
    d.mixin.spec.selector.withMatchLabels(up.config.podLabelSelector) +
    d.mixin.spec.template.metadata.withLabels(up.config.commonLabels) +
    d.mixin.spec.template.spec.withVolumes(
      if up.config.queryConfig != {} then
        [
          {
            configMap: {
              name: up.config.name,
            },
            name: 'query-config',
          },
        ] else [],
    ) +
    d.mixin.spec.template.spec.withContainers([c]),

  configmap:
    if up.config.queryConfig != {} then {
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
    } else null,

  serviceMonitor:
    if up.config.serviceMonitor == true then
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata+: {
          name: up.config.name,
          namespace: up.config.namespace,
        },
        spec: {
          selector: {
            matchLabels: up.config.podLabelSelector,
          },
          endpoints: [
            { port: 'http' },
          ],
        },
      } else null,

  manifests+:: {
    ['up-' + name]: up[name]
    for name in std.objectFields(up)
    if up[name] != null
  },
}
