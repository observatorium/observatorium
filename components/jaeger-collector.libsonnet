local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local jaegerAgent = import './jaeger-agent.libsonnet';

{
  local _self = self,

  config+:: {
    namespace:: error 'must set namespace for jaeger',
    image:: error 'must set image for jaeger',
    name:: error 'must set name for jaeger',
    replicas:: 1,
    pvc:: {
      class: 'standard',
      size: '50Gi',
    },
    commonLabels:: {
      'app.kubernetes.io/name': 'jaeger',
      'app.kubernetes.io/instance': _self.config.name,
      'app.kubernetes.io/component': 'tracing',
    },

    podLabelSelector:: {
      [labelName]: _self.config.commonLabels[labelName]
      for labelName in std.objectFields(_self.config.commonLabels)
    },
  },

  headlessService:
    service.new(
      _self.config.name + '-collector-headless',
      _self.config.podLabelSelector,
      [
        service.mixin.spec.portsType.newNamed('grpc', 14250, 14250),
      ],
    ) +
    service.mixin.metadata.withNamespace(_self.config.namespace) +
    service.mixin.metadata.withLabels(_self.config.commonLabels) +
    service.mixin.spec.withClusterIp('None'),

  queryService:
    service.new(
      _self.config.name + '-query',
      _self.config.podLabelSelector,
      [
        service.mixin.spec.portsType.newNamed('query', 16686, 16686),
      ],
    ) +
    service.mixin.metadata.withNamespace(_self.config.namespace) +
    service.mixin.metadata.withLabels(_self.config.commonLabels),

  adminService:
    service.new(
      _self.config.name + '-admin',
      _self.config.podLabelSelector,
      [
        service.mixin.spec.portsType.newNamed('admin-http', 14269, 14269),
      ],
    ) +
    service.mixin.metadata.withNamespace(_self.config.namespace) +
    service.mixin.metadata.withLabels(_self.config.commonLabels),

  agentService:
    service.new(
      _self.config.name + '-agent-discovery',
      { 'app.kubernetes.io/tracing': 'jaeger-agent' },
      [
        service.mixin.spec.portsType.newNamed('metrics', 14271, 14271),
      ],
    ) +
    service.mixin.metadata.withNamespace(_self.config.namespace) +
    service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': 'jaeger-agent' }),

  volumeClaim:
    local claim = k.core.v1.persistentVolumeClaim;
    claim.new() +
    claim.mixin.metadata.withName('jaeger-store-data') +
    claim.mixin.metadata.withNamespace(_self.config.namespace) +
    claim.mixin.metadata.withLabels(_self.config.commonLabels) +
    claim.mixin.spec.withAccessModes('ReadWriteOnce') +
    claim.mixin.spec.resources.withRequests({ storage: _self.config.pvc.size },) +
    claim.mixin.spec.withStorageClassName(_self.config.pvc.class),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;
    local mount = container.volumeMountsType;
    local volume = k.apps.v1.statefulSet.mixin.spec.template.spec.volumesType;

    local c =
      container.new($.deployment.metadata.name, _self.config.image) +
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

    deployment.new(_self.config.name, _self.config.replicas, c, $.deployment.metadata.labels) +
    deployment.mixin.metadata.withNamespace(_self.config.namespace) +
    deployment.mixin.metadata.withLabels(_self.config.commonLabels) +
    deployment.mixin.spec.selector.withMatchLabels($.deployment.metadata.labels) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
    deployment.mixin.spec.template.spec.withVolumes(volume.fromPersistentVolumeClaim($.volumeClaim.metadata.name, $.volumeClaim.metadata.name)),

  withServiceMonitor:: {
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: _self.config.name,
        namespace: _self.config.namespace,
      },
      spec: {
        namespaceSelector: {
          matchNames: _self.config.namespace,
        },
        selector: {
          matchLabels: _self.config.commonLabels,
        },
        endpoints: [
          { port: 'admin-http' },
        ],
      },
    },
  },

  manifests+:: {
    ['jaeger-' + name]: _self[name]
    for name in std.objectFields(_self)
    if _self[name] != null
  },
}
