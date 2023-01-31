// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,

  name: 'observatorium-xyz',
  namespace: error 'must provide namespace',
  version: error 'must provide version',
  image: error 'must provide image',
  imagePullPolicy: 'IfNotPresent',
  replicas: error 'must provide replicas',
  objectStorageConfig: error 'must provide object storage config',
  query: {
    concurrency: error 'must provide max concurrent setting for querier config',
    shardFactor: 16,
    enableSharedQueries: false,
  },
  ports: {
    gossip: 7946,
  },
  replicationFactor: 1,
  shardFactor: 16,
  limits: {
    maxOutstandingPerTenant: 256,
  },

  // TODO(kakkoyun): Is it duplicated with components?
  resources: {},
  volumeClaimTemplates: {},
  memberlist: {},
  etcd: {},
  rules: {},
  rulesStorageConfig: {
    type: 'local',
  },
  wal: {
    replayMemoryCeiling: error 'must provide replay memory ceiling',
  },

  indexQueryCache: '',
  storeChunkCache: '',
  resultsCache: '',

  etcdEndpoints: [],

  components:: {
    compactor: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    distributor: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    } + (
      if defaults.etcdEndpoints != [] then {
        ring: {
          kvstore: {
            store: 'etcd',
            etcd: {
              endpoints: defaults.etcdEndpoints,
            },
          },
        },
      } else {}
    ),
    ingester: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    } + (
      if defaults.etcdEndpoints != [] then {
        lifecycler+: {
          ring+: {
            kvstore: {
              store: 'etcd',
              etcd: {
                endpoints: defaults.etcdEndpoints,
              },
            },
          },
        },
      } else {}
    ),
    index_gateway: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    querier: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    query_frontend: {
      withLivenessProbe: true,
      withReadinessProbe: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    query_scheduler: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    ruler: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
  },

  // Loki config.
  config:: {
    local grpcServerMaxMsgSize = 104857600,
    local querierConcurrency = 32,
    local indexPeriodHours = 24,

    analytics: {
      reporting_enabled: false,
    },
    auth_enabled: true,
    chunk_store_config: {
      max_look_back_period: '0s',
    } + (
      if defaults.storeChunkCache != '' then {
        chunk_cache_config: {
          memcached: {
            batch_size: 100,
            parallelism: 100,
          },
          memcached_client: {
            addresses: defaults.storeChunkCache,
            timeout: '100ms',
            max_idle_conns: 100,
            update_interval: '1m',
            consistent_hash: true,
          },
        },
      } else {
        chunk_cache_config: {
          embedded_cache: {
            enabled: true,
            max_size_mb: 500,
          },
        },
      }
    ),

    memberlist+: if defaults.memberlist != {} then {
      bind_port: defaults.ports.gossip,
      abort_if_cluster_join_fails: false,
      min_join_backoff: '1s',
      max_join_backoff: '1m',
      max_join_retries: 10,
      join_members: [
        '%s.%s.svc.cluster.local:%d' % [
          defaults.name + '-' + defaults.memberlist.ringName,
          defaults.namespace,
          defaults.ports.gossip,
        ],
      ],
    } else {},

    common: {
      compactor_grpc_address: '%s.%s.svc.cluster.local:9095' % [
        defaults.name + '-compactor-grpc',
        defaults.namespace,
      ],
    },

    compactor: {
      compaction_interval: '2h',
      shared_store: 's3',
      working_directory: '/data/loki/compactor',
    },
    distributor: {
      ring: {
        kvstore: {
          store: if defaults.memberlist != {} then 'memberlist' else 'inmemory',
        },
      },
    },
    frontend: {
      scheduler_address: '%s.%s.svc.cluster.local:9095' % [
        defaults.name + '-query-scheduler-grpc',
        defaults.namespace,
      ],
      tail_proxy_url: '%s.%s.svc.cluster.local:3100' % [
        defaults.name + '-querier-http',
        defaults.namespace,
      ],
      compress_responses: true,
    },
    frontend_worker: {
      scheduler_address: '%s.%s.svc.cluster.local:9095' % [
        defaults.name + '-query-scheduler-grpc',
        defaults.namespace,
      ],
      grpc_client_config: {
        max_send_msg_size: grpcServerMaxMsgSize,
      },
      match_max_concurrent: true,
    },
    query_scheduler: {
      max_outstanding_requests_per_tenant:
        if defaults.query.enableSharedQueries
        then 1024
        else defaults.limits.maxOutstandingPerTenant,
    },
    ingester: {
      chunk_block_size: 262144,
      chunk_encoding: 'snappy',
      chunk_idle_period: '1h',
      chunk_retain_period: '5m',
      chunk_target_size: 2097152,
      lifecycler: {
        heartbeat_period: '5s',
        interface_names: [
          'eth0',
        ],
        join_after: if defaults.memberlist != {} then '60s' else '30s',
        num_tokens: 512,
        ring: {
          heartbeat_timeout: '1m',
          kvstore: {
            store: if defaults.memberlist != {} then 'memberlist' else 'inmemory',
          },
        },
      },
      max_transfer_retries: 0,
      wal: {
        enabled: true,
        dir: '/data/loki/wal',
        replay_memory_ceiling: defaults.wal.replayMemoryCeiling,
      },
    },
    ingester_client: {
      grpc_client_config: {
        max_recv_msg_size: 1024 * 1024 * 64,
      },
      remote_timeout: '1s',
    },
    limits_config: {
      ingestion_rate_strategy: 'global',
      ingestion_burst_size_mb: 20,
      ingestion_rate_mb: 10,
      max_label_name_length: 1024,
      max_label_value_length: 2048,
      max_label_names_per_series: 30,
      reject_old_samples: true,
      reject_old_samples_max_age: '%dh' % indexPeriodHours,
      creation_grace_period: '10m',
      enforce_metric_name: false,
      max_streams_per_user: 0,
      max_line_size: 256000,
      max_entries_limit_per_query: 5000,
      max_chunks_per_query: 2000000,
      max_query_length: '721h',
      max_query_parallelism:
        if defaults.query.enableSharedQueries
        then defaults.shardFactor * 16
        else 16,
      max_query_series: 500,
      cardinality_limit: 100000,
      max_global_streams_per_user: 10000,
      max_cache_freshness_per_query: '10m',
      per_stream_rate_limit: '3MB',
      per_stream_rate_limit_burst: '15MB',
      split_queries_by_interval: '30m',
      deletion_mode: 'disabled',
    },
    querier: {
      query_timeout: '1h',
      tail_max_duration: '1h',
      extra_query_delay: '0s',
      query_ingesters_within: '3h',
      engine: {
        max_look_back_period: '30s',
      },
      max_concurrent: defaults.query.concurrency,
    },
    query_range: {
      align_queries_with_step: true,
      cache_results: true,
      max_retries: 5,
      parallelise_shardable_queries: defaults.query.enableSharedQueries,
    } + (
      if defaults.resultsCache != '' then {
        results_cache: {
          cache: {
            memcached_client: {
              timeout: '500ms',
              consistent_hash: true,
              addresses: defaults.resultsCache,
              update_interval: '1m',
              max_idle_conns: 16,
            },
          },
        },
      } else {
        results_cache: {
          cache: {
            embedded_cache: {
              enabled: true,
              max_size_mb: 500,
            },
          },
        },
      }
    ),
    ruler: {
      enable_api: true,
      enable_sharding: true,
      wal: {
        dir: '/data/loki/wal',
        truncate_frequency: '60m',
        min_age: '5m',
        max_age: '4h',
      },
      rule_path: '/data',
      storage:
        (if defaults.rulesStorageConfig.type == 'local' then {
           type: 'local',
           'local': {
             directory: '/tmp/rules',
           },
         } else {
           type: defaults.rulesStorageConfig.type,
         }),
      ring: {
        kvstore: {
          store: if defaults.memberlist != {} then 'memberlist' else 'inmemory',
        },
      },
    },
    schema_config: {
      configs: [
        {
          from: '2020-10-01',
          index: {
            period: '24h',
            prefix: 'loki_index_',
          },
          object_store: 's3',
          schema: 'v11',
          store: 'boltdb-shipper',
        },
      ],
    },
    server: {
      graceful_shutdown_timeout: '5s',
      grpc_server_min_time_between_pings: '10s',
      grpc_server_ping_without_stream_allowed: true,
      grpc_server_max_concurrent_streams: 1000,
      grpc_server_max_recv_msg_size: grpcServerMaxMsgSize,
      grpc_server_max_send_msg_size: grpcServerMaxMsgSize,
      http_listen_port: 3100,
      http_server_idle_timeout: '120s',
      http_server_write_timeout: '1m',
      log_level: 'error',
    },
    storage_config: {
      boltdb_shipper: {
        index_gateway_client+: {
          server_address: '%s.%s.svc.cluster.local:9095' % [
            defaults.name + '-index-gateway-grpc',
            defaults.namespace,
          ],
        },
        active_index_directory: '/data/loki/index',
        cache_location: '/data/loki/index_cache',
        cache_ttl: '24h',
        resync_interval: '5m',
        shared_store: 's3',
      },
    } + (
      if defaults.indexQueryCache != '' then {
        index_queries_cache_config: {
          memcached: {
            batch_size: 100,
            parallelism: 100,
          },
          memcached_client: {
            addresses: defaults.indexQueryCache,
            consistent_hash: true,
          },
        },
      } else {}
    ),
  },

  // Loki config overrides.
  overrides:: {},

  commonLabels:: {
    'app.kubernetes.io/name': 'loki',
    'app.kubernetes.io/part-of': 'observatorium',
    'app.kubernetes.io/instance': defaults.name,
    'app.kubernetes.io/version': defaults.version,
  },

  podLabelSelector:: {
    [labelName]: defaults.commonLabels[labelName]
    for labelName in std.objectFields(defaults.commonLabels)
    if labelName != 'app.kubernetes.io/version'
  },
};

