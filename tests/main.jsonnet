local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local obs = (import '../jsonnet/environments/base/observatorium.jsonnet');
local up = (import '../jsonnet/components/up-job.libsonnet') +
           (import '../jsonnet/components/up-job.libsonnet').withResources + {
  config+:: {
    local cfg = self,
    name: 'observatorium-up',
    version: 'master-2020-03-31-6e67351',
    image: 'quay.io/observatorium/up:' + self.version,
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
    readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/api/v1/query' % [
      obs.apiGateway.service.metadata.name,
      obs.apiGateway.service.metadata.namespace,
      obs.apiGateway.service.spec.ports[0].port,
    ],
  },
};

up.manifests
