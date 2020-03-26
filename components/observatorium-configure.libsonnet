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

  query+:: {
    config+:: obs.config.query,
  },

  queryCache+:: {
    config+:: obs.config.queryCache,
  },

  apiGateway+:: {
    config+:: obs.config.apiGateway,
  },

  apiGatewayQuery+:: {
    config+:: obs.config.apiGatewayQuery,
  },

  up+:: {
    config+:: obs.config.up,
  },
}
