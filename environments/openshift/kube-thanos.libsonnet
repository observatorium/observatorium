local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;

(import '../kubernetes/kube-thanos.libsonnet') +
{
  thanos+:: {
    variables+: {
      image: '${IMAGE}',
      objectStorageConfig+: {
        name: '${THANOS_CONFIG_SECRET}',
        key: 'thanos.yaml',
      },
    },

    local namespace = '${NAMESPACE}',
    local s3Envvars = {
      spec+: {
        containers: [
          local container = sts.mixin.spec.template.spec.containersType;
          local env = container.envType;

          super.containers[0] {
            env+: [
              env.fromSecretRef('AWS_ACCESS_KEY_ID', '${THANOS_S3_SECRET}', 'aws_access_key_id'),
              env.fromSecretRef('AWS_SECRET_ACCESS_KEY', '${THANOS_S3_SECRET}', 'aws_secret_access_key'),
            ],
          },
        ],
      },
    },

    querier+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      deployment+:
        deployment.mixin.metadata.withNamespace(namespace) +
        deployment.mixin.spec.withReplicas('${{THANOS_QUERIER_REPLICAS}}'),
    },

    store+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+: {
        metadata+: {
          namespace: namespace,
        },
        spec+: {
          replicas: '${{THANOS_STORE_REPLICAS}}',

          // As we use Vault and want to be able to use rotation of credentials,
          // we need to provide the AWS key and secret via envvars, cause the thanos.yaml is written by hand.
          template+: s3Envvars,
        },
      },
    },

    receive+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+: {
        metadata+: {
          namespace: namespace,
        },
        spec+: {
          replicas: '${{THANOS_RECEIVE_REPLICAS}}',

          // As we use Vault and want to be able to use rotation of credentials,
          // we need to provide the AWS key and secret via envvars, cause the thanos.yaml is written by hand.
          template+: s3Envvars,
        },
      },
    },
  },
}
