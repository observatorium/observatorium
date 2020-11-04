{
  local up = self,

  config:: {
    name: error 'must provide name',
    version: error 'must provide version',
    image: error 'must provide image',
    backoffLimit: error 'must provide backoffLimit',
    endpointType: error 'must provide endpoint type',
    writeEndpoint: error 'must provide writeEndpoint',
    readEndpoint: error 'must provide readEndpoint',

    tls: {},

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-up',
      'app.kubernetes.io/instance': up.config.name,
      'app.kubernetes.io/version': up.config.version,
      'app.kubernetes.io/component': 'test',
    },
  },

  job:
    local c = {
      name: 'observatorium-up',
      image: up.config.image,
      args: [
        '--endpoint-type=' + up.config.endpointType,
        '--endpoint-write=' + up.config.writeEndpoint,
        '--endpoint-read=' + up.config.readEndpoint,
        '--period=1s',
        '--duration=2m',
        '--name=foo',
        '--labels=bar="baz"',
        '--latency=10s',
        '--initial-query-delay=5s',
        '--threshold=0.90',
      ] + if up.config.tls != {} then [
        '--tls-ca-file=/mnt/tls/' + up.config.tls.caKey,
      ] else [],
      volumeMounts: if up.config.tls != {} then [{
        name: 'tls',
        mountPath: '/mnt/tls',
        readOnly: true,
      }] else [],
    };

    {
      apiVersion: 'batch/v1',
      kind: 'Job',
      metadata: {
        name: up.config.name,
        labels: up.config.commonLabels,
      },
      spec: {
        backoffLimit: up.config.backoffLimit,
        template: {
          metadata: {
            labels: up.config.commonLabels,
          },
          spec: {
            containers: [c],
            restartPolicy: 'OnFailure',
            volumes: if up.config.tls != {} then [{
              configMap: {
                name: up.config.tls.configMapName,
              },
              name: 'tls',
            }] else [],
          },
        },
      },
    },

  withResources:: {
    local u = self,
    config+:: {
      resources: error 'must provide resources',
    },

    job+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-up' then c {
                resources: u.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },

  withGetToken:: {
    local up = self,
    config+:: {
      curlImage: error 'must provide image for cURL',
      tokenEndpoint: error 'must provide token endpoint',
      username: error 'must provide username',
      password: error 'must provide password',
      clientID: error 'must provide clientID',
      clientSecret: error 'must provide clientSecret',
    },

    job+: {
      spec+: {
        template+: {
          spec+: {
            local curl = {
              name: 'curl',
              image: up.config.curlImage,
              command: [
                '/bin/sh',
                '-c',
                |||
                  curl --request POST \
                      --silent \
                      --url %s \
                      --header 'content-type: application/x-www-form-urlencoded' \
                      --data grant_type=password \
                      --data username=%s \
                      --data password=%s \
                      --data client_id=%s \
                      --data client_secret=%s \
                      --data scope="openid email" | sed 's/^{.*"id_token":[^"]*"\([^"]*\)".*}/\1/' > /var/shared/token
                ||| % [
                  up.config.tokenEndpoint,
                  up.config.username,
                  up.config.password,
                  up.config.clientID,
                  up.config.clientSecret,
                ],
              ],
              volumeMounts: [
                { name: 'shared', mountPath: '/var/shared', readOnly: false },
              ],
            },

            initContainers+: [curl],
            containers: [
              if c.name == 'observatorium-up' then c {
                resources: up.config.resources,
                args+: [
                  '--token-file=/var/shared/token',
                ],
                volumeMounts: c.volumeMounts + [{
                  name: 'shared',
                  mountPath: '/var/shared',
                  readOnly: true,
                }],
              } else c
              for c in super.containers
            ],
            volumes+: [
              { emptyDir: {}, name: 'shared' },
            ],
          },
        },
      },
    },
  },

  withLogsFile:: {
    local up = self,
    config+:: {
      bashImage: error 'must provide image for bash',
    },

    job+: {
      spec+: {
        template+: {
          spec+: {
            local c = {
              name: 'logs-file',
              image: up.config.bashImage,
              command: [
                '/bin/sh',
                '-c',
                |||
                  cat > /var/logs-file/logs.yaml << EOF
                  spec:
                    logs: [ [ "$(date '+%s%N')", "log line"] ]
                  EOF
                |||,
              ],
              volumeMounts: [
                { name: 'logs-file', mountPath: '/var/logs-file', readOnly: false },
              ],
            },

            initContainers+: [c],
            containers: [
              if c.name == 'observatorium-up' then c {
                resources: up.config.resources,
                args+: [
                  '--logs-file=/var/logs-file/logs.yaml',
                ],
                volumeMounts: c.volumeMounts + [{
                  name: 'logs-file',
                  mountPath: '/var/logs-file',
                  readOnly: true,
                }],
              }
              for c in super.containers
            ],
            volumes+: [
              { emptyDir: {}, name: 'logs-file' },
            ],
          },
        },
      },
    },
  },

  manifests+:: {
    [up.config.name]: up.job,
  },
}
