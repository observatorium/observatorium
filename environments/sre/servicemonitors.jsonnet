local tenants = import '../../tenants.libsonnet';
local prom = import '../openshift/telemeter-prometheus-ams.jsonnet';

local sm =
  (import '../openshift/thanos.jsonnet') +
  (import 'kube-thanos/kube-thanos-servicemonitors.libsonnet') +
  (import '../openshift/telemeter.jsonnet') +
  {
    thanos+:: {
      querier+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-querier',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: $.thanos.querier.service.metadata.labels,
            },
          },
        },
      },
      store+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-store',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: $.thanos.store.service.metadata.labels,
            },
          },
        },
      },
      compactor+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-compactor',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: $.thanos.compactor.service.metadata.labels,
            },
          },
        },
      },
      receiveController+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-receive-controller',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: $.thanos.receiveController.service.metadata.labels,
            },
          },
        },
      },
      receive+: {
        ['serviceMonitor' + tenant.hashring]:
          super.serviceMonitor +
          {
            metadata: {
              name: 'observatorium-thanos-receive-' + tenant.hashring,
              labels: { prometheus: 'app-sre' },
            },
            spec+: {
              selector+: {
                matchLabels: $.thanos.receive['service-' + tenant.hashring].metadata.labels,
              },
            },
          }
        for tenant in tenants
      } {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-receive',
            labels: { prometheus: 'app-sre' },
          },
        },
      },
      ruler+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-ruler',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: $.thanos.ruler.service.metadata.labels,
            },
          },
        },
      },
    },
    memcached+: {
      serviceMonitor+: {
        metadata: {
          name: 'telemeter-memcached',
          labels: { prometheus: 'app-sre' },
        },
      },
    },
  } + (import '../../components/jaeger-collector.libsonnet') {
    jaeger+:: {
      serviceMonitor+: {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'observatorium-jaeger-collector',
          labels: { prometheus: 'app-sre' },
        },
        spec+: {
          selector+: {
            matchLabels: $.jaeger.adminService.metadata.labels,
          },
          endpoints: [
            { port: $.jaeger.adminService.spec.ports[0].name },
          ],
        },
      },
      agent+:: {
        local jaegerAgent = import '../../components/jaeger-agent.libsonnet',
        serviceMonitor+: {
          apiVersion: 'monitoring.coreos.com/v1',
          kind: 'ServiceMonitor',
          metadata: {
            name: 'observatorium-jaeger-agent',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: jaegerAgent.serviceLabels,
            },
            endpoints: [
              { port: $.jaeger.agentService.spec.ports[0].name },
            ],
          },
        },
      },
    },
  };


{
  [if std.objectHas(sm.thanos[name], 'serviceMonitor') then
    sm.thanos[name].serviceMonitor.metadata.name + '.servicemonitor']: sm.thanos[name].serviceMonitor {
    metadata+: { name+: '-{{environment}}' },
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  }
  for name in std.objectFields(sm.thanos)
} {
  [sm.thanos.receive['serviceMonitor' + tenant.hashring].metadata.name + '.servicemonitor']: sm.thanos.receive['serviceMonitor' + tenant.hashring] {
    metadata+: { name+: '-{{environment}}' },
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  }
  for tenant in tenants
} {
  'observatorium-prometheus-ams.servicemonitor': prom.prometheusAms.serviceMonitor {
    metadata: { name: prom.prometheusAms.serviceMonitor.metadata.name + '-{{environment}}' },
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  },
} {
  'observatorium-jaeger.servicemonitor': sm.jaeger.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  },
} {
  'observatorium-jaeger-agent.servicemonitor': sm.jaeger.agent.serviceMonitor {
    metadata+: { name+: '-{{environment}}' },
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  },
  'telemeter-memcached.servicemonitor': sm.memcached.serviceMonitor {
    metadata+: { name+: '-{{environment}}' },
    spec+: { namespaceSelector+: { matchNames: ['{{namespace}}'] } },
  },
}
