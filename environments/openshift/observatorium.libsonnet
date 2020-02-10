local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local list = import 'telemeter/lib/list.libsonnet';

local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local secret = k.core.v1.secret;
local deployment = k.apps.v1.deployment;

(import '../base/observatorium-api.libsonnet') +
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
        secret.new('observatorium-api-proxy', {
          session_secret: std.base64($.observatorium.proxyConfig.sessionSecret),
        }) +
        secret.mixin.metadata.withNamespace(namespace) +
        secret.mixin.metadata.withLabels({ 'app.kubernetes.io/name': $.observatorium.api.name }),

      service+:
        service.mixin.metadata.withNamespace(namespace) +
        service.mixin.metadata.withAnnotations({
          'service.alpha.openshift.io/serving-cert-secret-name': 'observatorium-api-tls',
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
                  ) +
                  container.mixin.resources.withRequests({
                    cpu: '${OBSERVATORIUM_API_PROXY_CPU_REQUEST}',
                    memory: '${OBSERVATORIUM_API_PROXY_MEMORY_REQUEST}',
                  }) +
                  container.mixin.resources.withLimits({
                    cpu: '${OBSERVATORIUM_API_PROXY_CPU_LIMITS}',
                    memory: '${OBSERVATORIUM_API_PROXY_MEMORY_LIMITS}',
                  }),
                ],
              },
            },
          },
        } +
        deployment.mixin.metadata.withNamespace(namespace) +
        deployment.mixin.spec.withReplicas('${{OBSERVATORIUM_API_REPLICAS}}') +  // additional parenthesis does matter, they convert argument to an int.
        deployment.mixin.spec.template.spec.withServiceAccount('prometheus-telemeter') +
        deployment.mixin.spec.template.spec.withServiceAccountName('prometheus-telemeter') +
        deployment.mixin.spec.template.spec.withVolumes([
          volume.fromSecret('secret-api-tls', 'observatorium-api-tls'),
          volume.fromSecret('secret-api-proxy', 'observatorium-api-proxy'),
        ]),
    },
  },
}
