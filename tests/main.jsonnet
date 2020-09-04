local obs = (import '../environments/base/observatorium.jsonnet');

local tls = {
  name: obs.config.name + '-tls',
  manifests: {
    [self.name + '-configmap']: {
      apiVersion: 'v1',
      data: {
        'ca.pem': importstr '../tmp/certs/ca.pem',
      },
      kind: 'ConfigMap',
      metadata: {
        name: tls.name,
      },
    },
    [self.name + '-secret']: {
      apiVersion: 'v1',
      stringData: {
        'cert.pem': importstr '../tmp/certs/server.pem',
        'key.pem': importstr '../tmp/certs/server.key',
      },
      kind: 'Secret',
      metadata: {
        name: tls.name,
      },
    },
  },
};

local upJob = (import '../components/up-job.libsonnet');

local dex = (import '../components/dex.libsonnet') + {
  config+:: {
    name: 'dex',
    namespace: 'dex',
  },
};

local upMetrics = upJob + upJob.withResources + {
  config+:: {
    name: 'observatorium-up-metrics',
    version: 'master-2020-06-15-d763595',
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
    endpointType: 'metrics',
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
} + upJob.withGetToken {
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

local upMetricsTLS = upMetrics {
  config+:: {
    name: 'observatorium-up-metrics-tls',
    writeEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/receive' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    readEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/metrics/v1/test/api/v1/query' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    tls: {
      configMapName: tls.name,
      caKey: 'ca.pem',
    },
  },
};

local upLogs = upJob + upJob.withResources + {
  config+:: {
    name: 'observatorium-up-logs',
    version: 'master-2020-06-15-d763595',
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
    endpointType: 'logs',
    writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/logs/v1/test/api/v1/push' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/logs/v1/test/api/v1/query' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
  },
} + upJob.withGetToken {
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
} + upJob.withLogsFile {
  config+:: {
    // Note: Keep debian here because we need coreutils' date
    // for timestamp generation in nanoseconds.
    bashImage: 'docker.io/debian',
  },
};

local upLogsTLS = upLogs {
  config+:: {
    name: 'observatorium-up-logs-tls',
    writeEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/logs/v1/test/api/v1/push' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    readEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/logs/v1/test/api/v1/query' % [
      obs.api.service.metadata.name,
      obs.api.service.metadata.namespace,
      obs.api.service.spec.ports[1].port,
    ],
    tls: {
      configMapName: tls.name,
      caKey: 'ca.pem',
    },
  },
};


tls.manifests +
upMetrics.manifests +
upLogs.manifests +
upMetricsTLS.manifests +
upLogsTLS.manifests
