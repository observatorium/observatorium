{
  local obs = self,

  compact+:: {
    config+:: obs.config.compact,
  },

  thanosReceiveController+:: {
    config+:: obs.config.thanosReceiveController,
  },

  receivers+:: {
    [hashring.hashring]+: {
      config+:: obs.config.receivers,
    }
    for hashring in obs.config.hashrings
  },

  rule+:: {
    config+:: obs.config.rule,
  },

  store+:: {
    ['shard' + i]+: {
      config+:: obs.config.store,
    }
    for i in std.range(0, obs.config.store.shards - 1)
  },

  storeCache+:: {
    config+:: obs.config.storeCache,
  },

  query+:: {
    config+:: obs.config.query,
  },

  queryFrontend+:: {
    config+:: obs.config.queryFrontend,
  },

  api+:: {
    config+:: obs.config.api,
  },

  apiQuery+:: {
    config+:: obs.config.apiQuery,
  },

  lokiRingStore+:: {
    config+:: obs.config.lokiRingStore,
  },

  lokiCaches+:: {
    config+:: obs.config.lokiCaches,
  },

  loki+:: {
    config+:: obs.config.loki,
  },
}
