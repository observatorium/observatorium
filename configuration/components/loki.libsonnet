local lokiMixins = import 'github.com/grafana/loki/production/ksonnet/loki/loki.libsonnet';
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

  // config this field will be merged with the lokiMixins default _config
  // taking precedence
  config:: {},

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

  rhobsLoki::
    lokiMixins {
      _config+:: {
        multi_zone_ingester_enabled: false,
        // Compactor
        using_boltdb_shipper: true,
        ruler_enabled: true,
        query_scheduler_enabled: true,
        use_index_gateway: true,
        memberlist_ring_enabled: true,
        // Label to be used to join gossip ring members
        gossip_member_label: 'loki.grafana.com/gossip',

        // Necessary for generating ConfigMap
        namespace: loki.config.namespace,
        boltdb_shipper_shared_store: 's3',
        cluster: 'bob',
        storage_backend: 's3',
        s3_address: 's3',
        s3_bucket_name: 'bob',

        querier+: {
          // This value should be set equal to (or less than) the CPU cores of the system the querier runs.
          // A higher value will lead to a querier trying to process more requests than there are available
          // cores and will result in scheduling delays.
          concurrency: loki.config.query.concurrency,
          // Necessary to avoid error in resource generation
          use_topology_spread: false,
        },

        queryFrontend+: {
          sharded_queries_enabled: loki.config.query.enableSharedQueries,
        },

        loki+: {
          // analytics config is not yet supported in the loki.mixins
          // We want this because, ... TODO(periklis) help
          analytics: {
            reporting_enabled: false,
          },
          // auth_enabled is disabled by default we want it because, ... TODO(periklis) help
          auth_enabled: true,
          chunk_store_config+: {
            chunk_cache_config+: (
              if loki.config.storeChunkCache == '' then {
                // embedded_cache is not yet supported in the loki.mixins
                // We want this because, ....
                embedded_cache: {
                  enabled: true,
                  max_size_mb: 500,
                },
                // Disabel memcached
                memcached:: {},
                memcached_client:: {},
              } else {
                // memcached_client is configured differently in loki.mixins so
                // let's set our own
                memcached_client: {
                  addresses: loki.config.storeChunkCache,
                  timeout: '100ms',
                  max_idle_conns: 100,
                  update_interval: '1m',
                },
              }
            ),
            // max_look_back_period is not yet supported in the loki.mixins
            // We want this because, ...
            max_look_back_period: '0s',
          },
          common+: {
            // compactor_grpc_address is not yet supported in the loki.mixins
            local compactorService = newService('compactor'),
            compactor_grpc_address: '%s.%s.svc.cluster.local:9095' % [compactorService.metadata.name, loki.config.namespace],
            // Disable compactor_address
            compactor_address:: {},
          },
          compactor+: {
            // compaction_interval is not yet supported in the loki.mixins
            compaction_interval: '2h',
            // loki.mixins set's a different workingdir
            working_directory: '/data/loki/compactor',
          },
          frontend+: {
            // loki.mixins doesn't support prefixes in resources names so we have to overwrite
            local schedulerService = newService('query_scheduler'),
            scheduler_address: '%s.%s.svc.cluster.local:9095' % [schedulerService.metadata.name, loki.config.namespace],
            // loki.mixins configures this parameter, to 5s, do we want it TODO(periklis) config decision
            log_queries_longer_than:: {},
            // loki.mixins doesn't support tail_proxy_url
            local querierService = newService('querier'),
            tail_proxy_url: '%s.%s.svc.cluster.local:3100' % [querierService.metadata.name, loki.config.namespace],
          },
          frontend_worker+: {
            // loki.mixins doesn't support prefixes in resources names so we have to overwrite
            local schedulerService = newService('query_scheduler'),
            scheduler_address: '%s.%s.svc.cluster.local:9095' % [schedulerService.metadata.name, loki.config.namespace],
          },
          ingester+: {
            // chunk_idle_period loki.mixins set's the default to 15m but we want
            // 1h because
            chunk_idle_period: '1h',
            // chunk_idle_period not defined in mixins, we want because
            chunk_encoding: 'snappy',
            // chunk_retain_period not defined in mixins, we want because
            chunk_retain_period: '5m',
            // chunk_target_size not defined in mixins, we want because
            chunk_target_size: 2097152,
            lifecycler+: {
              // join_after mixins define it at 30s, we want 60s because
              join_after: if loki.config.memberlist != {} then '60s' else '30s',
              ring+: {
                // replication_factor mixins set this to 3 not sure if it's interest to us
                // TODO(periklis) config decision
                replication_factor:: {},
              },
            },
            wal+: {
              // loki.mixins set's a different dir
              dir: '/data/loki/wal',
              // replay_memory_ceiling mixins set 7GB by default, we have 100MB
              replay_memory_ceiling: loki.config.wal.replayMemoryCeiling,
            },
          },
          limits_config+: {
            // cardinality_limit mixins don't support this.
            // We want this config because
            cardinality_limit: 100000,
            // creation_grace_period mixins don't support this.
            // We want this config because
            creation_grace_period: '10m',
            // deletion_mode mixins don't support this.
            // We want this config because
            deletion_mode: 'disabled',
            // max_chunks_per_query mixins don't support this.
            // We want this config because
            max_chunks_per_query: 2000000,
            // max_entries_limit_per_query mixins don't support this.
            // We want this config because
            max_entries_limit_per_query: 5000,
            // max_label's mixins don't support this.
            // We want this config because
            max_label_name_length: 1024,
            max_label_names_per_series: 30,
            max_label_value_length: 2048,
            // max_line_size mixins don't support this.
            // We want this config because
            max_line_size: 256000,
            // max_query_length mixins don't support this.
            // We want this cofig because
            max_query_length: '721h',
            // TODO (JoaoBraveCoding) mixins already preforms some calculations better let them handle it
            // max_query_parallelism: if loki.config.query.enableSharedQueries then loki.config.shardFactor * 16 else 16,
            // max_query_series mixins don't support this.
            // We want this cofig because
            max_query_series: 500,
            // per_stream_rate_limit's mixins don't support this.
            // We want this cofig because
            per_stream_rate_limit: '3MB',
            per_stream_rate_limit_burst: '15MB',
            // reject_old_samples_max_age's mixins set's this value to 168h
            // We want this cofig because
            reject_old_samples_max_age: '24h',
          },
          memberlist+: {
            // Both cluster_label + cluster_label_verification_disabled exist only for
            // backwards compatibility with 2.6.1, more info https://github.com/grafana/loki/blob/0030cafb167fd70375399599acd8568c9290746e/production/ksonnet/loki/memberlist.libsonnet#L20-L26
            cluster_label:: {},
            cluster_label_verification_disabled:: {},
            // loki.mixins doesn't support prefixes in resources names so we have to overwrite
            local gossipRingService = newGossipRingService(),
            join_members: ['%s.%s.svc.cluster.local:7946' % [gossipRingService.metadata.name, loki.config.namespace]],
          },
          querier+: {
            engine: {
              // max_look_back_period mixins don't support this.
              // We want this cofig because
              max_look_back_period: '30s',
            },
            // extra_query_delay mixins don't support this.
            // We want this cofig because
            extra_query_delay: '0s',
            // query_ingesters_within should be twice the max-chunk age
            // (1h default) for safety buffer
            // TODO(periklis) validate config and comment, mixins has this set to 2h
            query_ingesters_within: '3h',
            // query_timeout mixins don't support this.
            // We want this cofig because
            query_timeout: '1h',
            // tail_max_duration mixins don't support this.
            // We want this cofig because
            tail_max_duration: '1h',
          },
          query_range+: {
            results_cache+: {
              // cache mixins don't support embedded_cache
              // We want this cofig because
              cache: (
                if loki.config.resultsCache == '' then {
                  embedded_cache: {
                    enabled: true,
                    max_size_mb: 500,
                  },
                } else {
                  memcached_client: {
                    timeout: '500ms',
                    consistent_hash: true,
                    addresses: loki.config.resultsCache,
                    update_interval: '1m',
                    max_idle_conns: 16,
                  },
                }
              ),
            },
          },
          ruler+: {
            // alertmanager_url mixins set's the alertmanager_url
            alertmanager_url:: {},
            enable_alertmanager_v2:: {},
            // rule_path mixins set's the path to /tmp/rules
            rule_path: '/data',
            // storage mixins always configures a gcs bucket
            storage: {
              type: loki.config.rulesStorageConfig.type,
            },
            // wal mixins don't support this.
            // We want this cofig because
            wal: {
              dir: '/data/loki/wal',
              truncate_frequency: '60m',
              min_age: '5m',
              max_age: '4h',
            },
          },
          // schema_config only the index is dynamic in this config so we
          // overwrite it
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
          // server we use the defaults
          server+: {},
          storage_config+: {
            aws:: {},
            boltdb_shipper+: {
              active_index_directory: '/data/loki/index',
              cache_location: '/data/loki/index_cache',
              cache_ttl: '24h',
              index_gateway_client+: {
                local indexService = newService('index_gateway'),
                server_address: '%s.%s.svc.cluster.local:9095' % [indexService.metadata.name, loki.config.namespace],
              },
              resync_interval: '5m',
            },
            index_queries_cache_config:: {},
          } + (
            if loki.config.indexQueryCache != '' then {
              index_queries_cache_config: {
                memcached: {
                  batch_size: 100,
                  parallelism: 100,
                },
                memcached_client: {
                  addresses: loki.config.indexQueryCache,
                  consistent_hash: true,
                },
              },
            } else {}
          ),
          table_manager:: {},
          // This line is important it's what allows us to specify configuration in defaults.config
          // that will take precedence vs the default loki config in this repo
        } + loki.config.config,
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
    loki.rhobsLoki._config.loki.distributor.ring.kvstore.store == 'memberlist' &&
    loki.rhobsLoki._config.loki.ingester.lifecycler.ring.kvstore.store == 'memberlist' &&
    loki.rhobsLoki._config.loki.ruler.ring.kvstore.store == 'memberlist' &&
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
              local lokiConfigMap = newConfigMap();
              { name: 'config', configMap: { name: lokiConfigMap.metadata.name } },
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
        serviceName: name,
        template: {
          metadata: {
            labels: podLabelSelector,
          },
          spec: {
            containers: [newLokiContainer(name, component, config)],
            volumes: [
              local lokiConfigMap = newConfigMap();
              { name: 'config', configMap: { name: lokiConfigMap.metadata.name } },
            ] + if component == 'ruler' && loki.config.rulesStorageConfig.type == 'local' then [rulesVolume] else [],
            volumeClaimTemplates:: null,
          },
        },
      },
    },

  // metadataFormat for a given k8s object add a prefix to name, set namespace
  // and set common labels
  local metadataFormat(object) = object {
    metadata+: {
      name: loki.config.name + '-' + object.metadata.name,
      namespace: loki.config.namespace,
      labels: newCommonLabels(object.metadata.name),
    },
  },

  local newConfigMap() =
    loki.rhobsLoki.config_file {
      metadata+: {
        name: loki.config.name,
        namespace: loki.config.namespace,
        labels: loki.config.commonLabels,
      },
    },

  // newService for a given component, generate its service using the loki mixins
  local newService(component) =
    metadataFormat(loki.rhobsLoki[component + '_service']) {
      spec+: {
        selector: newPodLabelsSelector(component),
      },
    },

  // newGossipRingService is a headless service that has in its selectors the
  // gossip_member_label that we want to preserve
  local newGossipRingService() =
    metadataFormat(loki.rhobsLoki.gossip_ring_service) {
      metadata+: {
        labels: loki.config.commonLabels,
      },
      spec+: {
        selector+: loki.config.podLabelSelector,
      },
    },

  // serviceMonitors generates ServiceMonitors for all the components below, this
  // code can be removed once the loki.mixins improves ServiceMonitor generation
  // to generate 1 ServiceMonitor per component.
  serviceMonitors:: {
    [name]: {
      // DNS SRV records are used for discovering the schedulers or the frontends so the services
      // of these two componets have ports without the "component_name-" prefix
      // See https://github.com/grafana/loki/blob/4721d7efd308e7d85fe03464041179bb1414fe8c/production/ksonnet/loki/common.libsonnet#L25-L33
      local portPrefix = if !std.member(['query_frontend', 'query_scheduler'], name) then normalizedName(name) + '-' else '',
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
          { port: portPrefix + 'http-metrics' },
        ],
      },
    }
    for name in std.objectFields(loki.config.components)
    if std.member(['compactor', 'distributor', 'query_frontend', 'querier', 'query_scheduler', 'ingester', 'index_gateway', 'ruler'], name)
  },

  manifests: {
    //'config-map': loki.configmap,
    'config-map': newConfigMap(),
    'rules-config-map': loki.rulesConfigMap,
  } + {
    // Service generation for all the components
    [normalizedName(component) + '-service']: newService(component)
    for component in std.objectFields(loki.config.components)
  } + (
    // Service generation for gossip ring
    if loki.config.memberlist != {} then {
      [loki.config.memberlist.ringName]: newGossipRingService(),
    }
  ) + {
    [normalizedName(name) + '-deployment']: newDeployment(name, loki.config.components[name])
    for name in std.objectFields(loki.config.components)
    if !isStatefulSet(name)
  } + {
    [normalizedName(name) + '-statefulset']: newStatefulSet(name, loki.config.components[name])
    for name in std.objectFields(loki.config.components)
    if isStatefulSet(name)
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
    // Service Monitor generation for all the componets that enable it
    [normalizedName(name) + '-service-monitor']: loki.serviceMonitors[name]
    for name in std.objectFields(loki.config.components)
    if std.objectHas(loki.serviceMonitors, name) && loki.config.components[name].withServiceMonitor
  },
}
