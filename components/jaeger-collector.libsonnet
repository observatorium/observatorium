local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;

{
  jaeger+:: {
    headlessService:
      service.new(
        'jaeger-collector-headless',
        $.jaeger.deployment.metadata.labels,
        [
          service.mixin.spec.portsType.newNamed('grpc', 14250, 14250),
        ],
      ) +
      service.mixin.metadata.withNamespace('observatorium') +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }) +
      service.mixin.spec.withClusterIp('None'),

    queryService:
      service.new(
        'jaeger-query',
        $.jaeger.deployment.metadata.labels,
        [
          service.mixin.spec.portsType.newNamed('query', 16686, 16686),
        ],
      ) +
      service.mixin.metadata.withNamespace('observatorium') +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }),

    volumeClaim:
      local claim = k.core.v1.persistentVolumeClaim;
      claim.new() +
      claim.mixin.metadata.withName('jaeger-store-data') +
      claim.mixin.metadata.withNamespace('observatorium') +
      claim.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }) +
      claim.mixin.spec.withAccessModes('ReadWriteOnce') +
      claim.mixin.spec.resources.withRequests({ storage: '50Gi' },) +
      claim.mixin.spec.withStorageClassName('standard'),

    deployment:
      local deployment = k.apps.v1.deployment;
      local container = deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;
      local env = container.envType;
      local mount = container.volumeMountsType;
      local volume = k.apps.v1.statefulSet.mixin.spec.template.spec.volumesType;

      local c =
        container.new($.jaeger.deployment.metadata.name, 'jaegertracing/all-in-one:1.14.0') +
        container.withArgs([
          '--badger.directory-key=/var/jaeger/store/keys',
          '--badger.directory-value=/var/jaeger/store/values',
          '--badger.ephemeral=false',
        ],) +
        container.withEnv([
          env.new('SPAN_STORAGE_TYPE', 'badger'),
        ]) + container.withPorts(
          [
            containerPort.newNamed(14250, 'grpc'),
            containerPort.newNamed(14269, 'admin-http'),
            containerPort.newNamed(16686, 'query'),
          ],
        ) +
        container.withVolumeMounts([mount.new('jaeger-store-data', '/var/jaeger/store')]);

      deployment.new('jaeger-all-in-one', 1, c, $.jaeger.deployment.metadata.labels) +
      deployment.mixin.metadata.withNamespace('observatorium') +
      deployment.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }) +
      deployment.mixin.spec.selector.withMatchLabels($.jaeger.deployment.metadata.labels) +
      deployment.mixin.spec.template.spec.withVolumes(volume.fromPersistentVolumeClaim($.jaeger.volumeClaim.metadata.name, $.jaeger.volumeClaim.metadata.name)),
  },
}
