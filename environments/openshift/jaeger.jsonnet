local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

local app =
  (import '../kubernetes/jaeger.libsonnet') + {
    jaeger+:: {
      namespace:: '${NAMESPACE}',
      image:: '${IMAGE}:${IMAGE_TAG}',
      replicas:: '${{REPLICAS}}',

      queryService+: {
        metadata+: {
          annotations+: {
            'service.alpha.openshift.io/serving-cert-secret-name': 'querier-tls',
          },
        },
        spec+: {
          ports+: [
            { name: 'https', port: 16687, targetPort: 16687 },
          ],
        },
      },

      local deployment = k.apps.v1.deployment,
      local volume = deployment.mixin.spec.template.spec.volumesType,
      local container = deployment.mixin.spec.template.spec.containersType,
      local volumeMount = container.volumeMountsType,

      deployment+: {
        spec+: {
          template+: {
            spec+: {
              containers+: [
                container.new('proxy', '${PROXY_IMAGE}:${PROXY_IMAGE_TAG}') +
                container.withArgs([
                  '-provider=openshift',
                  '-https-address=:%d' % $.jaeger.queryService.spec.ports[1].port,
                  '-http-address=',
                  '-email-domain=*',
                  '-upstream=http://localhost:%d' % $.jaeger.queryService.spec.ports[0].port,
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
                  { name: 'https', containerPort: $.jaeger.queryService.spec.ports[1].port },
                ]) +
                container.withVolumeMounts(
                  [
                    volumeMount.new('secret-jaeger-tls', '/etc/tls/private'),
                    volumeMount.new('secret-jaeger-proxy', '/etc/proxy/secrets'),
                  ]
                ),
              ],

              serviceAccount: 'prometheus-telemeter',
              serviceAccountName: 'prometheus-telemeter',
              volumes+: [
                { name: 'secret-jaeger-tls', secret: { secretName: 'jaeger-tls' } },
                { name: 'secret-jaeger-proxy', secret: { secretName: 'jaeger-proxy' } },
              ],
            },
          },
        },
      },
    },
  };

{
  apiVersion: 'v1',
  kind: 'Template',
  metadata: {
    name: 'jaeger',
  },
  objects: [
    app.jaeger[name]
    for name in std.objectFields(app.jaeger)
  ],
  parameters: [
    { name: 'NAMESPACE', value: 'telemeter' },
    { name: 'IMAGE', value: 'jaegertracing' },
    { name: 'IMAGE_TAG', value: 'all-in-one:1.14.0' },
    { name: 'REPLICAS', value: 1 },
    { name: 'PROXY_IMAGE', value: 'openshift/oauth-proxy' },
    { name: 'PROXY_IMAGE_TAG', value: 'v1.1.0' },
  ],
}
