local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local cq = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',
    downstreamURL: error 'must provide downstreamURL',

    commonLabels:: {
      'app.kubernetes.io/name': 'cortex-query-frontend',
      'app.kubernetes.io/instance': cq.config.name,
      'app.kubernetes.io/version': cq.config.version,
      'app.kubernetes.io/component': 'query-cache',
    },

    podLabelSelector:: {
      [labelName]: cq.config.commonLabels[labelName]
      for labelName in std.objectFields(cq.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  configmap:
    local configmap = k.core.v1.configMap;

    configmap.new() +
    configmap.mixin.metadata.withName(cq.config.name) +
    configmap.mixin.metadata.withNamespace(cq.config.namespace) +
    configmap.mixin.metadata.withLabels(cq.config.commonLabels) +
    configmap.withData({
      'config.yaml': std.manifestYamlDoc(
        {
          auth_enabled: false,
          target: 'query-frontend',
          http_prefix: '',
          server: {
            http_listen_port: 9090,
          },
          frontend: {
            log_queries_longer_than: '5s',
            compress_responses: true,
          },
          query_range: {
            split_queries_by_interval: '24h',
            align_queries_with_step: true,
            cache_results: true,
            results_cache: {
              max_freshness: '1m',
              cache: {
                enable_fifocache: true,
                fifocache: {
                  size: 2048,
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
      cq.config.name,
      cq.config.podLabelSelector,
      [
        ports.newNamed('http', 9090, 9090),
      ],
    ) +
    service.mixin.metadata.withNamespace(cq.config.namespace) +
    service.mixin.metadata.withLabels(cq.config.commonLabels),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;
    local env = container.envType;
    local containerVolumeMount = container.volumeMountsType;

    local c =
      container.new('cortex-query-frontend', cq.config.image) +
      container.withArgs([
        '-log.level=debug',
        '-config.file=/etc/cache-config/config.yaml',
        '-querier.max-retries-per-request=0',
        '-server.http-read-timeout=900s',
        '-server.http-write-timeout=900s',
        '-querier.timeout=900s',
        '-frontend.downstream-url=' + cq.config.downstreamURL,
      ]) + container.withPorts([
        containerPort.newNamed(9090, 'http'),
      ]) + container.withVolumeMounts([
        containerVolumeMount.new('query-cache-config', '/etc/cache-config/'),
      ]);

    deployment.new(cq.config.name, cq.config.replicas, c, cq.config.commonLabels) +
    deployment.mixin.metadata.withNamespace(cq.config.namespace) +
    deployment.mixin.metadata.withLabels(cq.config.commonLabels) +
    deployment.mixin.spec.selector.withMatchLabels(cq.config.podLabelSelector) +
    deployment.mixin.spec.template.spec.withVolumes([
      { name: 'query-cache-config', configMap: { name: cq.configmap.metadata.name } },
    ]),

  withServiceMonitor:: {
    local cq = self,
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: cq.config.name,
        namespace: cq.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: cq.config.commonLabels,
        },
        endpoints: [
          { port: 'http' },
        ],
      },
    },
  },

  withResources:: {
    local cq = self,
    config+:: {
      resources: error 'must provide resources',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'cortex-query-frontend' then c {
                resources: cq.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },
}
