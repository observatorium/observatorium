local t = (import 'kube-thanos/thanos.libsonnet');
local loki = import '../../components/loki.libsonnet';
local config = import 'operator-config.libsonnet';
local obs = ((import '../../components/observatorium.libsonnet') + {
               config+:: config,
             } + (import '../../components/observatorium-configure.libsonnet'));

local patchObs = obs {
  compact+::
    t.compact.withVolumeClaimTemplate {
      config+:: obs.compact.config,
    },

  rule+::
    t.rule.withVolumeClaimTemplate {
      config+:: obs.rule.config,
    },

  receivers+:: {
    [hashring.hashring]+:
      t.receive.withVolumeClaimTemplate {
        config+:: obs.receivers[hashring.hashring].config,
      }
    for hashring in obs.config.hashrings
  },

  store+:: {
    ['shard' + i]+:
      t.store.withVolumeClaimTemplate {
        config+:: obs.store['shard' + i].config,
      }
    for i in std.range(0, obs.config.store.shards - 1)
  },

  loki+:: loki.withVolumeClaimTemplate {
    config+:: obs.loki.config,
  },
};

{
  manifests: std.mapWithKey(function(k, v) v {
    metadata+: {
      ownerReferences: [{
        apiVersion: config.apiVersion,
        blockOwnerdeletion: true,
        controller: true,
        kind: config.kind,
        name: config.name,
        uid: config.uid,
      }],
    },
  }, patchObs.manifests),
}
