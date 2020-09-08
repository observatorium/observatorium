local m = import 'memcached.libsonnet';

{
  local mc = self,

  config:: {
    name:: error 'must provide name',
    namespace:: error 'must provide namespace',
    version:: error 'must provide version',
    image:: error 'must provide image',
    exporterVersion:: error 'must provide exporter version',
    exporterImage:: error 'must provide exporter image',
    replicas:: error 'must provide replicas',

    enableChuckCache: false,
    enableIndexQueryCache: false,
    enableIndexWriteCache: false,
    enableResultsCache: false,

    commonLabels:: {
      'app.kubernetes.io/name': 'loki',
      'app.kubernetes.io/instance': mc.config.name,
      'app.kubernetes.io/version': mc.config.version,
    },
  },

  serviceMonitors: {},

  chunkCache::
    m +
    m.withServiceMonitor {
      config+:: {
        local cfg = self,
        name: mc.config.name + '-' + mc.config.commonLabels['app.kubernetes.io/name'] + '-chunk-cache',
        namespace: mc.config.namespace,
        commonLabels+:: mc.config.commonLabels {
          'app.kubernetes.io/component': 'chunk-cache',
        },
        version:: mc.config.version,
        image:: mc.config.image,
        exporterVersion: mc.config.exporterVersion,
        exporterImage:: mc.config.exporterImage,
        replicas: mc.config.replicas.chunk_cache,
        maxItemSize:: '2m',
        memoryLimitMb: 4096,
      },
    } + if std.objectHas(mc.serviceMonitors, 'chunk_cache') then {
      serviceMonitor+: mc.serviceMonitors.chunk_cache,
    } else {},

  indexQueryCache::
    m +
    m.withServiceMonitor {
      config+:: {
        local cfg = self,
        name: mc.config.name + '-' + mc.config.commonLabels['app.kubernetes.io/name'] + '-index-query-cache',
        namespace: mc.config.namespace,
        commonLabels+:: mc.config.commonLabels {
          'app.kubernetes.io/component': 'index-query-cache',
        },
        version:: mc.config.version,
        image:: mc.config.image,
        exporterVersion: mc.config.exporterVersion,
        exporterImage:: mc.config.exporterImage,
        replicas: mc.config.replicas.index_query_cache,
        maxItemSize:: '5m',
      },
    } + if std.objectHas(mc.serviceMonitors, 'index_query_cache') then {
      serviceMonitor+: mc.serviceMonitors.index_query_cache,
    } else {},

  indexWriteCache::
    m +
    m.withServiceMonitor {
      config+:: {
        local cfg = self,
        name: mc.config.name + '-' + mc.config.commonLabels['app.kubernetes.io/name'] + '-index-write-cache',
        namespace: mc.config.namespace,
        commonLabels+:: mc.config.commonLabels {
          'app.kubernetes.io/component': 'index-write-cache',
        },
        version:: mc.config.version,
        image:: mc.config.image,
        exporterVersion: mc.config.exporterVersion,
        exporterImage:: mc.config.exporterImage,
        replicas: mc.config.replicas.index_write_cache,
      },
    } + if std.objectHas(mc.serviceMonitors, 'index_write_cache') then {
      serviceMonitor+: mc.serviceMonitors.index_write_cache,
    } else {},

  resultsCache::
    m +
    m.withServiceMonitor {
      config+:: {
        local cfg = self,
        name: mc.config.name + '-' + mc.config.commonLabels['app.kubernetes.io/name'] + '-results-cache',
        namespace: mc.config.namespace,
        commonLabels+:: mc.config.commonLabels {
          'app.kubernetes.io/component': 'results-cache',
        },
        version:: mc.config.version,
        image:: mc.config.image,
        exporterVersion: mc.config.exporterVersion,
        exporterImage:: mc.config.exporterImage,
        replicas: mc.config.replicas.results_cache,
      },
    } + if std.objectHas(mc.serviceMonitors, 'results_cache') then {
      serviceMonitor+: mc.serviceMonitors.results_cache,
    } else {},


  withServiceMonitors: {
    local l = self,
    serviceMonitors:: {},

    manifests+:: {
      'chunk-cache-service-monitor': l.chunkCache.serviceMonitor,
      'index-query-cache-service-monitor': l.indexQueryCache.serviceMonitor,
      'index-write-cache-service-monitor': l.indexWriteCache.serviceMonitor,
      'results-cache-service-monitor': l.resultsCache.serviceMonitor,
    },
  },

  manifests::
    {} +
    (if mc.config.enableChuckCache then {
       'chunk-cache-service': mc.chunkCache.service,
       'chunk-cache-statefulset': mc.chunkCache.statefulSet,
     } else {}) +
    (if mc.config.enableIndexQueryCache then {
       'index-query-cache-service': mc.indexQueryCache.service,
       'index-query-cache-statefulset': mc.indexQueryCache.statefulSet,
     } else {}) +
    (if mc.config.enableIndexWriteCache then {
       'index-write-cache-service': mc.indexWriteCache.service,
       'index-write-cache-statefulset': mc.indexWriteCache.statefulSet,
     } else {}) +
    (if mc.config.enableResultsCache then {
       'results-cache-service': mc.resultsCache.service,
       'results-cache-statefulset': mc.resultsCache.statefulSet,
     } else {}),
}
