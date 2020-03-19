local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local obs = (import '../environments/base/observatorium.jsonnet');
local up = (import '../components/up.libsonnet') +
           (import '../components/up.libsonnet').withResources + {
  config+:: {
    name: 'observatorium-up',
    version: 'master-2020-01-09-89757a5',
    image: 'quay.io/observatorium/up:master-2020-01-09-89757a5',
    commonLabels+:: {
      'app.kubernetes.io/instance': 'e2e-test',
    },
    backoffLimit: 5,
    resources: {
      limits: {
        memory: '128Mi',
        cpu: '500m',
      },
    },
    writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/write' % [
      obs.apiGateway.service.metadata.name,
      obs.apiGateway.service.metadata.namespace,
      obs.apiGateway.service.spec.ports[0].port,
    ],
    readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/query' % [
      obs.apiGateway.service.metadata.name,
      obs.apiGateway.service.metadata.namespace,
      obs.apiGateway.service.spec.ports[0].port,
    ],
  },
};

up.manifests
