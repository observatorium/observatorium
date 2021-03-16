local obs = (import '../components/observatorium.libsonnet');

local dex = (import '../components/dex.libsonnet')({
  name: 'dex',
  namespace: 'dex',
});

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
    'test-ca-tls': {  // similar to OpenShift's service-ca injection
      apiVersion: 'v1',
      data: {
        'service-ca.crt': importstr '../tmp/certs/ca.pem',
      },
      kind: 'ConfigMap',
      metadata: {
        name: 'test-ca-tls',
        namespace: obs.api.service.metadata.namespace,
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
    [self.name + '-dex']: {
      apiVersion: 'v1',
      stringData: {
        'tls.crt': importstr '../tmp/certs/server.pem',
        'tls.key': importstr '../tmp/certs/server.key',
      },
      kind: 'Secret',
      metadata: {
        name: dex.config.tlsSecret,
        namespace: dex.config.namespace,
      },
    },
  },
};

local up = (import 'up/job/up.libsonnet');

local metricsConfig = {
  name: 'observatorium-up-metrics',
  version: 'master-2020-11-04-0c6ece8',
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
    requests: {
      memory: '128Mi',
      cpu: '50m',
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
  getToken: {
    image: 'docker.io/curlimages/curl',
    endpoint: 'https://%s.%s.svc.cluster.local:%d/dex/token' % [
      dex.service.metadata.name,
      dex.service.metadata.namespace,
      dex.service.spec.ports[0].port,
    ],
    username: 'admin@example.com',
    password: 'password',
    clientID: 'test',
    clientSecret: 'ZXhhbXBsZS1hcHAtc2VjcmV0',
    oidc: {
      configMapName: tls.name,
      caKey: 'ca.pem',
    },
  },
};

local upMetrics = up(metricsConfig);

local upMetricsTLS = up(metricsConfig {
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
});

local logsConfig = {
  name: 'observatorium-up-logs',
  version: 'master-2020-11-04-0c6ece8',
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
    requests: {
      memory: '128Mi',
      cpu: '50m',
    },
  },
  endpointType: 'logs',
  writeEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/logs/v1/test/loki/api/v1/push' % [
    obs.api.service.metadata.name,
    obs.api.service.metadata.namespace,
    obs.api.service.spec.ports[1].port,
  ],
  readEndpoint: 'http://%s.%s.svc.cluster.local:%d/api/logs/v1/test/loki/api/v1/query' % [
    obs.api.service.metadata.name,
    obs.api.service.metadata.namespace,
    obs.api.service.spec.ports[1].port,
  ],
  getToken: {
    image: 'docker.io/curlimages/curl',
    endpoint: 'https://%s.%s.svc.cluster.local:%d/dex/token' % [
      dex.service.metadata.name,
      dex.service.metadata.namespace,
      dex.service.spec.ports[0].port,
    ],
    username: 'admin@example.com',
    password: 'password',
    clientID: 'test',
    clientSecret: 'ZXhhbXBsZS1hcHAtc2VjcmV0',
    oidc: {
      configMapName: tls.name,
      caKey: 'ca.pem',
    },
  },
  sendLogs: {
    // Note: Keep debian here because we need coreutils' date
    // for timestamp generation in nanoseconds.
    image: 'docker.io/debian',
  },
};

local upLogs = up(logsConfig);

local upLogsTLS = up(logsConfig {
  name: 'observatorium-up-logs-tls',
  writeEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/logs/v1/test/loki/api/v1/push' % [
    obs.api.service.metadata.name,
    obs.api.service.metadata.namespace,
    obs.api.service.spec.ports[1].port,
  ],
  readEndpoint: 'https://%s.%s.svc.cluster.local:%d/api/logs/v1/test/loki/api/v1/query' % [
    obs.api.service.metadata.name,
    obs.api.service.metadata.namespace,
    obs.api.service.spec.ports[1].port,
  ],
  tls: {
    configMapName: tls.name,
    caKey: 'ca.pem',
  },
});

tls.manifests
{ 'observatorium-up-metrics': upMetrics.job } +
{ 'observatorium-up-metrics-tls': upMetricsTLS.job } +
{ 'observatorium-up-logs': upLogs.job } +
{ 'observatorium-up-logs-tls': upLogsTLS.job }
