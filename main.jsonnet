local app = (import 'kube-thanos.libsonnet') + {
  _config+:: {
    images+: {
      thanos: 'improbable/thanos:v0.5.0',
    },
  },
};

{ 'thanos-list': app.thanos.list } +
{ ['thanos-querier-' + name]: app.thanos.querier[name] for name in std.objectFields(app.thanos.querier) } +
{ ['thanos-store-' + name]: app.thanos.store[name] for name in std.objectFields(app.thanos.store) }
