local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local list = import 'telemeter/lib/list.libsonnet';

local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local secret = k.core.v1.secret;
local deployment = k.apps.v1.deployment;

(import '../kubernetes/observatorium.libsonnet') +
{
  observatorium+:: {
    local namespace = '${NAMESPACE}',
    namespace:: namespace,

    proxyImage:: '${PROXY_IMAGE}:${PROXY_IMAGE_TAG}',
    proxyConfig+:: {
      sessionSecret: '',
    },

    api+: {
      image:: '${OBSERVATORIUM_API_IMAGE}:${OBSERVATORIUM_API_IMAGE_TAG}',

      // The proxy secret is there to encrypt session created by the oauth proxy.
      proxySecret:
        secret.new('observatorium-proxy', {
          session_secret: std.base64($.observatorium.proxyConfig.sessionSecret),
        }) +
        secret.mixin.metadata.withNamespace(namespace) +
        secret.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.observatorium.api.name }),

      service+:
        service.mixin.metadata.withNamespace(namespace) +
        service.mixin.metadata.withAnnotations({
          'service.alpha.openshift.io/serving-cert-secret-name': 'observatorium-tls',
        }) + {
          spec+: {
            ports+: [
              service.mixin.spec.portsType.newNamed('https', 8081, 'https'),
            ],
          },
        },
      local volume = deployment.mixin.spec.template.spec.volumesType,
      local container = deployment.mixin.spec.template.spec.containersType,
      local volumeMount = container.volumeMountsType,
      deployment+:
        {
          spec+: {
            template+: {
              spec+: {
                containers: [
                  if c.name == 'observatorium-api' then c {
                    resources: {
                      requests: {
                        cpu: '${OBSERVATORIUM_API_CPU_REQUEST}',
                        memory: '${OBSERVATORIUM_API_MEMORY_REQUEST}',
                      },
                      limits: {
                        cpu: '${OBSERVATORIUM_API_CPU_LIMIT}',
                        memory: '${OBSERVATORIUM_API_MEMORY_LIMIT}',
                      },
                    },
                  } else c
                  for c in super.containers
                ] + [
                  container.new('proxy', $.observatorium.proxyImage) +
                  container.withArgs([
                    '-provider=openshift',
                    '-https-address=:%d' % $.observatorium.api.service.spec.ports[1].port,
                    '-http-address=',
                    '-email-domain=*',
                    '-upstream=http://localhost:%d' % $.observatorium.api.service.spec.ports[0].port,
                    '-openshift-service-account=prometheus-telemeter',
                    '-openshift-sar={"resource": "namespaces", "verb": "get", "name": "${NAMESPACE}", "namespace": "${NAMESPACE}"}',
                    '-openshift-delegate-urls={"/": {"resource": "namespaces", "verb": "get", "name": "${NAMESPACE}", "namespace": "${NAMESPACE}"}}',
                    '-tls-cert=/etc/tls/private/tls.crt',
                    '-tls-key=/etc/tls/private/tls.key',
                    '-client-secret-file=/var/run/secrets/kubernetes.io/serviceaccount/token',
                    '-cookie-secret-file=/etc/proxy/secrets/session_secret',
                    '-openshift-ca=/etc/pki/tls/cert.pem',
                    '-openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
                    '-skip-auth-regex=^/metrics',
                  ]) +
                  container.withPorts([
                    { name: 'https', containerPort: $.observatorium.api.service.spec.ports[1].port },
                  ]) +
                  container.withVolumeMounts(
                    [
                      volumeMount.new('secret-api-tls', '/etc/tls/private'),
                      volumeMount.new('secret-api-proxy', '/etc/proxy/secrets'),
                    ]
                  ),
                ],
              },
            },
          },
        } +
        deployment.mixin.metadata.withNamespace(namespace) +
        deployment.mixin.spec.withReplicas('${OBSERVATORIUM_API_REPLICAS}') +
        deployment.mixin.spec.template.spec.withServiceAccount('prometheus-telemeter') +
        deployment.mixin.spec.template.spec.withServiceAccountName('prometheus-telemeter') +
        deployment.mixin.spec.template.spec.withVolumes([
          volume.fromSecret('secret-api-tls', 'api-tls'),
          volume.fromSecret('secret-api-proxy', 'api-proxy'),
        ]),
    },
  },
} + {
  apiVersion: 'v1',
  kind: 'Template',
  metadata: {
    name: 'observatorium',
  },
  objects: [
    $.observatorium.api[name]
    for name in std.objectFields($.observatorium.api)
  ],
  parameters: [
    { name: 'NAMESPACE', value: 'telemeter' },
    { name: 'OBSERVATORIUM_API_IMAGE', value: 'quay.io/observatorium/observatorium' },
    { name: 'OBSERVATORIUM_API_IMAGE_TAG', value: 'master-2020-01-14-d076eab' },
    { name: 'PROXY_IMAGE', value: 'openshift/oauth-proxy' },
    { name: 'PROXY_IMAGE_TAG', value: 'v1.1.0' },
    { name: 'OBSERVATORIUM_API_REPLICAS', value: '3' },
    { name: 'OBSERVATORIUM_API_CPU_REQUEST', value: '100m' },
    { name: 'OBSERVATORIUM_API_CPU_LIMIT', value: '1' },
    { name: 'OBSERVATORIUM_API_MEMORY_REQUEST', value: '256Mi' },
    { name: 'OBSERVATORIUM_API_MEMORY_LIMIT', value: '1Gi' },
  ],
}
