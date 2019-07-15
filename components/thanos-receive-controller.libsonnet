local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  thanos+:: {
    receiveController: {
      configmap:
        local configmap = k.core.v1.configMap;

        configmap.new() +
        configmap.mixin.metadata.withName('observatorium-tenants') +
        configmap.mixin.metadata.withNamespace('observatorium') +
        configmap.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.receiveController.deployment.metadata.name }) +
        configmap.withData(std.manifestJsonEx((import '../tenants.libsonnet'), '  ')),

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

        local c =
          container.new($.thanos.receiveController.deployment.metadata.name, 'quay.io/observatorium/thanos-receive-controller:latest') +
          container.withArgs([
            '--configmap-name=%s' % $.thanos.receiveController.configmap.metadata.name,
            '--configmap-generated-name=%s-generated' % $.thanos.receiveController.configmap.metadata.name,
            '--file-name=hashrings.json',
          ]);

        deployment.new('thanos-receive-controller', 1, c, $.thanos.receiveController.deployment.metadata.labels) +
        deployment.mixin.metadata.withNamespace('observatorium') +
        deployment.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.receiveController.deployment.metadata.name }) +
        deployment.mixin.spec.selector.withMatchLabels($.thanos.receiveController.deployment.metadata.labels),
    },
  },
}
