local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;

(import '../kubernetes/kube-thanos.libsonnet') +
{
  thanos+:: {
    variables+: {
      image: '${IMAGE}',
    },

    local namespace = '${NAMESPACE}',

    querier+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      deployment+: {
        metadata+: {
          namespace: namespace,
        },
        spec+: {
          replicas: '${THANOS_QUERIER_REPLICAS}',

          // As we use Vault and want to be able to use rotation of credentials,
          // we need to provide the AWS key and secret via envvars, cause the thanos.yaml is written by hand.
          template+: {
            spec+: {
              containers: [
                local container = deployment.mixin.spec.template.spec.containersType;
                local env = container.envType;

                container.withEnv([
                  env.fromSecretRef('AWS_ACCESS_KEY_ID', 'telemeter-thanos-stage-s3', 'aws_access_key_id'),
                  env.fromSecretRef('AWS_SECRET_ACCESS_KEY', 'telemeter-thanos-stage-s3', 'aws_secret_access_key'),
                ]),
                super.containers[0],
              ],
            },
          },
        },
      },
    },
    store+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+:
        sts.mixin.metadata.withNamespace(namespace) +
        sts.mixin.spec.withReplicas('${THANOS_STORE_REPLICAS}'),
    },
    receive+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+:
        sts.mixin.metadata.withNamespace(namespace) +
        sts.mixin.spec.withReplicas('${THANOS_RECEIVE_REPLICAS}'),
    },
  },
}
