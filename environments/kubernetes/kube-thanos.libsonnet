local tenants = import '../../tenants.libsonnet';
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local service = k.core.v1.service;
local configmap = k.core.v1.configMap;
local sts = k.apps.v1.statefulSet;
local deployment = k.apps.v1.deployment;
local container = deployment.mixin.spec.template.spec.containersType;
local containerEnv = container.envType;
local sa = k.core.v1.serviceAccount;
local role = k.rbac.v1.role;
local rolebinding = k.rbac.v1.roleBinding;

(import 'kube-thanos/kube-thanos-querier.libsonnet') +
(import 'kube-thanos/kube-thanos-store.libsonnet') +
(import 'kube-thanos/kube-thanos-store-pvc.libsonnet') +
(import 'kube-thanos/kube-thanos-receive.libsonnet') +
(import 'kube-thanos/kube-thanos-receive-pvc.libsonnet') +
(import 'kube-thanos/kube-thanos-compactor.libsonnet') +
(import 'thanos-receive-controller/thanos-receive-controller.libsonnet') +
(import '../../components/thanos-querier-cache.libsonnet') +
{
  local jaegerAgent =
    container.new('jaeger-agent', 'jaegertracing/jaeger-agent:1.14.0') +
    container.withArgs([
      '--reporter.grpc.host-port=dns:///jaeger-collector-headless.%s.svc:14250' % $.thanos.namespace,
      '--reporter.type=grpc',
      '--jaeger.tags=pod.namespace=${NAMESPACE},pod.name=${POD}',
    ]) +
    container.withEnv([
      containerEnv.fromFieldPath('NAMESPACE', 'metadata.namespace'),
      containerEnv.fromFieldPath('POD', 'metadata.name'),
    ]) +
    container.withPorts([
      container.portsType.newNamed(6831, 'jaeger-thrift'),
      container.portsType.newNamed(5778, 'configs'),
    ]),

  local jaegerAgentFlag = |||
    --tracing.config=
      type: JAEGER
      config:
        agent_host: 'localhost'
        agent_port: 6831
  |||,

  thanos+:: {
    image: 'quay.io/thanos/thanos:v0.8.1',
    objectStorageConfig+: {
      name: 'thanos-objectstorage',
      key: 'thanos.yaml',
    },

    local namespace = 'observatorium',
    namespace:: namespace,

    querier+: {
      replicas:: 3,
      deployment+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                super.containers[0]
                { args: [
                  'query',
                  '--query.replica-label=replica',
                  '--grpc-address=0.0.0.0:%d' % $.thanos.querier.service.spec.ports[0].port,
                  '--http-address=0.0.0.0:%d' % $.thanos.querier.service.spec.ports[1].port,
                  '--store=dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [
                    $.thanos.store.service.metadata.name,
                    $.thanos.store.service.metadata.namespace,
                  ],
                  jaegerAgentFlag,
                ] + [
                  '--store=dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [
                    $.thanos.receive['service-' + tenant.hashring].metadata.name,
                    $.thanos.receive['service-' + tenant.hashring].metadata.namespace,
                  ]
                  for tenant in tenants
                ] },
              ] + [jaegerAgent],
            },
          },
        },
      },
    },
    store+: {
      replicas:: 1,
      pvc+:: {
        size: '50Gi',
      },
      statefulSet+: {
        spec+: {
          template+: {
            spec+: {
              containers+: [jaegerAgent],
            },
          },
        },
      },
    },
    compactor+: {
      statefulSet+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                super.containers[0] {
                  args: [
                    'compact',
                    '--wait',
                    '--retention.resolution-raw=14d',
                    '--retention.resolution-5m=14d',
                    '--retention.resolution-1h=14d',
                    '--objstore.config=$(OBJSTORE_CONFIG)',
                    '--data-dir=/var/thanos/compactor',
                    '--debug.accept-malformed-index',
                    jaegerAgentFlag,
                  ],
                },
              ] + [jaegerAgent],
            },
          },
        },
      },
    },
    receive+: {
      pvc+:: {
        size: '50Gi',
      },
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
      // Service for each statefulset will be headless,
      // - while overarching statefulset will have cluster IP.
      service+: { spec+: { clusterIP:: '' } },
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
                  if c.name == 'thanos-receive' then c {
                    args+: [
                      '--tsdb.retention=6h',
                      '--receive.hashrings-file=/var/lib/thanos-receive/hashrings.json',
                      '--receive.local-endpoint=http://$(NAME).%s.$(NAMESPACE).svc.cluster.local:%d/api/v1/receive' % [
                        $.thanos.receive['service-' + tenant.hashring].metadata.name,
                        $.thanos.receive['service-' + tenant.hashring].spec.ports[2].port,
                      ],
                      jaegerAgentFlag,
                    ],
                    volumeMounts+: [
                      { name: 'observatorium-tenants', mountPath: '/var/lib/thanos-receive' },
                    ],
                    env+: [
                      local env = sts.mixin.spec.template.spec.containersType.envType;

                      env.fromFieldPath('NAMESPACE', 'metadata.namespace'),
                    ],
                  } else c
                  for c in super.containers
                ] + [jaegerAgent],

                local volume = sts.mixin.spec.template.spec.volumesType,
                volumes: [
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
