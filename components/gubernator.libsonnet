// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  name: error 'must provide name',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide image',
  ports: {
    http: 8080,
    grpc: 8081,
  },
  resources: {},
  serviceMonitor: false,

  commonLabels:: {
    'app.kubernetes.io/name': 'gubernator',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'rate-limiter',
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if !std.setMember(labelName, ['app.kubernetes.io/version'])
  },
};

function(params) {
  local gubernator = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,
  // Safety checks for combined config of defaults and params
  assert std.isNumber(gubernator.config.replicas) && gubernator.config.replicas >= 0 : 'gubernator replicas has to be number >= 0',
  assert std.isObject(gubernator.config.resources),
  assert std.isBoolean(gubernator.config.serviceMonitor),

  serviceAccount: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: gubernator.config.name,
      namespace: gubernator.config.namespace,
      labels: gubernator.config.commonLabels,
    },
  },

  role: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'Role',
    metadata: {
      name: gubernator.config.name,
      namespace: gubernator.config.namespace,
      labels: gubernator.config.commonLabels,
    },

    rules: [{
      apiGroups: [''],
      resources: ['endpoints'],
      verbs: ['list', 'watch', 'get'],
    }],
  },

  roleBinding: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'RoleBinding',
    metadata: {
      name: gubernator.config.name,
      namespace: gubernator.config.namespace,
      labels: gubernator.config.commonLabels,
    },

    roleRef: {
      apiGroup: 'rbac.authorization.k8s.io',
      kind: 'Role',
      name: gubernator.role.metadata.name,
    },
    subjects: [{
      kind: 'ServiceAccount',
      name: gubernator.serviceAccount.metadata.name,
      namespace: gubernator.serviceAccount.metadata.namespace,
    }],
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: gubernator.config.name,
      namespace: gubernator.config.namespace,
      labels: gubernator.config.commonLabels,
    },
    spec: {
      ports: [
        {
          assert std.isString(name),
          assert std.isNumber(gubernator.config.ports[name]),

          name: name,
          port: gubernator.config.ports[name],
          targetPort: gubernator.config.ports[name],
        }
        for name in std.objectFields(gubernator.config.ports)
      ],
      selector: gubernator.config.podLabelSelector,
    },
  },

  deployment:
    local c = {
      name: 'gubernator',
      image: gubernator.config.image,
      env: [
        { name: 'GUBER_K8S_NAMESPACE', valueFrom: { fieldRef: { fieldPath: 'metadata.namespace' } } },
        { name: 'GUBER_K8S_POD_IP', valueFrom: { fieldRef: { fieldPath: 'status.podIP' } } },
        { name: 'GUBER_HTTP_ADDRESS', value: '0.0.0.0:%s' % gubernator.config.ports.http },
        { name: 'GUBER_GRPC_ADDRESS', value: '0.0.0.0:%s' % gubernator.config.ports.grpc },
        { name: 'GUBER_K8S_POD_PORT', value: std.toString(gubernator.config.ports.grpc) },
        { name: 'GUBER_K8S_ENDPOINTS_SELECTOR', value: 'app.kubernetes.io/name=gubernator' },
      ],
      ports: [
        { name: port.name, containerPort: port.port }
        for port in gubernator.service.spec.ports
      ],
      readinessProbe: {
        failureThreshold: 3,
        periodSeconds: 30,
        initialDelaySeconds: 10,
        timeoutSeconds: 1,
        httpGet: {
          scheme: 'HTTP',
          port: gubernator.config.ports.http,
          path: '/v1/HealthCheck',
        },
      },
      resources: gubernator.config.resources,
    };

    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: gubernator.config.name,
        namespace: gubernator.config.namespace,
        labels: gubernator.config.commonLabels,
      },
      spec: {
        replicas: gubernator.config.replicas,
        selector: { matchLabels: gubernator.config.podLabelSelector },
        strategy: {
          rollingUpdate: {
            maxSurge: 0,
            maxUnavailable: 1,
          },
        },
        template: {
          metadata: {
            labels: gubernator.config.commonLabels,
          },
          spec: {
            containers: [c],
            serviceAccountName: gubernator.serviceAccount.metadata.name,
            restartPolicy: 'Always',
          },
        },
      },
    },

  serviceMonitor: if gubernator.config.serviceMonitor == true then {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata+: {
      name: gubernator.config.name,
      namespace: gubernator.config.namespace,
    },
    spec: {
      selector: { matchLabels: gubernator.config.podLabelSelector },
      endpoints: [
        { port: 'http' },
      ],
    },
  },
}
