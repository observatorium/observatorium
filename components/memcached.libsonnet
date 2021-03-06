// Convert number to k8s "quantity" (ie 1.5Gi -> "1536Mi")
// as per https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/api/resource/quantity.go
// Original from https://github.com/grafana/jsonnet-libs/blob/master/memcached/memcached.libsonnet
local bytesToKubernetesQuantity(i) =
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
  '%d%s' % [remove_factors_remainder(i, 1024), suffix];

// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  name: error 'must provide name',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide image',
  exporterVersion: error 'must provide exporter version',
  exporterImage: error 'must provide exporter image',
  replicas: error 'must provide replicas',
  resources: {},
  serviceMonitor: false,

  maxItemSize: '1m',
  memoryLimitMb: 1024,
  connectionLimit: 1024,

  cpuRequest:: '500m',
  cpuLimit:: '3',

  overprovisionFactor:: 1.2,
  memoryRequestBytes:: std.ceil((defaults.memoryLimitMb * defaults.overprovisionFactor) + 100) * 1024 * 1024,
  memoryLimitBytes:: defaults.memoryLimitMb * 1.5 * 1024 * 1024,

  component:: error 'must provide component',
  commonLabels:: {
    'app.kubernetes.io/name': 'memcached',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': defaults.component,
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local mc = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,
  // Safety checks for combined config of defaults and params
  assert std.isNumber(mc.config.replicas) && mc.config.replicas >= 0 : 'memcached replicas has to be number >= 0',
  assert std.isObject(mc.config.resources) : 'memcached resources has to be an object',
  assert std.isBoolean(mc.config.serviceMonitor),

  service: {
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

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: mc.config.name,
      namespace: mc.config.namespace,
      labels: mc.config.commonLabels,
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
      securityContext: {
        runAsUser: 65534,
      },
      ports: [
        { name: 'client', containerPort: mc.service.spec.ports[0].port },
      ],
      resources: if std.objectHas(mc.config.resources, 'memcached') then mc.config.resources.memcached else {
        requests: {
          cpu: mc.config.cpuRequest,
          memory: bytesToKubernetesQuantity(mc.config.memoryRequestBytes),
        },
        limits: {
          cpu: mc.config.cpuLimit,
          memory: bytesToKubernetesQuantity(mc.config.memoryLimitBytes),
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
      securityContext: {
        runAsUser: 65534,
      },
      ports: [
        { name: 'metrics', containerPort: mc.service.spec.ports[1].port },
      ],
      resources: if std.objectHas(mc.config.resources, 'exporter') then mc.config.resources.exporter else {},
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
            serviceAccountName: mc.serviceAccount.metadata.name,
            securityContext: {
              fsGroup: 65534,
            },
            containers: [memcached, exporter],
            volumeClaimTemplates:: null,
          },
        },
      },
    },

  serviceMonitor: if mc.config.serviceMonitor == true then {
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
}
