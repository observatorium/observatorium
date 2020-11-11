// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  namespace: error 'must set namespace for jaeger',
  image: error 'must set image for jaeger',
  replicas: error 'must provide replicas',
  pvc: {
    class: 'standard',
    size: '50Gi',
  },
  ports: {
    admin: 14269,
    query: 16686,
    grpc: 14250,
    metrics: 14271,
  },
  resources: {},
  serviceMonitor: false,

  commonLabels:: {
    'app.kubernetes.io/name': 'jaeger-collector',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'tracing',
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local j = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,
  // Safety checks for combined config of defaults and params
  assert std.isNumber(j.config.replicas) && j.config.replicas >= 0 : 'jaeger replicas has to be number >= 0',
  assert std.isObject(j.config.resources),
  assert std.isBoolean(j.config.serviceMonitor),

  headlessService: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'jaeger-collector-headless',
      namespace: j.namespace,
      labels: { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
    },
    spec: {
      ports: [
        { name: 'grpc', targetPort: j.config.grpc, port: j.config.grpc },
      ],
      selector: $.jaeger.deployment.metadata.labels,
      clusterIP: 'None',
    },
  },

  queryService: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'jaeger-query',
      namespace: j.namespace,
      labels: { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
    },
    spec: {
      ports: [
        { name: 'query', targetPort: j.config.query, port: j.config.query },
      ],
      selector: $.jaeger.deployment.metadata.labels,
    },
  },

  adminService: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'jaeger-admin',
      namespace: j.namespace,
      labels: { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
    },
    spec: {
      ports: [
        { name: 'admin', targetPort: j.config.admin, port: j.config.admin },
      ],
      selector: $.jaeger.deployment.metadata.labels,
    },
  },

  agentService: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'jaeger-agent-discovery',
      namespace: j.namespace,
      labels: { 'app.kubernetes.io/name': 'jaeger-agent' },
    },
    spec: {
      ports: [
        { name: 'metrics', targetPort: j.config.metrics, port: j.config.metrics },
      ],
      selector: { 'app.kubernetes.io/tracing': 'jaeger-agent' },
    },
  },

  volumeClaim: {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      name: 'jaeger-store-data',
      namespace: j.namespace,
      labels: { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
    },
    spec: {
      accessModes: ['ReadWriteOnce'],
      storageClassName: j.pvc.class,
      resources: {
        requests: {
          storage: j.pvc.size,
        },
      },
    },
  },

  deployment:
    local c = {
      name: $.jaeger.deployment.metadata.name,
      image: j.image,
      args: ['--collector.queue-size=4000'],
      env: [{
        name: 'SPAN_STORAGE_TYPE',
        value: 'memory',
      }],
      ports: [
        {
          assert std.isString(name),
          assert std.isNumber(j.config.ports[name]),

          name: name,
          port: j.config.ports[name],
          targetPort: j.config.ports[name],
        }
        for name in std.objectFields(j.config.ports)
      ],
      volumeMounts: [
        { name: 'jaeger-store-data', mountPath: '/var/jaeger/store', readOnly: false },
      ],
      livenessProbe: { failureThreshold: 4, periodSeconds: 30, httpGet: {
        scheme: 'HTTP',
        port: j.config.admin,
        path: '/',
      } },
      readinessProbe: { failureThreshold: 3, periodSeconds: 30, initialDelaySeconds: 10, httpGet: {
        scheme: 'HTTP',
        port: j.config.admin,
        path: '/',
      } },
      resources: {
        requests: { cpu: '1', memory: '1Gi' },
        limits: { cpu: '4', memory: '4Gi' },
      },
    };


    {
      local labels = { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'jaeger-all-in-one',
        namespace: j.namespace,
        labels: labels,
      },
      spec: {
        replicas: j.replicas,
        selector: { matchLabels: $.jaeger.deployment.metadata.labels },
        strategy: {
          rollingUpdate: {
            maxSurge: 0,
            maxUnavailable: 1,
          },
        },
        template: {
          metadata: {
            labels: labels,
          },
          spec: {
            containers: [c],
            volumes: [{
              name: $.jaeger.volumeClaim.metadata.name,
              persistentVolumeClaim: {
                claimName: $.jaeger.volumeClaim.metadata.name,
              },
            }],
          },
        },
      },
    },

  serviceMonitor: if j.config.serviceMonitor == true then {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata+: {
      name: j.config.name,
      namespace: j.config.namespace,
      labels: j.config.commonLabels,
    },
    spec: {
      selector: { matchLabels: j.config.podLabelSelector },
      endpoints: [
        { port: 'http' },
      ],
    },
  },
}
