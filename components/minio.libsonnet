{
  local minio = self,

  config:: {
    namespace: error 'must provide namespace',
  },

  deployment: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name: 'minio',
      namespace: minio.config.namespace,
    },
    spec: {
      selector: {
        matchLabels: {
          'app.kubernetes.io/name': 'minio',
        },
      },
      strategy: {
        type: 'Recreate',
      },
      template: {
        metadata: {
          labels: {
            'app.kubernetes.io/name': 'minio',
          },
        },
        spec: {
          containers: [
            {
              command: [
                '/bin/sh',
                '-c',
                |||
                  mkdir -p /storage/thanos && \
                  mkdir -p /storage/loki && \
                  /usr/bin/minio server /storage
                |||,
              ],
              env: [
                {
                  name: 'MINIO_ACCESS_KEY',
                  value: 'minio',
                },
                {
                  name: 'MINIO_SECRET_KEY',
                  value: 'minio123',
                },
              ],
              image: 'minio/minio',
              name: 'minio',
              ports: [
                {
                  containerPort: 9000,
                },
              ],
              volumeMounts: [
                {
                  mountPath: '/storage',
                  name: 'storage',
                },
              ],
            },
          ],
          volumes: [
            {
              name: 'storage',
              persistentVolumeClaim: {
                claimName: 'minio',
              },
            },
          ],
        },
      },
    },
  },

  pvc: {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      labels: {
        'app.kubernetes.io/name': 'minio',
      },
      name: 'minio',
      namespace: minio.config.namespace,
    },
    spec: {
      accessModes: [
        'ReadWriteOnce',
      ],
      resources: {
        requests: {
          storage: '10Gi',
        },
      },
    },
  },

  secretThanos: {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'thanos-objectstorage',
      namespace: minio.config.bucketSecretNamespace,
    },
    stringData: {
      'thanos.yaml': |||
        type: s3
        config:
          bucket: thanos
          endpoint: %s.%s.svc.cluster.local:9000
          insecure: true
          access_key: minio
          secret_key: minio123
      ||| % [minio.service.metadata.name, minio.config.namespace],
    },
    type: 'Opaque',
  },

  secretLoki: {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'loki-objectstorage',
      namespace: minio.config.bucketSecretNamespace,
    },
    stringData: {
      endpoint: 'http://minio:minio123@%s.%s.svc.cluster.local.:9000/loki' % [minio.service.metadata.name, minio.config.namespace],
    },
    type: 'Opaque',
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: 'minio',
      namespace: minio.config.namespace,
    },
    spec: {
      ports: [
        {
          port: 9000,
          protocol: 'TCP',
          targetPort: 9000,
        },
      ],
      selector: {
        'app.kubernetes.io/name': 'minio',
      },
      type: 'ClusterIP',
    },
  },

  manifests+:: {
    'minio-deployment': minio.deployment,
    'minio-pvc': minio.pvc,
    'minio-secret-thanos': minio.secretThanos,
    'minio-secret-loki': minio.secretLoki,
    'minio-service': minio.service,
  },
}
