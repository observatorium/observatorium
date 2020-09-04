local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

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

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-up',
      'app.kubernetes.io/instance': up.config.name,
      'app.kubernetes.io/version': up.config.version,
      'app.kubernetes.io/component': 'test',
    },

    tls: {},
  },

  job:
    local job = k.batch.v1.job;
    local container = job.mixin.spec.template.spec.containersType;

    local c =
      container.new('observatorium-up', up.config.image) +
      container.withArgs(
        [
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
        ] + (
          if up.config.tls != {} then
            [
              '--tls-ca-file=/mnt/tls/' + up.config.tls.caKey,
            ]
          else []
        )
      ) +
      container.withVolumeMounts(
        (if up.config.tls != {} then [
           {
             name: 'tls',
             mountPath: '/mnt/tls',
             readOnly: true,
           },
         ] else [])
      );

    job.new() +
    job.mixin.metadata.withName(up.config.name) +
    job.mixin.spec.withBackoffLimit(up.config.backoffLimit) +
    job.mixin.spec.template.metadata.withLabels(up.config.commonLabels) +
    job.mixin.spec.template.spec.withContainers([c]) {
      spec+: {
        template+: {
          spec+: {
            restartPolicy: 'OnFailure',
          },
        },
      },
    } + job.mixin.spec.template.spec.withVolumes(
      (if up.config.tls != {} then [
         {
           configMap: {
             name: up.config.tls.configMapName,
           },
           name: 'tls',
         },
       ] else [])
    ),

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
    local job = k.batch.v1.job,
    local container = job.mixin.spec.template.spec.containersType,
    local u = self,
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
            local c =
              container.new('curl', u.config.curlImage) +
              container.withCommand([
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
                  u.config.tokenEndpoint,
                  u.config.username,
                  u.config.password,
                  u.config.clientID,
                  u.config.clientSecret,
                ],
              ]) +
              container.withVolumeMounts({
                name: 'shared',
                mountPath: '/var/shared',
                readOnly: false,
              }),

            initContainers+: [c],

            containers: [
              if c.name == 'observatorium-up' then c {
                resources: u.config.resources,
                args+: [
                  '--token-file=/var/shared/token',
                ],
              } + container.withVolumeMounts(
                c.volumeMounts + [
                  {
                    name: 'shared',
                    mountPath: '/var/shared',
                    readOnly: true,
                  },
                ]
              ) else c
              for c in super.containers
            ],
            volumes+: [
              {
                emptyDir: {},
                name: 'shared',
              },
            ],
          },
        },
      },
    },
  },

  withLogsFile:: {
    local job = k.batch.v1.job,
    local container = job.mixin.spec.template.spec.containersType,
    local u = self,
    config+:: {
      bashImage: error 'must provide image for bash',
    },

    job+: {
      spec+: {
        template+: {
          spec+: {
            local c =
              container.new('logs-file', u.config.bashImage) +
              container.withCommand([
                '/bin/sh',
                '-c',
                |||
                  cat > /var/logs-file/logs.yaml << EOF
                  spec:
                    logs: [ [ "$(date '+%s%N')", "log line"] ]
                  EOF
                |||,
              ]) +
              container.withVolumeMounts(
                [
                  {
                    name: 'logs-file',
                    mountPath: '/var/logs-file',
                    readOnly: false,
                  },
                ],
              ),

            initContainers+: [c],

            containers: [
              if c.name == 'observatorium-up' then c {
                resources: u.config.resources,
                args+: [
                  '--logs-file=/var/logs-file/logs.yaml',
                ],
              } + container.withVolumeMounts(
                c.volumeMounts +
                [
                  {
                    name: 'logs-file',
                    mountPath: '/var/logs-file',
                    readOnly: true,
                  },
                ],
              ) else c
              for c in super.containers
            ],
            volumes+: [
              {
                emptyDir: {},
                name: 'logs-file',
              },
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