function(params) {
  local loki = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,
  // Safety checks for combined config of defaults and params.
  assert std.isNumber(loki.config.replicationFactor),
  assert std.isNumber(loki.config.query.concurrency),
  assert std.isObject(loki.config.limits) : 'limits has to be an object',
  assert std.isObject(loki.config.replicas) : 'replicas has to be an object',
  assert std.isObject(loki.config.resources) : 'replicas has to be an object',
  assert std.isObject(loki.config.volumeClaimTemplates) : 'volumeClaimTemplates has to be an object',
  assert std.isObject(loki.config.memberlist) : 'memberlist has to be an object',
  assert std.isObject(loki.config.etcd) : 'etcd has to be an object',
  assert std.isArray(loki.config.etcdEndpoints) : 'etcdEndpoints has to be an array',

  configmap:: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: loki.config.name,
      namespace: loki.config.namespace,
      labels: loki.config.commonLabels,
    },
    data: {
      'config.yaml': std.manifestYamlDoc(loki.config.config),
      'overrides.yaml': std.manifestYamlDoc(loki.config.overrides),
    },
  },

  rulesConfigMap:: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: rulesConfigName(),
      namespace: loki.config.namespace,
      labels: loki.config.commonLabels,
    },
    data: {
      [name]: std.manifestYamlDoc(loki.config.rules[name])
      for name in std.objectFields(loki.config.rules)
      if std.length(loki.config.rules) != 0
    },
  },

  local normalizedName(id) =
    std.strReplace(id, '_', '-'),

  local rulesConfigName() =
    loki.config.name + '-rules',

  local newPodLabelsSelector(component) =
    loki.config.podLabelSelector {
      'app.kubernetes.io/component': normalizedName(component),
    },

  local newCommonLabels(component) =
    loki.config.commonLabels {
      'app.kubernetes.io/component': normalizedName(component),
    },

  local joinGossipRing(component) =
    loki.config.config.distributor.ring.kvstore.store == 'memberlist' &&
    loki.config.config.ingester.lifecycler.ring.kvstore.store == 'memberlist' &&
    loki.config.config.ruler.ring.kvstore.store == 'memberlist' &&
    std.member(['distributor', 'ingester', 'querier', 'ruler'], component),

  local isStatefulSet(component) =
    std.member(['compactor', 'ingester', 'index_gateway', 'ruler'], component),

  local newLokiContainer(name, component, config) =
    local osc = loki.config.objectStorageConfig;
    local rsc = loki.config.rulesStorageConfig;
    local replicas = loki.config.replicas[component];

    assert rsc.type == 'local' || rsc.type == 's3';
    assert rsc.type == 's3' &&
           std.objectHas(rsc, 'secretName') && rsc.secretName != '' &&
           ((std.objectHas(rsc, 'endpointKey') && rsc.endpointKey != '') ||
            (std.objectHas(rsc, 'bucketsKey') && rsc.bucketsKey != '' &&
             std.objectHas(rsc, 'regionKey') && rsc.regionKey != '' &&
             std.objectHas(rsc, 'accessKeyIdKey') && rsc.accessKeyIdKey != '' &&
             std.objectHas(rsc, 'secretAccessKeyKey') && rsc.secretAccessKeyKey != ''));

    local readinessProbe = { readinessProbe: {
      initialDelaySeconds: 15,
      timeoutSeconds: 1,
      httpGet: {
        scheme: 'HTTP',
        port: 3100,
        path: '/ready',
      },
    } };

    local livenessProbe = {
      livenessProbe: {
        failureThreshold: 10,
        periodSeconds: 30,
        httpGet: {
          scheme: 'HTTP',
          port: 3100,
          path: '/metrics',
        },
      },
    };

    local rulesVolumeMount = { name: 'rules', mountPath: '/tmp/rules', readOnly: false };

    local resources = { resources: config.resources };
    {
      name: name,
      image: loki.config.image,
      imagePullPolicy: loki.config.imagePullPolicy,
      args: [
        '-target=' + normalizedName(component),
        '-config.file=/etc/loki/config/config.yaml',
        '-limits.per-user-override-config=/etc/loki/config/overrides.yaml',
        '-log.level=error',
      ] + (
        if std.objectHas(osc, 'endpointKey') then [
          '-s3.url=$(S3_URL)',
          '-s3.force-path-style=true',
        ] else [
          '-s3.buckets=$(S3_BUCKETS)',
          '-s3.region=$(S3_REGION)',
          '-s3.access-key-id=$(AWS_ACCESS_KEY_ID)',
          '-s3.secret-access-key=$(AWS_SECRET_ACCESS_KEY)',
        ]
      ) + (
        if component == 'ruler' && std.objectHas(rsc, 'endpointKey') then [
          '-ruler.storage.s3.url=$(RULER_S3_URL)',
          '-ruler.storage.s3.force-path-style=true',
        ] else if component == 'ruler' then [
          '-ruler.storage.s3.buckets=$(RULER_S3_BUCKETS)',
          '-ruler.storage.s3.region=$(RULER_S3_REGION)',
          '-ruler.storage.s3.access-key-id=$(RULER_AWS_ACCESS_KEY_ID)',
          '-ruler.storage.s3.secret-access-key=$(RULER_AWS_SECRET_ACCESS_KEY)',
        ] else []
      ),
      env: (
        if std.objectHas(osc, 'endpointKey') then [
          { name: 'S3_URL', valueFrom: { secretKeyRef: {
            name: osc.secretName,
            key: osc.endpointKey,
          } } },
        ] else [
          { name: 'S3_BUCKETS', valueFrom: { secretKeyRef: {
            name: osc.secretName,
            key: osc.bucketsKey,
          } } },
          { name: 'S3_REGION', valueFrom: { secretKeyRef: {
            name: osc.secretName,
            key: osc.regionKey,
          } } },
          { name: 'AWS_ACCESS_KEY_ID', valueFrom: { secretKeyRef: {
            name: osc.secretName,
            key: osc.accessKeyIdKey,
          } } },
          { name: 'AWS_SECRET_ACCESS_KEY', valueFrom: { secretKeyRef: {
            name: osc.secretName,
            key: osc.secretAccessKeyKey,
          } } },
        ]
      ) + (
        if component == 'ruler' && std.objectHas(rsc, 'endpointKey') then [
          { name: 'RULER_S3_URL', valueFrom: { secretKeyRef: {
            name: rsc.secretName,
            key: rsc.endpointKey,
          } } },
        ] else if component == 'ruler' then [
          { name: 'RULER_S3_BUCKETS', valueFrom: { secretKeyRef: {
            name: rsc.secretName,
            key: rsc.bucketsKey,
          } } },
          { name: 'RULER_S3_REGION', valueFrom: { secretKeyRef: {
            name: rsc.secretName,
            key: rsc.regionKey,
          } } },
          { name: 'RULER_AWS_ACCESS_KEY_ID', valueFrom: { secretKeyRef: {
            name: rsc.secretName,
            key: rsc.accessKeyIdKey,
          } } },
          { name: 'RULER_AWS_SECRET_ACCESS_KEY', valueFrom: { secretKeyRef: {
            name: rsc.secretName,
            key: rsc.secretAccessKeyKey,
          } } },
        ] else []
      ),
      ports: [
        { name: 'metrics', containerPort: 3100 },
        { name: 'grpc', containerPort: 9095 },
      ] + if joinGossipRing(component) then
        [{ name: 'gossip-ring', containerPort: 7946 }]
      else [],
      volumeMounts: [
        { name: 'config', mountPath: '/etc/loki/config/', readOnly: false },
        { name: 'storage', mountPath: '/data', readOnly: false },
      ] + if component == 'ruler' && loki.config.rulesStorageConfig.type == 'local' then [rulesVolumeMount] else [],
    } + {
      [name]: readinessProbe[name]
      for name in std.objectFields(readinessProbe)
      if config.withReadinessProbe
    } + {
      [name]: livenessProbe[name]
      for name in std.objectFields(livenessProbe)
      if config.withLivenessProbe
    } + {
      [name]: resources[name]
      for name in std.objectFields(resources)
      if std.length(config.resources) > 0
    },

  local newDeployment(component, config) =
    local name = loki.config.name + '-' + normalizedName(component);
    local podLabelSelector =
      newPodLabelsSelector(component) +
      if joinGossipRing(component) then
        { 'loki.grafana.com/gossip': 'true' }
      else {};

    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: name,
        namespace: loki.config.namespace,
        labels: newCommonLabels(component),
      },
      spec: {
        replicas: loki.config.replicas[component],
        selector: { matchLabels: podLabelSelector },
        template: {
          metadata: {
            labels: podLabelSelector,
          },
          spec: {
            containers: [newLokiContainer(name, component, config)],
            volumes: [
              { name: 'config', configMap: { name: loki.configmap.metadata.name } },
              { name: 'storage', emptyDir: {} },
            ],
          },
        },
      },
    },

  local newStatefulSet(component, config) =
    local name = loki.config.name + '-' + normalizedName(component);
    local podLabelSelector =
      newPodLabelsSelector(component) +
      if joinGossipRing(component) then
        { 'loki.grafana.com/gossip': 'true' }
      else {};

    local rulesVolume = { name: 'rules', configMap: { name: rulesConfigName() } };

    {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: name,
        namespace: loki.config.namespace,
        labels: newCommonLabels(component),
      },
      spec: {
        replicas: loki.config.replicas[component],
        selector: { matchLabels: podLabelSelector },
        serviceName: newGrpcService(component).metadata.name,
        template: {
          metadata: {
            labels: podLabelSelector,
          },
          spec: {
            containers: [newLokiContainer(name, component, config)],
            volumes: [
              { name: 'config', configMap: { name: loki.configmap.metadata.name } },
            ] + if component == 'ruler' && loki.config.rulesStorageConfig.type == 'local' then [rulesVolume] else [],
            volumeClaimTemplates:: null,
          },
        },
      },
    },

  local newGrpcService(component) = {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: loki.config.name + '-' + normalizedName(component) + '-grpc',
      namespace: loki.config.namespace,
      labels: newCommonLabels(component),
    },
    spec: {
      ports: [
        { name: 'grpc', targetPort: 9095, port: 9095 },
      ],
      selector: newPodLabelsSelector(component),
      clusterIP: 'None',
    },
  },

  local newHttpService(component) = {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: loki.config.name + '-' + normalizedName(component) + '-http',
      namespace: loki.config.namespace,
      labels: newCommonLabels(component),
    },
    spec: {
      ports: [
        { name: 'metrics', targetPort: 3100, port: 3100 },
      ],
      selector: newPodLabelsSelector(component),
    },
  },

  memberlistService:: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: loki.config.name + '-' + loki.config.memberlist.ringName,
      namespace: loki.config.namespace,
      labels: loki.config.commonLabels,
    },
    spec: {
      ports: [
        { name: 'gossip', targetPort: loki.config.ports.gossip, port: loki.config.ports.gossip, protocol: 'TCP' },
      ],
      selector: loki.config.podLabelSelector { 'loki.grafana.com/gossip': 'true' },
      clusterIP: 'None',
    },
  },

  serviceMonitors:: {
    [name]: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: loki.config.name + '-' + normalizedName(name),
        namespace: loki.config.namespace,
        labels: loki.config.commonLabels,
      },
      spec: {
        selector: {
          matchLabels: loki.config.podLabelSelector {
            'app.kubernetes.io/component': normalizedName(name),
          },
        },
        endpoints: [
          { port: 'metrics' },
        ],
      },
    }
    for name in std.objectFields(loki.config.components)
    if std.member(['compactor', 'distributor', 'query_frontend', 'querier', 'query_scheduler', 'ingester', 'index_gateway', 'ruler'], name)
  },

  manifests: {
    'config-map': loki.configmap,
    'rules-config-map': loki.rulesConfigMap,
  } + {
    [normalizedName(name) + '-deployment']: newDeployment(name, loki.config.components[name])
    for name in std.objectFields(loki.config.components)
    if !isStatefulSet(name)
  } + {
    [normalizedName(name) + '-statefulset']: newStatefulSet(name, loki.config.components[name])
    for name in std.objectFields(loki.config.components)
    if isStatefulSet(name)
  } + {
    [normalizedName(name) + '-grpc-service']: newGrpcService(name)
    for name in std.objectFields(loki.config.components)
  } + {
    [normalizedName(name) + '-http-service']: newHttpService(name)
    for name in std.objectFields(loki.config.components)
    if std.member(['compactor', 'distributor', 'query_frontend', 'querier', 'query_scheduler', 'ingester', 'index_gateway', 'ruler'], name)
  } + (
    if std.length(loki.config.resources) != {} then {
      [normalizedName(name) + '-deployment']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  resources: loki.config.resources[name],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(loki.config.resources)
      if !isStatefulSet(name)
    } + {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  resources: loki.config.resources[name],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(loki.config.resources)
      if isStatefulSet(name)
    }
  ) + (
    if std.length(loki.config.replicas) != {} then {
      [normalizedName(name) + '-deployment']+: {
        spec+: {
          replicas: loki.config.replicas[name],
        },
      }
      for name in std.objectFields(loki.config.replicas)
      if !isStatefulSet(name)
    } + {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          replicas: loki.config.replicas[name],
        },
      }
      for name in std.objectFields(loki.config.replicas)
      if isStatefulSet(name)
    }
  ) + (
    if std.length(loki.config.volumeClaimTemplates) != {} then {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          template+: {
            spec+: {
              volumes: std.filter(function(v) v.name != 'storage', super.volumes),
            },
          },
          volumeClaimTemplates: [loki.config.volumeClaimTemplate {
            metadata+: {
              name: 'storage',
              labels+: loki.config.podLabelSelector,
            },
          }],
        },
      }
      for name in std.objectFields(loki.config.components)
      if isStatefulSet(name)
    }
  ) + (
    if loki.config.memberlist != {} then {
      [loki.config.memberlist.ringName]: loki.memberlistService,
    } else {}
  ) + (
    if loki.config.replicationFactor > 0 then {
      [normalizedName(name) + '-deployment']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  args+: [
                    '-distributor.replication-factor=%d' % loki.config.replicationFactor,
                  ],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(loki.config.components)
      if !isStatefulSet(name)
    } + {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  args+: [
                    '-distributor.replication-factor=%d' % loki.config.replicationFactor,
                  ],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(loki.config.components)
      if isStatefulSet(name)
    } else {}
  ) + {
    [normalizedName(name) + '-service-monitor']: loki.serviceMonitors[name]
    for name in std.objectFields(loki.config.components)
    if std.objectHas(loki.serviceMonitors, name) && loki.config.components[name].withServiceMonitor
  },
}
