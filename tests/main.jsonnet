local obs = (import '../environments/base/observatorium.jsonnet');

local dex = (import '../components/dex.libsonnet') + {
  config+:: {
    name: 'dex',
    namespace: 'dex',
  },
};

local up = (import '../components/up-job.libsonnet') + (import '../components/up-job.libsonnet').withResources + {
  config+:: {
    name: 'observatorium-up',
    version: 'master-2020-05-15-716e0b4',
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
    writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/receive' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/query' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
  },
} + (import '../components/up-job.libsonnet').withGetToken + {
  config+:: {
    curlImage: 'docker.io/curlimages/curl',
    tokenEndpoint: 'http://%s.%s.svc.cluster.local:%d/dex/token' % [
      dex.service.metadata.name,
      dex.service.metadata.namespace,
      dex.service.spec.ports[0].port,
    ],
    username: 'admin@example.com',
    password: 'password',
    clientID: 'test',
    clientSecret: 'ZXhhbXBsZS1hcHAtc2VjcmV0',
  },
};

up.manifests
