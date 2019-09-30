local tenants = import '../../tenants.libsonnet';
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
(import 'kube-thanos/kube-thanos-compactor.libsonnet') +
(import 'kube-thanos/kube-thanos-pvc.libsonnet') +
(import 'thanos-receive-controller/thanos-receive-controller.libsonnet') +
(import '../../components/thanos-querier-cache.libsonnet') +
{
  thanos+:: {
    variables+:: {
      image: 'quay.io/thanos/thanos:v0.6.1',
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
    compactor+: {
      service+:
        service.mixin.metadata.withNamespace(namespace),
      statefulSet+: {
        metadata+: {
          namespace: namespace,
        },
        spec+: {
          template+: {
            spec+: {
              containers: [
                super.containers[0] {
                  args: [
                    'compact',
                    '--wait',
                    '--retention.resolution-raw=16d',
                    '--retention.resolution-5m=16d',
                    '--retention.resolution-1h=16d',
                    '--objstore.config=$(OBJSTORE_CONFIG)',
                    '--data-dir=/var/thanos/compactor',
                    '--debug.accept-malformed-index',
                  ],
                },
              ],
            },
          },
        },
      },
    },
    receive+: {
      service+: service.mixin.metadata.withNamespace(namespace),
      statefulSet+:: sts.mixin.metadata.withNamespace(namespace),
    } + {
      local labels = { 'app.kubernetes.io/instance': tenant.hashring },
      ['service-' + tenant.hashring]:
        super.service +
        service.mixin.metadata.withName('thanos-receive-' + tenant.hashring) +
        {
          metadata+: {
            labels+: labels,
          },
          spec+: {
            selector+: labels,
          },
        }
      for tenant in tenants
    } + {
      ['statefulSet-' + tenant.hashring]:
        super.statefulSet +
        {
          metadata+: {
            name: 'thanos-receive-' + tenant.hashring,
            labels+: {
              'controller.receive.thanos.io': 'thanos-receive-controller',
              'controller.receive.thanos.io/hashring': tenant.hashring,
            } + $.thanos.receive['service-' + tenant.hashring].metadata.labels,
          },
          spec+: {
            replicas: tenant.replicas,
            selector+: {
              matchLabels+: $.thanos.receive['service-' + tenant.hashring].metadata.labels,
            },
            serviceName: $.thanos.receive['service-' + tenant.hashring].metadata.name,
            template+: {
              metadata+: { labels+: $.thanos.receive['service-' + tenant.hashring].metadata.labels },
              spec+: {
                // This patch should probably move upstream to kube-thanos
                containers: [
                  super.containers[0] {
                    args+: [
                      '--tsdb.retention=6h',
                      '--receive.hashrings-file=/var/lib/thanos-receive/hashrings.json',
                      '--receive.local-endpoint=http://$(NAME).%s.$(NAMESPACE).svc.cluster.local:%d/api/v1/receive' % [
                        $.thanos.receive['service-' + tenant.hashring].metadata.name,
                        $.thanos.receive['service-' + tenant.hashring].spec.ports[2].port,
                      ],
                    ],
                    volumeMounts+: [
                      { name: 'observatorium-tenants', mountPath: '/var/lib/thanos-receive' },
                    ],
                    env+: [
                      local env = sts.mixin.spec.template.spec.containersType.envType;

                      env.fromFieldPath('NAMESPACE', 'metadata.namespace'),
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
        }
      for tenant in tenants
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
