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

  proxySecret: {
    apiVersion: 'v1',
    kind: 'Secret',
    type: 'Opaque',
    metadata: {
      name: op.config.oauthProxy.sessionSecretName,
      namespace: op.config.namespace,
      labels: op.config.commonLabels,
    },
    data: {
      session_secret: std.base64(op.config.oauthProxy.sessionSecret),
    },
  },

  service+: {
    metadata+: {
      annotations+: {
        'service.alpha.openshift.io/serving-cert-secret-name': op.config.oauthProxy.tlsSecretName,
      },
    },
    spec+: {
      ports+: [
        { name: 'https', targetPort: 'https', port: op.config.oauthProxy.httpsPort },
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
            {
              name: 'oauth-proxy',
              image: sm.config.oauthProxy.image,
              args: [
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
              ],
              ports: [
                { name: 'https', containerPort: sm.config.oauthProxy.httpsPort },
              ],
              volumeMounts: [
                {
                  mountPath: '/etc/tls/private',
                  name: sm.config.oauthProxy.tlsSecretName,
                  readOnly: false,
                },
                {
                  mountPath: '/etc/proxy/secrets',
                  name: sm.config.oauthProxy.sessionSecretName,
                  readOnly: false,
                },
              ],
              resources: sm.config.oauthProxy.resources,
            },
          ],
          serviceAccountName: sm.config.oauthProxy.serviceAccountName,
          volumes+: [
            {
              name: sm.config.oauthProxy.tlsSecretName,
              secret: {
                secretName: sm.config.oauthProxy.tlsSecretName,
              },
            },
            {
              name: sm.config.oauthProxy.sessionSecretName,
              secret: {
                secretName: sm.config.oauthProxy.sessionSecretName,
              },
            },
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
