// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,
  name: error 'must provide name',
  image: error 'must provide image',
  version: error 'must provide version',
  namespace: error 'must provide namespace',
  tlsSecret: '%s-tls' % [defaults.name],
  tlsCertKey: 'tls.crt',  // the key in the config map for the cert
  tlsKeyKey: 'tls.key',  // the key in the config map for the cert key
  config: {
    issuer: 'https://%s.%s.svc.cluster.local:5556/dex' % [defaults.name, defaults.namespace],
    storage: {
      type: 'sqlite3',
      config: { file: '/storage/dex.db' },
    },
    web: {
      https: '0.0.0.0:5556',
      tlsCert: '/etc/dex/tls/tls.crt',
      tlsKey: '/etc/dex/tls/tls.key',
    },
    logger: { level: 'debug' },
  },
  ports: { http: 5556 },

  commonLabels:: {
    'app.kubernetes.io/name': 'dex',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
    'app.kubernetes.io/component': 'identity-provider',
  },
};

function(params) {
  local dex = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,

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
      accessModes: ['ReadWriteOnce'],
      resources: {
        requests: { storage: '1Gi' },
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
          assert std.isString(name),
          assert std.isNumber(dex.config.ports[name]),

          name: name,
          port: dex.config.ports[name],
          targetPort: dex.config.ports[name],
          protocol: 'TCP',
        }
        for name in std.objectFields(dex.config.ports)
      ],
      selector: dex.config.commonLabels,
      type: 'ClusterIP',
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
                { name: name, containerPort: dex.config.ports[name] }
                for name in std.objectFields(dex.config.ports)
              ],
              volumeMounts: [
                { name: 'config', mountPath: '/etc/dex/cfg' },
                { name: 'storage', mountPath: '/storage', readOnly: false },
                { name: 'tls', mountPath: '/etc/dex/tls' },
              ],
            },
          ],
          volumes: [
            {
              name: 'config',
              secret: {
                secretName: dex.config.name,
                items: [
                  { key: 'config.yaml', path: 'config.yaml' },
                ],
              },
            },
            {
              name: 'storage',
              persistentVolumeClaim: { claimName: dex.config.name },
            },
            {
              name: 'tls',
              secret: {
                secretName: dex.config.tlsSecret,
                items: [
                  {
                    key: dex.config.tlsCertKey,
                    path: 'tls.crt',
                  },
                  {
                    key: dex.config.tlsKeyKey,
                    path: 'tls.key',
                  },
                ],
              },
            },
          ],
        },
      },
    },
  },
}
