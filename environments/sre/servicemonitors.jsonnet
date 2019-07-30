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
            namespaceSelector: {
              matchNames: ['telemeter-stage'],
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
            namespaceSelector: {
              matchNames: ['telemeter-stage'],
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
            namespaceSelector: {
              matchNames: ['telemeter-stage'],
            },
          },
        },
      },
    },
  };

{
  'observatorium-thanos-querier-serviceMonitor': sm.thanos.querier.serviceMonitor,
  'observatorium-thanos-store-serviceMonitor': sm.thanos.store.serviceMonitor,
  'observatorium-thanos-receive-serviceMonitor': sm.thanos.receive.serviceMonitor,
}
