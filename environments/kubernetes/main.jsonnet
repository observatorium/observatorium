local app =
  (import 'kube-thanos.libsonnet') +
  (import 'telemeter.libsonnet');

{ ['thanos-querier-' + name]: app.thanos.querier[name] for name in std.objectFields(app.thanos.querier) } +
{ ['thanos-receive-' + name]: app.thanos.receive[name] for name in std.objectFields(app.thanos.receive) } +
{ ['thanos-compactor-' + name]: app.thanos.compactor[name] for name in std.objectFields(app.thanos.compactor) } +
{ ['thanos-store-' + name]: app.thanos.store[name] for name in std.objectFields(app.thanos.store) } +
{ ['thanos-receive-controller-' + name]: app.thanos.receiveController[name] for name in std.objectFields(app.thanos.receiveController) } +
{ ['telemeter-' + name]: app.telemeterServer[name] for name in std.objectFields(app.telemeterServer) }
