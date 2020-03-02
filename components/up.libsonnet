local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local up = self,

  config:: {
    name: error 'must provide name',
    version: error 'must provide version',
    image: error 'must provide image',
    backoffLimit: error 'must provide backoffLimit',
    writeEndpoint: error 'must provide writeEndpoint',
    readEndpoint: error 'must provide readEndpoint',

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-up',
      'app.kubernetes.io/instance': up.config.name,
      'app.kubernetes.io/version': up.config.version,
      'app.kubernetes.io/component': 'test',
    },
  },

  job:
    local job = k.batch.v1.job;
    local container = job.mixin.spec.template.spec.containersType;

    local c =
      container.new('observatorium-up', up.config.image) +
      container.withArgs([
        '--endpoint-write=' + up.config.writeEndpoint,
        '--endpoint-read=' + up.config.readEndpoint,
        '--period=1s',
        '--duration=2m',
        '--name=foo',
        '--labels=bar="baz"',
        '--latency=10s',
        '--initial-query-delay=5s',
        '--threshold=0.90',
      ]);

    job.new() +
    job.mixin.metadata.withName('observatorium-up') +
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

  manifests+:: {
    'observatorium-up': up.job,
  },
}
