local app =
  (import 'kube-thanos.libsonnet') +
  (import 'telemeter.libsonnet') +
  (import 'jaeger.libsonnet') +
  (import 'memcached.libsonnet');

{ ['thanos-querier-' + name]: app.thanos.querier[name] for name in std.objectFields(app.thanos.querier) } +
{ ['thanos-receive-' + name]: app.thanos.receive[name] for name in std.objectFields(app.thanos.receive) } +
{ ['thanos-compactor-' + name]: app.thanos.compactor[name] for name in std.objectFields(app.thanos.compactor) } +
{ ['thanos-store-' + name]: app.thanos.store[name] for name in std.objectFields(app.thanos.store) } +
{ ['thanos-receive-controller-' + name]: app.thanos.receiveController[name] for name in std.objectFields(app.thanos.receiveController) } +
{ ['thanos-querier-cache-' + name]: app.thanos.querierCache[name] for name in std.objectFields(app.thanos.querierCache) } +
{ ['telemeter-' + name]: app.telemeterServer[name] for name in std.objectFields(app.telemeterServer) } +
{ ['jaeger-' + name]: app.jaeger[name] for name in std.objectFields(app.jaeger) } +
{ ['memcached-' + name]: app.memcached[name] for name in std.objectFields(app.memcached) }
