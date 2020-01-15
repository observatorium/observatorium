local app =
  (import 'kube-thanos.libsonnet') +
  (import 'telemeter.libsonnet') +
  (import 'observatorium.libsonnet') +
  (import 'jaeger.libsonnet');

{ ['observatorium-api-' + name]: app.observatorium.api[name] for name in std.objectFields(app.observatorium.api) } +
{ ['thanos-querier-' + name]: app.thanos.querier[name] for name in std.objectFields(app.thanos.querier) } +
{ ['thanos-receive-' + name]: app.thanos.receive[name] for name in std.objectFields(app.thanos.receive) } +
{ ['thanos-compactor-' + name]: app.thanos.compactor[name] for name in std.objectFields(app.thanos.compactor) } +
{ ['thanos-store-' + name]: app.thanos.store[name] for name in std.objectFields(app.thanos.store) } +
{ ['thanos-ruler-' + name]: app.thanos.ruler[name] for name in std.objectFields(app.thanos.ruler) } +
{ ['thanos-receive-controller-' + name]: app.thanos.receiveController[name] for name in std.objectFields(app.thanos.receiveController) } +
{ ['thanos-querier-cache-' + name]: app.thanos.querierCache[name] for name in std.objectFields(app.thanos.querierCache) } +
{ ['thanos-rules-' + name]: app.thanos.rules[name] for name in std.objectFields(app.thanos.rules) } +
{ ['telemeter-' + name]: app.telemeterServer[name] for name in std.objectFields(app.telemeterServer) } +
{ ['telemeter-memcached-' + name]: app.memcached[name] for name in std.objectFields(app.memcached) } +
{ ['jaeger-' + name]: app.jaeger[name] for name in std.objectFields(app.jaeger) }
