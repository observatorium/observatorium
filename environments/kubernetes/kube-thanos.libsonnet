local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;
local sa = k.core.v1.serviceAccount;
local role = k.rbac.v1.role;
local rolebinding = k.rbac.v1.roleBinding;

(import 'kube-thanos/kube-thanos-querier.libsonnet') +
(import 'kube-thanos/kube-thanos-store.libsonnet') +
(import 'kube-thanos/kube-thanos-receive.libsonnet') +
// (import 'kube-thanos/kube-thanos-pvc.libsonnet') +
(import '../../components/thanos-receive-controller.libsonnet') +
{
  thanos+:: {
    variables+:: {
      image: 'improbable/thanos:v0.6.0-rc.0',
      objectStorageConfig+: {
        name: 'thanos-objectstorage',
        key: 'thanos.yaml',
      },
    },

    local namespace = 'observatorium',

    querier+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      deployment+:
        deployment.mixin.metadata.withNamespace(namespace) +
        deployment.mixin.spec.withReplicas(3),
    },
    store+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+:
        sts.mixin.metadata.withNamespace(namespace) +
        sts.mixin.spec.withReplicas(3),
    },
    receive+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+: {
        metadata+: {
          name: 'thanos-receive-default',
          namespace: namespace,
          labels+: {
            'controller.receive.thanos.io': 'thanos-receive-controller',
          },
        },
        spec+: {
          replicas: 3,
          template+: {
            spec+: {
              // This patch should probably move upstream to kube-thanos
              containers: [
                super.containers[0] {
                  args+: [
                    '--tsdb.retention=6h',
                    '--receive.hashrings-file=/var/lib/thanos-receive/hashrings.json',
                  ],
                  volumeMounts+: [
                    { name: 'observatorium-tenants', mountPath: '/var/lib/thanos-receive' },
                  ],
                },
              ],

              local volume = sts.mixin.spec.template.spec.volumesType,
              volumes+: [
                volume.withName('observatorium-tenants') +
                volume.mixin.configMap.withName('%s-generated' % $.thanos.receiveController.configmap.metadata.name),
              ],
            },
          },
        },
      },
    },
    receiveController+: {
      serviceAccount+:
        sa.mixin.metadata.withNamespace(namespace),
      role+:
        role.mixin.metadata.withNamespace(namespace),
      roleBinding+:
        rolebinding.mixin.metadata.withNamespace(namespace),
      configmap+:
        configmap.mixin.metadata.withNamespace(namespace),
      service+:
        service.mixin.metadata.withNamespace(namespace),
      deployment+:
        deployment.mixin.metadata.withNamespace(namespace),
    },
  },
}
