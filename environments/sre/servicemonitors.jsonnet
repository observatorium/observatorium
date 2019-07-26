local sm =
  (import 'kube-thanos/kube-thanos-servicemonitors.libsonnet') +
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
              matchLabels: { 'app.kubernetes.io/name': 'thanos-querier' },
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
              matchLabels: { 'app.kubernetes.io/name': 'thanos-store' },
            },
          },
        },
      },
      receive+: {
        serviceMonitor+: {
          metadata: {
            name: 'observatorium-thanos-receive',
            labels: { prometheus: 'app-sre' },
          },
          spec+: {
            selector+: {
              matchLabels: { 'app.kubernetes.io/name': 'thanos-receive' },
            },
          },
        },
      },
    },
  };

{
  'observatorium-thanos-querier-stage.servicemonitor': sm.thanos.querier.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-stage'] } },
  },
  'observatorium-thanos-store-stage.servicemonitor': sm.thanos.store.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-stage'] } },
  },
  'observatorium-thanos-receive-stage.servicemonitor': sm.thanos.receive.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-stage'] } },
  },
  'observatorium-thanos-querier-production.servicemonitor': sm.thanos.querier.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-production'] } },
  },
  'observatorium-thanos-store-production.servicemonitor': sm.thanos.store.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-production'] } },
  },
  'observatorium-thanos-receive-production.servicemonitor': sm.thanos.receive.serviceMonitor {
    spec+: { namespaceSelector+: { matchNames: ['telemeter-production'] } },
  },
}
