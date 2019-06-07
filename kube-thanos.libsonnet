local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;
local list = import 'telemeter/lib/list.libsonnet';

(import 'kube-thanos/kube-thanos-querier.libsonnet') +
(import 'kube-thanos/kube-thanos-store.libsonnet') +
(import 'kube-thanos/kube-thanos-pvc.libsonnet') +
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

  local t = super.thanos,
  thanos+:: {
    querier+: {
      deployment+:
        deployment.mixin.spec.withReplicas(3),
    },
    store+: {
      statefulSet+:
        sts.mixin.spec.withReplicas(5),
    },

    list:
      local querier = {
        ['thanos-querier-' + name]: t.querier[name]
        for name in std.objectFields(t.querier)
      };
      local store = {
        ['thanos-store-' + name]: t.store[name]
        for name in std.objectFields(t.store)
      };

      list.asList('thanos', querier + store, []),
  },
}
