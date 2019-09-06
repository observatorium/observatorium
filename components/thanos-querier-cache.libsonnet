local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  thanos+:: {
    querierCache: {
      configmap:
        local configmap = k.core.v1.configMap;

        configmap.new() +
        configmap.mixin.metadata.withName('observatorium-cache-conf') +
        configmap.mixin.metadata.withNamespace('observatorium') +
        configmap.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.querierCache.deployment.metadata.name }) +
        configmap.withData({
          'observatorium-cache-config.yaml': std.manifestYamlDoc(
            {
              auth_enabled: false,
              target: 'query-frontend',
              http_prefix: null,
              server: {
                http_listen_port: 9091,
              },
              frontend: {
                split_queries_by_day: true,
                align_queries_with_step: true,
                cache_results: true,
                compress_responses: true,
                results_cache: {
                  max_freshness: '1m',
                  cache: {
                    enable_fifocache: true,
                    fifocache: {
                      size: 1024,
                      validity: '6h',
                    },
                  },
                },
              },
            }
          ),
        }),

      service:
        local service = k.core.v1.service;
        local ports = service.mixin.spec.portsType;

        service.new(
          'observatorium-cache',
          $.thanos.querierCache.deployment.metadata.labels,
          [
            ports.newNamed('http', 9091, 9091),
          ],
        ) +
        service.mixin.metadata.withNamespace('observatorium') +
        service.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.querierCache.deployment.metadata.name }),

      deployment:
        local deployment = k.apps.v1.deployment;
        local container = deployment.mixin.spec.template.spec.containersType;
        local containerPort = container.portsType;
        local env = container.envType;

        local c =
          container.new($.thanos.querierCache.deployment.metadata.name, 'quay.io/cortexproject/cortex:master-8533a216') +
          container.withArgs([
            '-config.file=/etc/cache-config/%s.yaml' % $.thanos.querierCache.configmap.metadata.name,
            '-frontend.downstream-url=%s.%s.svc.cluster.local' % [
              $.thanos.querierCache.service.metadata.name,
              $.thanos.querierCache.service.metadata.namespace,
            ],
          ]) + container.withEnv([
            env.fromFieldPath('NAMESPACE', 'metadata.namespace'),
          ]) + container.withPorts(
            [
              containerPort.newNamed(9001, 'http'),
              containerPort.newNamed(9005, 'grpc'),
            ],
          );

        deployment.new('observatorium-querier-cache', 1, c, $.thanos.querierCache.deployment.metadata.labels) +
        deployment.mixin.metadata.withNamespace('observatorium') +
        deployment.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.thanos.querierCache.deployment.metadata.name }) +
        deployment.mixin.spec.selector.withMatchLabels($.thanos.querierCache.deployment.metadata.labels),
    },
  },
}
