local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local gateway = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',
    uiEndpoint: error 'must provide uiEndpoint',
    readEnpoint: error 'must provide readEnpoint',
    writeEndpoint: error 'must provide writeEndpoint',

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-api-gateway',
      'app.kubernetes.io/instance': gateway.config.name,
      'app.kubernetes.io/version': gateway.config.version,
      'app.kubernetes.io/component': 'api-gateway',
    },

    podLabelSelector:: {
      [labelName]: gateway.config.commonLabels[labelName]
      for labelName in std.objectFields(gateway.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      gateway.config.name,
      gateway.config.podLabelSelector,
      [
        ports.newNamed('http', 8080, 8080),
      ],
    ) +
    service.mixin.metadata.withNamespace(gateway.config.namespace) +
    service.mixin.metadata.withLabels(gateway.config.commonLabels),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;

    local c =
      container.new('observatorium-api-gateway', gateway.config.image) +
      container.withArgs([
        '--web.listen=0.0.0.0:8080',
        '--metrics.ui.endpoint=' + gateway.config.uiEndpoint,
        '--metrics.read.endpoint=' + gateway.config.readEndpoint,
        '--metrics.write.endpoint=' + gateway.config.writeEndpoint,
        '--log.level=warn',
      ]) +
      container.withPorts(
        containerPort.newNamed(8080, 'http')
      ) +
      container.mixin.livenessProbe +
      container.mixin.livenessProbe.withPeriodSeconds(30) +
      container.mixin.livenessProbe.withFailureThreshold(8) +
      container.mixin.livenessProbe.httpGet.withPort(gateway.service.spec.ports[0].port) +
      container.mixin.livenessProbe.httpGet.withScheme('HTTP') +
      container.mixin.livenessProbe.httpGet.withPath('/-/healthy') +
      container.mixin.readinessProbe +
      container.mixin.readinessProbe.withPeriodSeconds(5) +
      container.mixin.readinessProbe.withFailureThreshold(20) +
      container.mixin.readinessProbe.httpGet.withPort(gateway.service.spec.ports[0].port) +
      container.mixin.readinessProbe.httpGet.withScheme('HTTP') +
      container.mixin.readinessProbe.httpGet.withPath('/-/ready');

    deployment.new(gateway.config.name, 1, c, gateway.config.commonLabels) +
    deployment.mixin.metadata.withNamespace(gateway.config.namespace) +
    deployment.mixin.metadata.withLabels(gateway.config.commonLabels) +
    deployment.mixin.spec.selector.withMatchLabels(gateway.config.podLabelSelector) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1),

  withServiceMonitor:: {
    local trc = self,
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: trc.config.name,
        namespace: trc.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: trc.config.commonLabels,
        },
        endpoints: [
          { port: 'http' },
        ],
      },
    },
  },

  withResources:: {
    local gateway = self,

    config+:: {
      resources: error 'must provide resources',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-api-gateway' then c {
                resources: gateway.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },
}
