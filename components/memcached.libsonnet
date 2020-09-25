local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local mc = self,

  config:: {
    name:: error 'must provide name',
    namespace:: error 'must provide namespace',
    version:: error 'must provide version',
    image:: error 'must provide image',
    exporterVersion:: error 'must provide exporter version',
    exporterImage:: error 'must provide exporter image',
    replicas:: error 'must provide replicas',

    maxItemSize:: '1m',
    memoryLimitMb:: 1024,
    connectionLimit:: 1024,

    cpuRequest:: '500m',
    cpuLimit:: '3',

    overprovisionFactor:: 1.2,
    memoryRequestBytes:: std.ceil((mc.config.memoryLimitMb * mc.config.overprovisionFactor) + 100) * 1024 * 1024,
    memoryLimitBytes:: mc.config.memoryLimitMb * 1.5 * 1024 * 1024,

    commonLabels:: {
      'app.kubernetes.io/name': 'memcached',
      'app.kubernetes.io/instance': mc.config.name,
      'app.kubernetes.io/version': mc.config.version,
      'app.kubernetes.io/component': 'store-cache',
    },

    podLabelSelector:: {
      [labelName]: mc.config.commonLabels[labelName]
      for labelName in std.objectFields(mc.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      mc.config.name,
      mc.config.podLabelSelector,
      [
        ports.newNamed('client', 11211, 11211),
        ports.newNamed('metrics', 9150, 9150),
      ]
    ) +
    service.mixin.metadata.withNamespace(mc.config.namespace) +
    service.mixin.metadata.withLabels(mc.config.commonLabels) +
    service.mixin.spec.withClusterIp('None'),

  statefulSet:
    local sts = k.apps.v1.statefulSet;
    local container = sts.mixin.spec.template.spec.containersType;

    local memcached =
      container.new('memcached', mc.config.image) +
      container.withTerminationMessagePolicy('FallbackToLogsOnError') +
      container.withPorts([
        { name: 'client', containerPort: mc.service.spec.ports[0].port },
      ]) +
      container.withArgs([
        '-m %(memoryLimitMb)s' % mc.config,
        '-I %(maxItemSize)s' % mc.config,
        '-c %(connectionLimit)s' % mc.config,
        '-v',
      ]) +
      container.mixin.resources.withRequests({
        cpu: mc.config.cpuRequest,
        memory: mc.util.bytesToK8sQuantity(mc.config.memoryRequestBytes),
      }) +
      container.mixin.resources.withLimits({
        cpu: mc.config.cpuLimit,
        memory: mc.util.bytesToK8sQuantity(mc.config.memoryLimitBytes),
      });

    local exporter =
      container.new('exporter', mc.config.exporterImage) +
      container.withPorts([
        { name: 'metrics', containerPort: mc.service.spec.ports[1].port },
      ]) +
      container.withArgs([
        '--memcached.address=localhost:%d' % mc.service.spec.ports[0].port,
        '--web.listen-address=0.0.0.0:%d' % mc.service.spec.ports[1].port,
      ]);

    sts.new(mc.config.name, mc.config.replicas, [memcached, exporter], [], mc.config.commonLabels) +
    sts.mixin.metadata.withNamespace(mc.config.namespace) +
    sts.mixin.metadata.withLabels(mc.config.commonLabels) +
    sts.mixin.spec.withServiceName(mc.service.metadata.name) +
    sts.mixin.spec.selector.withMatchLabels(mc.config.podLabelSelector)
    {
      spec+: {
        volumeClaimTemplates:: null,
      },
    },

  withServiceMonitor:: {
    local mc = self,
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: mc.config.name,
        namespace: mc.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: mc.config.podLabelSelector,
        },
        endpoints: [
          { port: 'metrics' },
        ],
      },
    },
  },

  withResources:: {
    local mc = self,
    config+:: {
      resources: error 'must provide resources',
    },

    statefulSet+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'memcached' then c {
                resources: mc.config.resources.memcached,
              } else if c.name == 'exporter' then c {
                resources: mc.config.resources.exporter,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },

  util:: {
    // Convert number to k8s "quantity" (ie 1.5Gi -> "1536Mi")
    // as per https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/api/resource/quantity.go
    // Original from https://github.com/grafana/jsonnet-libs/blob/master/memcached/memcached.libsonnet
    bytesToK8sQuantity(i)::
      local remove_factors_exponent(x, y) =
        if x % y > 0
        then 0
        else remove_factors_exponent(x / y, y) + 1;
      local remove_factors_remainder(x, y) =
        if x % y > 0
        then x
        else remove_factors_remainder(x / y, y);
      local suffixes = ['', 'Ki', 'Mi', 'Gi'];
      local suffix = suffixes[remove_factors_exponent(i, 1024)];
      '%d%s' % [remove_factors_remainder(i, 1024), suffix],
  },
}
