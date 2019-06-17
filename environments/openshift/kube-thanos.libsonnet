local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;

(import '../kubernetes/kube-thanos.libsonnet') +
{
  thanos+:: {
    querier+: {
      deployment+:
        deployment.mixin.spec.withReplicas('${THANOS_QUERIER_REPLICAS}'),
    },
    store+: {
      statefulSet+:
        deployment.mixin.spec.withReplicas('${THANOS_STORE_REPLICAS}'),
    },
  },
}
