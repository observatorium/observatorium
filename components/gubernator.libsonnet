{
  local gubernator = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    resources: {
      requests: {},
      limits: {},
    },

    commonLabels:: {
      'app.kubernetes.io/name': 'gubernator',
      'app.kubernetes.io/instance': gubernator.config.name,
      'app.kubernetes.io/version': gubernator.config.version,
      'app.kubernetes.io/component': 'rate-limiter',
    },

    podLabelSelector:: {
      [labelName]: gubernator.config.commonLabels[labelName]
      for labelName in std.objectFields(gubernator.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

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
        { name: 'http', targetPort: 8080, port: 8080 },
        { name: 'grpc', targetPort: 8081, port: 8081 },
      ],
      selector: gubernator.config.podLabelSelector,
    },
  },

  deployment:
    local c = {
      name: 'gubernator',
      image: gubernator.config.image,
      env: [
        {
          name: 'GUBER_K8S_NAMESPACE',
          valueFrom: { fieldRef: { fieldPath: 'metadata.namespace' } },
        },
        {
          name: 'GUBER_K8S_POD_IP',
          valueFrom: { fieldRef: { fieldPath: 'status.podIP' } },
        },
        {
          name: 'GUBER_HTTP_ADDRESS',
          value: '0.0.0.0:%s' % gubernator.service.spec.ports[0].targetPort,
        },
        {
          name: 'GUBER_GRPC_ADDRESS',
          value: '0.0.0.0:%s' % gubernator.service.spec.ports[1].targetPort,
        },
        {
          name: 'GUBER_K8S_POD_PORT',
          value: std.toString(gubernator.service.spec.ports[1].port),
        },
        {
          name: 'GUBER_K8S_ENDPOINTS_SELECTOR',
          value: 'app.kubernetes.io/name=gubernator',
        },
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
          port: gubernator.service.spec.ports[0].port,
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
            serviceAccount: gubernator.serviceAccount.metadata.name,
            restartPolicy: 'Always',
          },
        },
      },
    },

  withServiceMonitor:: {
    local gubernator = self,

    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: gubernator.config.name,
        namespace: gubernator.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: gubernator.config.podLabelSelector,
        },
        endpoints: [
          { port: 'http' },
        ],
      },
    },
  },

  manifests+:: {
    'gubernator-deployment': gubernator.deployment,
    'gubernator-service': gubernator.service,
  },
}
