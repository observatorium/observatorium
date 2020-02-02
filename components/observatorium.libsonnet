local t = (import 'kube-thanos/thanos.libsonnet');
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local obs = self,

  compact::
    t.compact +
    t.compact.withVolumeClaimTemplate +
    t.compact.withRetention +
    t.compact.withDownsamplingDisabled + {
      config+:: {
        name: 'thanos-compact',
        namespace: obs.config.namespace,
        replicas: 1,
      },
    },

  thanosReceiveController:: (import 'thanos-receive-controller/thanos-receive-controller.libsonnet') + {
    config+:: {
      name: 'thanos-receive-controller',
      namespace: obs.config.namespace,
      replicas: 1,
    },
  },

  receivers:: {
    [hashring.hashring]:
      t.receive +
      t.receive.withVolumeClaimTemplate +
      t.receive.withRetention +
      t.receive.withHashringConfigMap + {
        config+:: {
          name: 'thanos-receive-' + hashring.hashring,
          namespace: obs.config.namespace,
          image: error 'must provide image',
          version: error 'must provide version',
          objectStorageConfig: error 'must provide objectStorageConfig',
          volumeClaimTemplate: error 'must provide volumeClaimTemplate',
          replicas: 3,
          replicationFactor: 3,
          retention: '6h',
          hashringConfigMapName: '%s-generated' % obs.thanosReceiveController.configmap.metadata.name,
        },
        statefulSet+: {
          metadata+: {
            labels+: {
              'controller.receive.thanos.io/hashring': hashring.hashring,
              'controller.receive.thanos.io': 'thanos-receive-controller',
            },
          },
        },
      }
    for hashring in obs.config.hashrings
  },

  rule:: t.rule + t.rule.withVolumeClaimTemplate + {
    config+:: {
      name: 'thanos-rule',
      namespace: obs.config.namespace,
      replicas: 2,
    },
  },

  store:: t.store + t.store.withVolumeClaimTemplate + {
    config+:: {
      name: 'thanos-store',
      namespace: obs.config.namespace,
      replicas: 1,
    },
  },

  query:: t.query {
    config+:: {
      name: 'thanos-query',
      namespace: obs.config.namespace,
      replicas: 1,
      stores: [
        'dnssrv+_grpc._tcp.%s.%s.svc.cluster.local' % [service.metadata.name, service.metadata.namespace]
        for service in [obs.rule.service, obs.store.service] + [obs.receivers[hashring].service for hashring in std.objectFields(obs.receivers)]
      ],
      replicaLabels: ['prometheus_replica', 'ruler_replica', 'replica'],
    },
  },

  queryCache:: (import 'cortex-query-frontend.libsonnet') + {
    config+:: {
      name: 'cortex-query-cache',
      namespace: obs.config.namespace,
      downstreamURL: 'http://%s.%s.svc.cluster.local.:%d' % [
        obs.query.service.metadata.name,
        obs.query.service.metadata.namespace,
        obs.query.service.spec.ports[1].port,
      ],
    },
  },
} + {
  local obs = self,

  rule+:: {
    config+:: {
      queriers: ['dnssrv+_http._tcp.%s.%s.svc.cluster.local' % [obs.query.service.metadata.name, obs.query.service.metadata.namespace]],
    },
  },

  manifests+:: {
    ['thanos-query-' + name]: obs.query[name]
    for name in std.objectFields(obs.query)
  } + {
    ['query-cache-' + name]: obs.queryCache[name]
    for name in std.objectFields(obs.queryCache)
  } + {
    ['thanos-receive-' + hashring + '-' + name]: obs.receivers[hashring][name]
    for hashring in std.objectFields(obs.receivers)
    for name in std.objectFields(obs.receivers[hashring])
  } + {
    thanosReceiveService:
      local service = k.core.v1.service;
      local ports = service.mixin.spec.portsType;

      service.new(
        'thanos-receive',
        { 'app.kubernetes.io/component': 'thanos-receive' },
        [
          ports.newNamed('grpc', 10901, 10901),
          ports.newNamed('http', 10902, 10902),
          ports.newNamed('remote-write', 19291, 19291),
        ]
      ) +
      service.mixin.metadata.withNamespace(obs.config.namespace),
  } + {
    ['thanos-compact-' + name]: obs.compact[name]
    for name in std.objectFields(obs.compact)
  } + {
    ['thanos-store-' + name]: obs.store[name]
    for name in std.objectFields(obs.store)
  } + {
    ['thanos-rule-' + name]: obs.rule[name]
    for name in std.objectFields(obs.rule)
  } + {
    ['thanos-receive-controller-' + name]: obs.thanosReceiveController[name]
    for name in std.objectFields(obs.thanosReceiveController)
  },
}
