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

  specMixin:: {
    local sm = self,
    config+:: {
      oauthProxy: {
        image: error 'must provide image',
        httpsPort: error 'must provide httpsPort',
        upstream: error 'must provide upstream',
        tlsSecretName: error 'must provide tlsSecretName',
        sessionSecretName: error 'must provide sessionSecretName',
        serviceAccountName: error 'must provide serviceAccountName',
        resources: error 'must provide resources',
      },
    },
    spec+: {
      template+: {
        spec+: {
          containers+: [
            container.new('oauth-proxy', sm.config.oauthProxy.image) +
            container.withArgs([
              '-provider=openshift',
              '-https-address=:' + sm.config.oauthProxy.httpsPort,
              '-http-address=',
              '-email-domain=*',
              '-upstream=' + sm.config.oauthProxy.upstream,
              '-openshift-service-account=' + sm.config.oauthProxy.serviceAccountName,
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
              { name: 'https', containerPort: sm.config.oauthProxy.httpsPort },
            ]) +
            container.withVolumeMounts(
              [
                volumeMount.new(sm.config.oauthProxy.tlsSecretName, '/etc/tls/private'),
                volumeMount.new(sm.config.oauthProxy.sessionSecretName, '/etc/proxy/secrets'),
              ]
            ) + {
              resources: sm.config.oauthProxy.resources,
            },
          ],
          serviceAccountName: sm.config.oauthProxy.serviceAccountName,
          volumes+: [
            volume.fromSecret(sm.config.oauthProxy.tlsSecretName, sm.config.oauthProxy.tlsSecretName),
            volume.fromSecret(sm.config.oauthProxy.sessionSecretName, sm.config.oauthProxy.sessionSecretName),
          ],
        },
      },
    },
  },

  deploymentMixin:: {
    local dm = self,
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

    deployment+: op.specMixin {
      config+:: {
        oauthProxy+: {
          image: dm.config.oauthProxy.image,
          httpsPort: dm.config.oauthProxy.httpsPort,
          upstream: dm.config.oauthProxy.upstream,
          tlsSecretName: dm.config.oauthProxy.tlsSecretName,
          sessionSecretName: dm.config.oauthProxy.sessionSecretName,
          sessionSecret: dm.config.oauthProxy.sessionSecret,
          serviceAccountName: dm.config.oauthProxy.serviceAccountName,
          resources: dm.config.oauthProxy.resources,
        },
      },
    },
  },

  statefulSetMixin:: {
    local sm = self,
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

    statefulSet+: op.specMixin {
      config+:: {
        oauthProxy+: {
          image: sm.config.oauthProxy.image,
          httpsPort: sm.config.oauthProxy.httpsPort,
          upstream: sm.config.oauthProxy.upstream,
          tlsSecretName: sm.config.oauthProxy.tlsSecretName,
          sessionSecretName: sm.config.oauthProxy.sessionSecretName,
          serviceAccountName: sm.config.oauthProxy.serviceAccountName,
          resources: sm.config.oauthProxy.resources,
        },
      },
    },
  },
}
