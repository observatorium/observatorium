local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local deployment = k.apps.v1.deployment;
local volume = deployment.mixin.spec.template.spec.volumesType;
local container = deployment.mixin.spec.template.spec.containersType;
local volumeMount = container.volumeMountsType;
local secret = k.core.v1.secret;
local service = k.core.v1.service;

{
  local op = self,
  config+:: {
    oauthProxy: {
      image: error 'must provide image',
      httpsPort: error 'must provide httpsPort',
      upstream: error 'must provide upstream',
      tlsSecretName: error 'must provide tlsSecretName',
      sessionSecretName: error 'must provide sessionSecretName',
      sessionSecret: error 'must provide proxySessionSecret',
      serviceAccountName: error 'must provide serviceAccountName',
      resources: error 'must provide resources',
    },
  },

  proxySecret:
    secret.new(op.config.oauthProxy.sessionSecretName, {
      session_secret: std.base64(op.config.oauthProxy.sessionSecret),
    }) +
    secret.mixin.metadata.withNamespace(op.config.namespace) +
    secret.mixin.metadata.withLabels(op.config.commonLabels),

  service+:
    service.mixin.metadata.withAnnotations({
      'service.alpha.openshift.io/serving-cert-secret-name': op.config.oauthProxy.tlsSecretName,
    }) + {
      spec+: {
        ports+: [
          service.mixin.spec.portsType.newNamed('https', op.config.oauthProxy.httpsPort, 'https'),
        ],
      },
    },

  deployment+: {
    spec+: {
      template+: {
        spec+: {
          containers+: [
            container.new('oauth-proxy', op.config.oauthProxy.image) +
            container.withArgs([
              '-provider=openshift',
              '-https-address=:' + op.config.oauthProxy.httpsPort,
              '-http-address=',
              '-email-domain=*',
              '-upstream=' + op.config.oauthProxy.upstream,
              '-openshift-service-account=' + op.config.oauthProxy.serviceAccountName,
              '-openshift-sar={"resource": "namespaces", "verb": "get", "name": "${NAMESPACE}", "namespace": "${NAMESPACE}"}',
              '-openshift-delegate-urls={"/": {"resource": "namespaces", "verb": "get", "name": "${NAMESPACE}", "namespace": "${NAMESPACE}"}}',
              '-tls-cert=/etc/tls/private/tls.crt',
              '-tls-key=/etc/tls/private/tls.key',
              '-client-secret-file=/var/run/secrets/kubernetes.io/serviceaccount/token',
              '-cookie-secret-file=/etc/proxy/secrets/session_secret',
              '-openshift-ca=/etc/pki/tls/cert.pem',
              '-openshift-ca=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
            ]) +
            container.withPorts([
              { name: 'https', containerPort: op.config.oauthProxy.httpsPort },
            ]) +
            container.withVolumeMounts(
              [
                volumeMount.new(op.config.oauthProxy.tlsSecretName, '/etc/tls/private'),
                volumeMount.new(op.config.oauthProxy.sessionSecretName, '/etc/proxy/secrets'),
              ]
            ) + {
              resources: op.config.oauthProxy.resources,
            },
          ],
          serviceAccountName: op.config.oauthProxy.serviceAccountName,
          volumes+: [
            volume.fromSecret(op.config.oauthProxy.tlsSecretName, op.config.oauthProxy.tlsSecretName),
            volume.fromSecret(op.config.oauthProxy.sessionSecretName, op.config.oauthProxy.sessionSecretName),
          ],
        },
      },
    },
  },
}
