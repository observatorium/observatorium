local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;

(import 'kube-thanos/kube-thanos-querier.libsonnet') +
(import 'kube-thanos/kube-thanos-store.libsonnet') +
(import 'kube-thanos/kube-thanos-receive.libsonnet') +
// (import 'kube-thanos/kube-thanos-pvc.libsonnet') +
{
  _config+:: {
    namespace: 'monitoring',

    thanos+: {
      objectStorageConfig+: {
        name: 'thanos-objectstorage',
        key: 'thanos.yaml',
      },
    },
  },

  thanos+:: {
    querier+: {
      deployment+:
        deployment.mixin.spec.withReplicas(3),
    },
    store+: {
      statefulSet+:
        sts.mixin.spec.withReplicas(5),
    },
  },
}
