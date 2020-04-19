local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local trc = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',
    hashrings: error 'must provide hashring configuration',

    commonLabels:: {
      'app.kubernetes.io/name': 'thanos-receive-controller',
      'app.kubernetes.io/instance': trc.config.name,
      'app.kubernetes.io/version': trc.config.version,
      'app.kubernetes.io/component': 'kubernetes-controller',
    },

    podLabelSelector:: {
      [labelName]: trc.config.commonLabels[labelName]
      for labelName in std.objectFields(trc.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },
  serviceAccount:
    local sa = k.core.v1.serviceAccount;

    sa.new() +
    sa.mixin.metadata.withName(trc.config.name) +
    sa.mixin.metadata.withNamespace(trc.config.namespace) +
    sa.mixin.metadata.withLabels(trc.config.commonLabels),

  role:
    local role = k.rbac.v1.role;
    local rules = role.rulesType;

    role.new() +
    role.mixin.metadata.withName(trc.config.name) +
    role.mixin.metadata.withNamespace(trc.config.namespace) +
    role.mixin.metadata.withLabels(trc.config.commonLabels) +
    role.withRules([
      rules.new() +
      rules.withApiGroups(['']) +
      rules.withResources([
        'configmaps',
      ]) +
      rules.withVerbs(['list', 'watch', 'get', 'create', 'update']),
      rules.new() +
      rules.withApiGroups(['apps']) +
      rules.withResources([
        'statefulsets',
      ]) +
      rules.withVerbs(['list', 'watch', 'get']),
    ]),

  roleBinding:
    local rb = k.rbac.v1.roleBinding;

    rb.new() +
    rb.mixin.metadata.withName(trc.config.name) +
    rb.mixin.metadata.withNamespace(trc.config.namespace) +
    rb.mixin.metadata.withLabels(trc.config.commonLabels) +
    rb.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
    rb.mixin.roleRef.withName(trc.role.metadata.name) +
    rb.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
    rb.withSubjects([{
      kind: 'ServiceAccount',
      name: trc.serviceAccount.metadata.name,
      namespace: trc.serviceAccount.metadata.namespace,
    }]),

  configmap:
    local configmap = k.core.v1.configMap;

    configmap.new() +
    configmap.mixin.metadata.withName(trc.config.name + '-tenants') +
    configmap.mixin.metadata.withNamespace(trc.config.namespace) +
    configmap.mixin.metadata.withLabels(trc.config.commonLabels) +
    configmap.withData({
      'hashrings.json': std.manifestJsonEx(trc.config.hashrings, '  '),
    }),

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      trc.config.name,
      trc.config.podLabelSelector,
      [
        ports.newNamed('http', 8080, 8080),
      ],
    ) +
    service.mixin.metadata.withNamespace(trc.config.namespace) +
    service.mixin.metadata.withLabels(trc.config.commonLabels),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;

    local c =
      container.new('thanos-receive-controller', trc.config.image) +
      container.withArgs([
        '--configmap-name=%s' % trc.configmap.metadata.name,
        '--configmap-generated-name=%s-generated' % trc.configmap.metadata.name,
        '--file-name=hashrings.json',
        '--namespace=$(NAMESPACE)',
      ]) +
      container.withEnv([
        env.fromFieldPath('NAMESPACE', 'metadata.namespace'),
      ]) +
      container.withPorts(
        containerPort.newNamed(8080, 'http')
      );

    deployment.new(trc.config.name, 1, c, trc.config.commonLabels) +
    deployment.mixin.metadata.withNamespace(trc.config.namespace) +
    deployment.mixin.metadata.withLabels(trc.config.commonLabels) +
    deployment.mixin.spec.template.spec.withServiceAccount(trc.serviceAccount.metadata.name) +
    deployment.mixin.spec.selector.withMatchLabels(trc.config.podLabelSelector),

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
    local trc = self,
    config+:: {
      resources: error 'must provide resources',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'thanos-receive-controller' then c {
                resources: trc.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },
}
