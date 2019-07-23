local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local secret = k.core.v1.secret;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;
local container = deployment.mixin.spec.template.spec.containersType;
local volume = k.apps.v1beta2.statefulSet.mixin.spec.template.spec.volumesType;
local volumeMount = container.volumeMountsType;
local serviceAccount = k.core.v1.serviceAccount;
local clusterRole = k.rbac.v1.clusterRole;
local policyRule = clusterRole.rulesType;
local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;
local list = import 'telemeter/lib/list.libsonnet';

(import '../kubernetes/kube-thanos.libsonnet') +
{
  thanos+:: {
    variables+: {
      image: '${THANOS_IMAGE}:${THANOS_IMAGE_TAG}',
      proxyImage: '${PROXY_IMAGE}:${PROXY_IMAGE_TAG}',
      objectStorageConfig+: {
        name: '${THANOS_CONFIG_SECRET}',
        key: 'thanos.yaml',
      },
      proxyConfig+: {
        sessionSecret: '',
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
      // The proxy secret is there to encrypt session created by the oauth proxy.
      proxySecret:
        secret.new('querier-proxy', {
          session_secret: std.base64($.thanos.variables.proxyConfig.sessionSecret),
        }) +
        secret.mixin.metadata.withNamespace(namespace) +
        secret.mixin.metadata.withLabels({ app: 'thanos-querier' }),

      service+:
        service.mixin.metadata.withNamespace(namespace) +
        service.mixin.metadata.withAnnotations({
          'service.alpha.openshift.io/serving-cert-secret-name': 'querier-tls',
        }) + {
          spec+: {
            ports+: [
              service.mixin.spec.portsType.newNamed('https', 9091, 'https'),
            ],
          },
        },

      deployment+:
        {
          spec+: {
            template+: {
              spec+: {
                containers+: [
                  container.new('proxy', $.thanos.variables.proxyImage) +
                  container.withArgs([
                    '-provider=openshift',
                    '-https-address=:%d' % $.thanos.querier.service.spec.ports[2].port,
                    '-http-address=',
                    '-email-domain=*',
                    '-upstream=http://localhost:%d' % $.thanos.querier.service.spec.ports[1].port,
                    '-openshift-service-account=telemeter-server',
                    '-openshift-sar={"resource": "namespaces", "verb": "get"}',
                    '-openshift-delegate-urls={"/": {"resource": "namespaces", "verb": "get"}}',
                    '-tls-cert=/etc/tls/private/tls.crt',
                    '-tls-key=/etc/tls/private/tls.key',
                    '-client-secret-file=/var/run/secrets/kubernetes.io/serviceaccount/token',
                    '-cookie-secret-file=/etc/proxy/secrets/session_secret',
                    '-openshift-ca=/etc/pki/tls/cert.pem',
                    '-openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
                    '-skip-auth-regex=^/metrics',
                  ]) +
                  container.withPorts([
                    { name: 'https', containerPort: $.thanos.querier.service.spec.ports[2].port },
                  ]) +
                  container.withVolumeMounts(
                    [
                      volumeMount.new('secret-querier-tls', '/etc/tls/private'),
                      volumeMount.new('secret-querier-proxy', '/etc/proxy/secrets'),
                    ]
                  ),
                ],
              },
            },
          },
        } +
        deployment.mixin.metadata.withNamespace(namespace) +
        deployment.mixin.spec.withReplicas('${{THANOS_QUERIER_REPLICAS}}') +
        deployment.mixin.spec.template.spec.withServiceAccount('telemeter-server') +
        deployment.mixin.spec.template.spec.withVolumes([
          volume.fromSecret('secret-querier-tls', 'querier-tls'),
          volume.fromSecret('secret-querier-proxy', 'querier-proxy'),
        ]),
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

    receiveController+: {
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
          name: 'PROXY_IMAGE',
          value: 'openshift/oauth-proxy',
        },
        {
          name: 'PROXY_IMAGE_TAG',
          value: 'v1.1.0',
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
