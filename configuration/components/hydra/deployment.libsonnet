{
  apiVersion: 'apps/v1',
  kind: 'Deployment',
  metadata: {
    name: 'hydra',
    labels: {
      app: 'hydra',
    },
  },
  spec: {
    selector: {
      matchLabels: {
        app: 'hydra',
      },
    },
    replicas: 1,
    strategy: {
      type: 'RollingUpdate',
    },
    template: {
      metadata: {
        labels: {
          app: 'hydra',
        },
      },
      spec: {
        volumes: [
          {
            name: 'hydra-config',
            configMap: {
              name: 'hydra-config',
            },
          },
          {
            name: 'hydra-sqlite',
            emptyDir: {},
          },
        ],
        initContainers: [
          {
            name: 'hydra-sql-migrate',
            imagePullPolicy: 'IfNotPresent',
            args: [
              'migrate',
              'sql',
              '-e',
              '--yes',
              '--config',
              '/data/hydra/config.yaml',
            ],
            volumeMounts: [
              {
                mountPath: '/data/hydra/',
                name: 'hydra-config',
                readOnly: true,
              },
              {
                mountPath: '/var/lib/sqlite',
                name: 'hydra-sqlite',
              },
            ],
          },
        ],
        containers: [
          {
            name: 'hydra',
            imagePullPolicy: 'IfNotPresent',
            args: [
              'serve',
              'all',
              '--dangerous-force-http',
              '--config',
              '/data/hydra/config.yaml',
            ],
            volumeMounts: [
              {
                mountPath: '/data/hydra/',
                name: 'hydra-config',
                readOnly: true,
              },
              {
                mountPath: '/var/lib/sqlite',
                name: 'hydra-sqlite',
              },
            ],
            ports: [
              {
                name: 'token',
                containerPort: 5555,
              },
              {
                name: 'public',
                containerPort: 4444,
              },
              {
                name: 'admin',
                containerPort: 4445,
              },
            ],
            resources: {},
          },
        ],
      },
    },
  },
}
