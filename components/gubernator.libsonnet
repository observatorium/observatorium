local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local gubernator = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    resources: {
      requests: {},
      limits: {},
    },

    commonLabels:: {
      'app.kubernetes.io/name': 'gubernator',
      'app.kubernetes.io/instance': gubernator.config.name,
      'app.kubernetes.io/version': gubernator.config.version,
      'app.kubernetes.io/component': 'rate-limiter',
    },

    podLabelSelector:: {
      [labelName]: gubernator.config.commonLabels[labelName]
      for labelName in std.objectFields(gubernator.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      gubernator.config.name,
      gubernator.config.podLabelSelector,
      [
        ports.newNamed('http', 80, 80),
        ports.newNamed('grpc', 81, 81),
      ],
    ) +
    service.mixin.metadata.withNamespace(gubernator.config.namespace) +
    service.mixin.metadata.withLabels(gubernator.config.commonLabels),

  deployment:
    local deployments = k.apps.v1.deployment;
    local container = deployments.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;
    local containerVolumeMount = container.volumeMountsType;

    local c =
      container.new('gubernator', gubernator.config.image) +
      container.withEnv([
        env.fromFieldPath('GUBER_K8S_NAMESPACE', 'metadata.namespace'),
        env.fromFieldPath('GUBER_K8S_POD_IP', 'status.podIP'),
        env.new('GUBER_K8S_POD_PORT', gubernator.service.spec.ports[1].port),
        env.new('GUBER_K8S_ENDPOINTS_SELECTOR', 'app.kubernetes.io/name=gubernator'),
      ]) +
      container.withPorts([
        containerPort.newNamed(80, 'http'),
        containerPort.newNamed(81, 'grpc'),
      ]) +
      container.mixin.resources.withLimits(gubernator.config.resources.limits) +
      container.mixin.resources.withRequests(gubernator.config.resources.requests);

    deployments.new('gubernator', gubernator.config.replicas, c, gubernator.config.commonLabels) +
    deployments.mixin.metadata.withName(gubernator.config.name) +
    deployments.mixin.metadata.withNamespace(gubernator.config.namespace) +
    deployments.mixin.spec.selector.withMatchLabels(gubernator.config.podLabelSelector) +
    deployments.mixin.spec.template.metadata.withLabels(gubernator.config.commonLabels) +
    deployments.mixin.spec.template.spec.withRestartPolicy('always') +
    deployments.mixin.spec.template.spec.withContainers([c]),

  withServiceMonitor:: {
    local gubernator = self,

    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: gubernator.config.name,
        namespace: gubernator.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: gubernator.config.podLabelSelector,
        },
        endpoints: [
          { port: 'http' },
        ],
      },
    },
  },

  manifests+:: {
    'gubernator-deployment': gubernator.deployment,
    'gubernator-service': gubernator.service,
  },
}
