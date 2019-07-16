local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  thanos+:: {
    receiveController: {
      serviceAccount:
        local sa = k.core.v1.serviceAccount;

        sa.new() +
        sa.mixin.metadata.withName('thanos-receive-controller') +
        sa.mixin.metadata.withNamespace('observatorium'),

      role:
        local role = k.rbac.v1.role;
        local rules = role.rulesType;

        role.new() +
        role.mixin.metadata.withName('thanos-receive-controller') +
        role.mixin.metadata.withNamespace('observatorium') +
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
        rb.mixin.metadata.withName('thanos-receive-controller') +
        rb.mixin.metadata.withNamespace('observatorium') +
        rb.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
        rb.mixin.roleRef.withName($.thanos.receiveController.role.metadata.name) +
        rb.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
        rb.withSubjects([{
          kind: 'ServiceAccount',
          name: $.thanos.receiveController.serviceAccount.metadata.name,
          namespace: $.thanos.receiveController.serviceAccount.metadata.namespace,
        }]),

      configmap:
        local configmap = k.core.v1.configMap;

        configmap.new() +
        configmap.mixin.metadata.withName('observatorium-tenants') +
        configmap.mixin.metadata.withNamespace('observatorium') +
        configmap.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.receiveController.deployment.metadata.name }) +
        configmap.withData({
          'hashrings.json': std.manifestJsonEx((import '../tenants.libsonnet'), '  '),
        }),

      service:
        local service = k.core.v1.service;
        local ports = service.mixin.spec.portsType;

        service.new(
          'thanos-receive-controller',
          $.thanos.receiveController.deployment.metadata.labels,
          [
            ports.newNamed('http', 8080, 8080),
          ],
        ) +
        service.mixin.metadata.withNamespace('observatorium'),

      deployment:
        local deployment = k.apps.v1.deployment;
        local container = deployment.mixin.spec.template.spec.containersType;
        local env = container.envType;

        local c =
          container.new($.thanos.receiveController.deployment.metadata.name, 'quay.io/observatorium/thanos-receive-controller:latest') +
          container.withArgs([
            '--configmap-name=%s' % $.thanos.receiveController.configmap.metadata.name,
            '--configmap-generated-name=%s-generated' % $.thanos.receiveController.configmap.metadata.name,
            '--file-name=hashrings.json',
            '--namespace=$(NAMESPACE)',
          ]) + container.withEnv([
            env.fromFieldPath('NAMESPACE', 'metadata.namespace'),
          ]);

        deployment.new('thanos-receive-controller', 1, c, $.thanos.receiveController.deployment.metadata.labels) +
        deployment.mixin.metadata.withNamespace('observatorium') +
        deployment.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.receiveController.deployment.metadata.name }) +
        deployment.mixin.spec.template.spec.withServiceAccount($.thanos.receiveController.serviceAccount.metadata.name) +
        deployment.mixin.spec.selector.withMatchLabels($.thanos.receiveController.deployment.metadata.labels),
    },
  },
}
