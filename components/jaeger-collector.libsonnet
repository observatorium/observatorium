local jaegerAgent = import './jaeger-agent.libsonnet';

{
  jaeger+:: {
    local j = self,
    namespace:: error 'must set namespace for jaeger',
    image:: error 'must set image for jaeger',
    replicas:: 1,
    pvc:: {
      class: 'standard',
      size: '50Gi',
    },

    headlessService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'jaeger-collector-headless',
        namespace: j.namespace,
        labels: $.jaeger.deployment.metadata.labels {
          'app.kubernetes.io/name': $.jaeger.deployment.metadata.name,
        },
      },
      spec: {
        ports: [
          { name: 'grpc', targetPort: 14250, port: 14250 },
        ],
        ClusterIP: 'None',
      },
    },

    queryService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'jaeger-query',
        namespace: j.namespace,
        labels: $.jaeger.deployment.metadata.labels {
          'app.kubernetes.io/name': $.jaeger.deployment.metadata.name,
        },
      },
      spec: {
        ports: [
          { name: 'query', targetPort: 16686, port: 16686 },
        ],
      },
    },

    adminService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'jaeger-admin',
        namespace: j.namespace,
        labels: $.jaeger.deployment.metadata.labels {
          'app.kubernetes.io/name': $.jaeger.deployment.metadata.name,
        },
      },
      spec: {
        ports: [
          { name: 'admin-http', targetPort: 14269, port: 14269 },
        ],
      },
    },

    agentService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'jaeger-agent-discovery',
        namespace: j.namespace,
        labels: { 'app.kubernetes.io/tracing': 'jaeger-agent' },
      },
      spec: {
        ports: [
          { name: 'metrics', targetPort: 14271, port: 14271 },
        ],
      },
    },

    volumeClaim: {
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
        name: '$.jaeger.deployment.metadata.name',
        image: j.image,
        args: [
          '--collector.queue-size=4000',
        ],
        env: [
          {
            name: 'SPAN_STORAGE_TYPE',
            value: 'memory',
          },
        ],
        ports: [
          { name: 'admin-http', targetPort: 14269, port: 14269 },
          { name: 'metrics', targetPort: 14271, port: 14271 },
          { name: 'query', targetPort: 16686, port: 16686 },
          { name: 'grpc', targetPort: 14250, port: 14250 },
        ],
        volumeMounts: [
          {
            name: 'jaeger-store-data',
            mountPath: '/var/jaeger/store',
          },
        ],
        livenessProbe: { failureThreshold: 4, periodSeconds: 30, httpGet: {
          scheme: 'HTTP',
          port: 14269,
          path: '',
        } },
        readinessProbe: { failureThreshold: 3, periodSeconds: 30, initailDelaySeconds: 10, httpGet: {
          scheme: 'HTTP',
          port: 14269,
          path: '/',
        } },
        terminationMessagePolicy: 'FallbackToLogsOnError',
        resources: {
          requests: { cpu: '1', memory: '1Gi' },
          limits: { cpu: '4', memory: '4Gi' },
        },
      };

      {
        apiVersion: 'apps/v1',
        kind: 'Deployment',
        metadata: {
          name: 'jaeger-all-in-one',
          namespace: j.namespace,
          labels: { 'app.kubernetes.io/name': $.jaeger.deployment.metadata.name },
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
              labels: gubernator.config.commonLabels,
            },
            spec: {
              containers: [c],
              serviceAccount: gubernator.serviceAccount.metadata.name,
              restartPolicy: 'Always',
              volumes: [
                {
                  name: $.jaeger.volumeClaim.metadata.name,
                  persistentVolumeClaim: {
                    claimName: $.jaeger.volumeClaim.metadata.name,
                  },
                },
              ],
            },
          },
        },
      },
  },
}
