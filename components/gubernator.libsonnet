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

  serviceAccount:
    local sa = k.core.v1.serviceAccount;

    sa.new() +
    sa.mixin.metadata.withName(gubernator.config.name) +
    sa.mixin.metadata.withNamespace(gubernator.config.namespace) +
    sa.mixin.metadata.withLabels(gubernator.config.commonLabels),

  role:
    local role = k.rbac.v1.role;
    local rules = role.rulesType;

    role.new() +
    role.mixin.metadata.withName(gubernator.config.name) +
    role.mixin.metadata.withNamespace(gubernator.config.namespace) +
    role.mixin.metadata.withLabels(gubernator.config.commonLabels) +
    role.withRules([
      rules.new() +
      rules.withApiGroups(['']) +
      rules.withResources([
        'endpoints',
      ]) +
      rules.withVerbs(['list', 'watch', 'get']),
    ]),

  roleBinding:
    local rb = k.rbac.v1.roleBinding;

    rb.new() +
    rb.mixin.metadata.withName(gubernator.config.name) +
    rb.mixin.metadata.withNamespace(gubernator.config.namespace) +
    rb.mixin.metadata.withLabels(gubernator.config.commonLabels) +
    rb.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
    rb.mixin.roleRef.withName(gubernator.role.metadata.name) +
    rb.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
    rb.withSubjects([{
      kind: 'ServiceAccount',
      name: gubernator.serviceAccount.metadata.name,
      namespace: gubernator.serviceAccount.metadata.namespace,
    }]),

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      gubernator.config.name,
      gubernator.config.podLabelSelector,
      [
        ports.newNamed('http', 8080, 80),
        ports.newNamed('grpc', 8081, 81),
      ],
    ) +
    service.mixin.metadata.withNamespace(gubernator.config.namespace) +
    service.mixin.metadata.withLabels(gubernator.config.commonLabels),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;
    local containerVolumeMount = container.volumeMountsType;

    local c =
      container.new('gubernator', gubernator.config.image) +
      container.withEnv([
        env.fromFieldPath('GUBER_K8S_NAMESPACE', 'metadata.namespace'),
        env.fromFieldPath('GUBER_K8S_POD_IP', 'status.podIP'),
        // env.new('GUBER_HTTP_ADDRESS', '0.0.0.0:%s' % gubernator.service.spec.ports[0].targetPort),
        // env.new('GUBER_GRPC_ADDRESS', '0.0.0.0:%s' % gubernator.service.spec.ports[1].targetPort),
        env.new('GUBER_K8S_POD_PORT', std.toString(gubernator.service.spec.ports[1].port)),
        env.new('GUBER_K8S_ENDPOINTS_SELECTOR', 'app.kubernetes.io/name=gubernator'),
      ]) +
      container.withPorts([containerPort.newNamed(p.targetPort, p.name) for p in gubernator.service.spec.ports]) +
      container.mixin.readinessProbe.withFailureThreshold(3) +
      container.mixin.readinessProbe.withPeriodSeconds(30) +
      container.mixin.readinessProbe.withInitialDelaySeconds(10) +
      container.mixin.readinessProbe.withTimeoutSeconds(1) +
      container.mixin.readinessProbe.httpGet.withPath('/v1/HealthCheck').withPort(gubernator.service.spec.ports[0].targetPort).withScheme('HTTP') +
      container.mixin.resources.withLimits(gubernator.config.resources.limits) +
      container.mixin.resources.withRequests(gubernator.config.resources.requests);

    deployment.new() +
    deployment.mixin.metadata.withName(gubernator.config.name) +
    deployment.mixin.metadata.withNamespace(gubernator.config.namespace) +
    deployment.mixin.metadata.withLabels(gubernator.config.commonLabels) +
    deployment.mixin.spec.withReplicas(gubernator.config.replicas) +
    deployment.mixin.spec.selector.withMatchLabels(gubernator.config.podLabelSelector) +
    deployment.mixin.spec.template.metadata.withLabels(gubernator.config.commonLabels) +
    deployment.mixin.spec.template.spec.withServiceAccount(gubernator.serviceAccount.metadata.name) +
    deployment.mixin.spec.template.spec.withRestartPolicy('Always') +
    deployment.mixin.spec.template.spec.withContainers([c]) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1),

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
