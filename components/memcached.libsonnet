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

    component:: 'store-cache',
    commonLabels:: {
      'app.kubernetes.io/name': 'memcached',
      'app.kubernetes.io/instance': mc.config.name,
      'app.kubernetes.io/version': mc.config.version,
      'app.kubernetes.io/component': mc.config.component,
    },

    podLabelSelector:: {
      [labelName]: mc.config.commonLabels[labelName]
      for labelName in std.objectFields(mc.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: mc.config.name,
        namespace: mc.config.namespace,
        labels: mc.config.commonLabels,
      },
      spec: {
        ports: [
          { name: 'client', targetPort: 11211, port: 11211 },
          { name: 'metrics', targetPort: 9150, port: 9150 },
        ],
        selector: mc.config.podLabelSelector,
        clusterIP: 'None',
      },
    },

  statefulSet:
    local memcached = {
      name: 'memcached',
      image: mc.config.image,
      args: [
        '-m %(memoryLimitMb)s' % mc.config,
        '-I %(maxItemSize)s' % mc.config,
        '-c %(connectionLimit)s' % mc.config,
        '-v',
      ],
      ports: [
        { name: 'client', containerPort: mc.service.spec.ports[0].port },
      ],
      resources: {
        requests: {
          cpu: mc.config.cpuRequest,
          memory: mc.util.bytesToK8sQuantity(mc.config.memoryRequestBytes),
        },
        limits: {
          cpu: mc.config.cpuLimit,
          memory: mc.util.bytesToK8sQuantity(mc.config.memoryLimitBytes),
        },
      },
      terminationMessagePolicy: 'FallbackToLogsOnError',
    };

    local exporter = {
      name: 'exporter',
      image: mc.config.exporterImage,
      args: [
        '--memcached.address=localhost:%d' % mc.service.spec.ports[0].port,
        '--web.listen-address=0.0.0.0:%d' % mc.service.spec.ports[1].port,
      ],
      ports: [
        { name: 'metrics', containerPort: mc.service.spec.ports[1].port },
      ],
    };

    {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: mc.config.name,
        namespace: mc.config.namespace,
        labels: mc.config.commonLabels,
      },
      spec: {
        replicas: mc.config.replicas,
        selector: { matchLabels: mc.config.podLabelSelector },
        serviceName: mc.service.metadata.name,
        template: {
          metadata: {
            labels: mc.config.commonLabels,
          },
          spec: {
            containers: [memcached, exporter],
            volumeClaimTemplates:: null,
          },
        },
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
