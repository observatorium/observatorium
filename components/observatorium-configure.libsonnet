{
  local obs = self,

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
