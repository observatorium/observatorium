local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  memcached+:: {
    name:: 'memcached',
    image:: 'docker.io/memcached:1.5.20-alpine',
    replicas:: 3,

    local namespace = 'observatorium',
    namespace:: namespace,

    service:
      local service = k.core.v1.service;
      local ports = service.mixin.spec.portsType;

      service.new(
        $.memcached.name,
        $.memcached.statefulSet.metadata.labels,
        [
          ports.newNamed('memcached', 11211, 11211),
        ]
      ) +
      service.mixin.metadata.withNamespace(namespace) +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.memcached.service.metadata.name }) +
      service.mixin.spec.withClusterIp('None'),

    statefulSet:
      local sts = k.apps.v1.statefulSet;
      local volume = sts.mixin.spec.template.spec.volumesType;
      local container = sts.mixin.spec.template.spec.containersType;
      local containerEnv = container.envType;
      local containerVolumeMount = container.volumeMountsType;

      local c =
        container.new($.memcached.statefulSet.metadata.name, $.memcached.image) +
        container.withPorts([
          { name: 'memcached', containerPort: $.memcached.service.spec.ports[0].port },
        ]) +
        container.mixin.resources.withRequests({ cpu: '100m', memory: '512Mi' }) +
        container.mixin.resources.withLimits({ cpu: '1', memory: '1Gi' });

      sts.new($.memcached.name, $.memcached.replicas, c, [], $.memcached.statefulSet.metadata.labels) +
      sts.mixin.metadata.withNamespace(namespace) +
      sts.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.memcached.statefulSet.metadata.name }) +
      sts.mixin.spec.withServiceName($.memcached.service.metadata.name) +
      sts.mixin.spec.selector.withMatchLabels($.memcached.statefulSet.metadata.labels) +
      {
        spec+: {
          volumeClaimTemplates:: null,
        },
      },
  },
}
