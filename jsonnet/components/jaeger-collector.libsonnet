local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local jaegerAgent = import './jaeger-agent.libsonnet';

{
  jaeger+:: {
    local j = self,
    namespace:: error 'must set namespace for jaeger',
    image:: error 'must set image for jaeger',
    replicas:: 1,
    pvc:: {
      class: 'standard',
      size: '50Gi',
    },

    headlessService:
      service.new(
        'jaeger-collector-headless',
        $.jaeger.deployment.metadata.labels,
        [
          service.mixin.spec.portsType.newNamed('grpc', 14250, 14250),
        ],
      ) +
      service.mixin.metadata.withNamespace(j.namespace) +
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
      service.mixin.metadata.withNamespace(j.namespace) +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }),

    adminService:
      service.new(
        'jaeger-admin',
        $.jaeger.deployment.metadata.labels,
        [
          service.mixin.spec.portsType.newNamed('admin-http', 14269, 14269),
        ],
      ) +
      service.mixin.metadata.withNamespace(j.namespace) +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }),

    agentService:
      service.new(
        'jaeger-agent-discovery',
        { 'app.kubernetes.io/tracing': 'jaeger-agent' },
        [
          service.mixin.spec.portsType.newNamed('metrics', 14271, 14271),
        ],
      ) +
      service.mixin.metadata.withNamespace(j.namespace) +
      service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': 'jaeger-agent' }),

    volumeClaim:
      local claim = k.core.v1.persistentVolumeClaim;
      claim.new() +
      claim.mixin.metadata.withName('jaeger-store-data') +
      claim.mixin.metadata.withNamespace(j.namespace) +
      claim.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }) +
      claim.mixin.spec.withAccessModes('ReadWriteOnce') +
      claim.mixin.spec.resources.withRequests({ storage: j.pvc.size },) +
      claim.mixin.spec.withStorageClassName(j.pvc.class),

    deployment:
      local deployment = k.apps.v1.deployment;
      local container = deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;
      local env = container.envType;
      local mount = container.volumeMountsType;
      local volume = k.apps.v1.statefulSet.mixin.spec.template.spec.volumesType;

      local c =
        container.new($.jaeger.deployment.metadata.name, j.image) +
        container.withArgs([
          '--collector.queue-size=4000',
        ],) +
        container.withEnv([
          env.new('SPAN_STORAGE_TYPE', 'memory'),
        ]) + container.withPorts(
          [
            containerPort.newNamed(14250, 'grpc'),
            containerPort.newNamed(14269, 'admin-http'),
            containerPort.newNamed(16686, 'query'),
          ],
        ) +
        container.withVolumeMounts([mount.new('jaeger-store-data', '/var/jaeger/store')]) +
        container.mixin.readinessProbe.withFailureThreshold(3) +
        container.mixin.readinessProbe.withPeriodSeconds(30) +
        container.mixin.readinessProbe.withInitialDelaySeconds(10) +
        container.mixin.readinessProbe.httpGet.withPath('/').withPort(14269).withScheme('HTTP') +
        container.mixin.livenessProbe.withPeriodSeconds(30) +
        container.mixin.livenessProbe.withFailureThreshold(4) +
        container.mixin.livenessProbe.httpGet.withPath('/').withPort(14269).withScheme('HTTP') +
        container.mixin.resources.withRequests({ cpu: '1', memory: '1Gi' }) +
        container.mixin.resources.withLimits({ cpu: '4', memory: '4Gi' });

      deployment.new('jaeger-all-in-one', j.replicas, c, $.jaeger.deployment.metadata.labels) +
      deployment.mixin.metadata.withNamespace(j.namespace) +
      deployment.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name }) +
      deployment.mixin.spec.selector.withMatchLabels($.jaeger.deployment.metadata.labels) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
      deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
      deployment.mixin.spec.template.spec.withVolumes(volume.fromPersistentVolumeClaim($.jaeger.volumeClaim.metadata.name, $.jaeger.volumeClaim.metadata.name)),
  },
}
