{
  local dex = self,

  config:: {
    name: error 'must provide name',
    image: error 'must provide image',
    version: error 'must provide version',
    namespace: error 'must provide namespace',
    config: {
      issuer: 'http://%s.%s.svc.cluster.local:5556/dex' % [dex.config.name, dex.config.namespace],
      storage: {
        type: 'sqlite3',
        config: {
          file: '/storage/dex.db',
        },
      },
      web: {
        http: '0.0.0.0:5556',
      },
      logger: {
        level: 'debug',
      },
    },

    commonLabels:: {
      'app.kubernetes.io/name': 'dex',
      'app.kubernetes.io/instance': dex.config.name,
      'app.kubernetes.io/version': dex.config.version,
      'app.kubernetes.io/component': 'identity-provider',
    },

  },

  deployment: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      labels: dex.config.commonLabels,
      name: dex.config.name,
      namespace: dex.config.namespace,
    },
    spec: {
      selector: {
        matchLabels: dex.config.commonLabels,
      },
      template: {
        metadata: {
          labels: dex.config.commonLabels,
        },
        spec: {
          containers: [
            {
              image: dex.config.image,
              name: 'dex',
              command: ['/usr/local/bin/dex', 'serve', '/etc/dex/cfg/config.yaml'],
              ports: [
                {
                  name: 'http',
                  containerPort: 5556,
                },
              ],
              volumeMounts: [
                {
                  name: 'config',
                  mountPath: '/etc/dex/cfg',
                },
                {
                  name: 'storage',
                  mountPath: '/storage',
                  readOnly: false,
                },
              ],
            },
          ],
          volumes: [
            {
              name: 'config',
              secret: {
                secretName: dex.config.name,
                items: [
                  {
                    key: 'config.yaml',
                    path: 'config.yaml',
                  },
                ],
              },
            },
            {
              name: 'storage',
              persistentVolumeClaim: {
                claimName: dex.config.name,
              },
            },
          ],
        },
      },
    },
  },

  secret: {
    apiVersion: 'v1',
    stringData: {
      'config.yaml': std.manifestYamlDoc(dex.config.config),
    },
    kind: 'Secret',
    metadata: {
      labels: dex.config.commonLabels,
      name: dex.config.name,
      namespace: dex.config.namespace,
    },
  },

  pvc: {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      labels: dex.config.commonLabels,
      name: dex.config.name,
      namespace: dex.config.namespace,
    },
    spec: {
      accessModes: [
        'ReadWriteOnce',
      ],
      resources: {
        requests: {
          storage: '1Gi',
        },
      },
    },
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      labels: dex.config.commonLabels,
      name: dex.config.name,
      namespace: dex.config.namespace,
    },
    spec: {
      ports: [
        {
          port: 5556,
          protocol: 'TCP',
          targetPort: 5556,
        },
      ],
      selector: dex.config.commonLabels,
      type: 'ClusterIP',
    },
  },

  manifests+:: {
    'dex-deployment': dex.deployment,
    'dex-secret': dex.secret,
    'dex-pvc': dex.pvc,
    'dex-service': dex.service,
  },
}
