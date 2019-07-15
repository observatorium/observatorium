local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;
local list = import 'telemeter/lib/list.libsonnet';

(import '../kubernetes/kube-thanos.libsonnet') +
{
  thanos+:: {
    variables+: {
      image: '${THANOS_IMAGE}:${THANOS_IMAGE_TAG}',
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

    receiveController+:{
      configmap+:
        configmap.mixin.metadata.withNamespace(namespace),
      service+:
        service.mixin.metadata.withNamespace(namespace),
      deployment+:
        deployment.mixin.metadata.withNamespace(namespace),
    },
  },
} + {
  local thanos = super.thanos,
  thanos+:: {
    template+:
      local objects = {
        ['querier-' + name]: thanos.querier[name]
        for name in std.objectFields(thanos.querier)
      } + {
        ['store-' + name]: thanos.store[name]
        for name in std.objectFields(thanos.store)
      } + {
        ['receive-' + name]: thanos.receive[name]
        for name in std.objectFields(thanos.receive)
      } + {
        ['receive-controller-' + name]: thanos.receiveController[name]
        for name in std.objectFields(thanos.receiveController)
      };

      list.asList('thanos', objects, [
        {
          name: 'NAMESPACE',
          value: 'telemeter',
        },
        {
          name: 'THANOS_IMAGE',
          value: 'improbable/thanos',
        },
        {
          name: 'THANOS_IMAGE_TAG',
          value: 'v0.6.0-rc.0',
        },
        {
          name: 'THANOS_QUERIER_REPLICAS',
          value: '3',
        },
        {
          name: 'THANOS_STORE_REPLICAS',
          value: '5',
        },
        {
          name: 'THANOS_RECEIVE_REPLICAS',
          value: '5',
        },
        {
          name: 'THANOS_CONFIG_SECRET',
          value: 'thanos-objectstorage',
        },
        {
          name: 'THANOS_S3_SECRET',
          value: 'telemeter-thanos-stage-s3',
        },
      ]),
  },
}
