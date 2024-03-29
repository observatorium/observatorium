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
    // This value should be set equal to (or less than) the CPU cores of the system the querier runs.
    // A higher value will lead to a querier trying to process more requests than there are available
    // cores and will result in scheduling delays.
    concurrency: error 'must provide max concurrent setting for querier config',
    shardFactor: 16,
    enableSharedQueries: false,
  },
  ruler: {
    externalUrl: '',
  },
  replicationFactor: 1,
  logLevel: 'info',
  shardFactor: 16,
  limits: {
    maxOutstandingPerTenant: 256,
  },
  // Default ports used by the different loki components
  // These values are used to template some parts of the config
  ports: {
    http: 3100,
    grpc: 9095,
    gossip: 7946,
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
      withPodAntiAffinity: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    } + (
      if defaults.etcdEndpoints != [] then {
        compactor_ring+: {
          kvstore: {
            store: 'etcd',
            etcd: {
              endpoints: defaults.etcdEndpoints,
            },
          },
        },
      } else {}
    ),
    distributor: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      withPodAntiAffinity: false,
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
      withPodAntiAffinity: false,
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
      withPodAntiAffinity: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    querier: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      withPodAntiAffinity: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    query_frontend: {
      withLivenessProbe: true,
      withReadinessProbe: false,
      withPodAntiAffinity: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    query_scheduler: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      withPodAntiAffinity: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
      withServiceMonitor: false,
    },
    ruler: {
      withLivenessProbe: true,
      withReadinessProbe: true,
      withPodAntiAffinity: false,
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
  assert std.isNumber(loki.config.replicationFactor) || std.isString(loki.config.replicationFactor),
  assert std.isNumber(loki.config.query.concurrency) || std.isString(loki.config.query.concurrency),
  assert std.isNumber(loki.config.wal.replayMemoryCeiling) || std.isString(loki.config.wal.replayMemoryCeiling),
  assert std.isObject(loki.config.limits) : 'limits has to be an object',
  assert std.isObject(loki.config.replicas) : 'replicas has to be an object',
  assert std.isObject(loki.config.resources) : 'replicas has to be an object',
  assert std.isObject(loki.config.volumeClaimTemplates) : 'volumeClaimTemplates has to be an object',
  assert std.isObject(loki.config.memberlist) : 'memberlist has to be an object',
  assert std.isObject(loki.config.etcd) : 'etcd has to be an object',
  assert std.isArray(loki.config.etcdEndpoints) : 'etcdEndpoints has to be an array',

  mixins::
    lokiMixins {
      // Some values in the loki config are ENV var, we do this because upstream
      // we will want to update them without having to regenerate the templates
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
        cluster: loki.name,
        namespace: loki.config.namespace,
        boltdb_shipper_shared_store: 's3',
        storage_backend: 's3',
        replication_factor: '${LOKI_REPLICATION_FACTOR}',

        querier+: {
          concurrency: '${LOKI_QUERIER_MAX_CONCURRENCY}',
        },

        queryFrontend+: {
          // This value will set the value of parallelise_shardable_queries and
          // have an impact in the generation of the query_frontend and query_scheduler
          // See parallelise_shardable_queries https://grafana.com/docs/loki/latest/configuration/#query_range
          sharded_queries_enabled: loki.config.query.enableSharedQueries,
        },

        // Docs: https://grafana.com/docs/loki/latest/configuration/#supported-contents-and-default-values-of-lokiyaml
        loki+: {
          // Disable reporting analytics to grafana.com.
          analytics: {
            reporting_enabled: false,
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#chunk_store_config
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L296-L309
          chunk_store_config+: {
            chunk_cache_config+: (
              if loki.config.storeChunkCache == '' then {
                // Only necessary for CI tests
                embedded_cache: {
                  enabled: true,
                  // Default is 100 we want 500, review value
                  max_size_mb: 500,
                },
                // Disable memcached as it's enabled by default in the mixins
                memcached:: {},
                memcached_client:: {},
              } else {
                // Configured differently in loki.mixins
                memcached_client: {
                  addresses: loki.config.storeChunkCache,
                  // Default is 16 we want 100, review value
                  max_idle_conns: 100,
                },
              }
            ),
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#common
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L167-L169
          common+: {
            // compactor_grpc_address is not yet supported in the loki.mixins
            // when supported we will still need support for prefixing resource names
            // before removing this
            local compactorService = newService('compactor'),
            compactor_grpc_address: '%s.%s.svc.cluster.local:%s' % [compactorService.metadata.name, loki.config.namespace, loki.config.ports.grpc],
            // Disable compactor_address as it's enabled by default in the mixins
            compactor_address:: {},
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#compactor
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/boltdb_shipper.libsonnet#L27-L30
          compactor+: {
            // Nowadays the default is 10m but previously it was 2h. We want to
            // keep this value for now, as it's still unclear the benefit we would gain.
            compaction_interval: '2h',
            // Mixins set's a different working directory, we might be able to remove this
            // once we move all componets to mixins
            working_directory: '/data/loki/compactor',
            compactor_ring: {
              kvstore: {
                store: 'memberlist',
              },
            },
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#frontend
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L181-L184
          frontend+: {
            // Mixins don't support prefixes in resources names so we have to overwrite
            local schedulerService = newHeadlessService('query_scheduler'),
            scheduler_address: '%s.%s.svc.cluster.local:%s' % [schedulerService.metadata.name, loki.config.namespace, loki.config.ports.grpc],
            // Mixins don't support tail_proxy_url, when support is added we will
            // still need support for prefixing resource names before removing this
            local querierService = newService('querier'),
            tail_proxy_url: '%s.%s.svc.cluster.local:%s' % [querierService.metadata.name, loki.config.namespace, loki.config.ports.http],
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#frontend_worker
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L185-L190
          frontend_worker+: {
            // Mixins dont support prefixes in resources names so we have to overwrite
            local schedulerService = newHeadlessService('query_scheduler'),
            scheduler_address: '%s.%s.svc.cluster.local:%s' % [schedulerService.metadata.name, loki.config.namespace, loki.config.ports.grpc],
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#ingester
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L234-L257
          ingester+: {
            // All the following fields overwrite the loki.mixins since the mixins
            // still suggest using gzip which we know it performs worst than snappy
            chunk_idle_period: '1h',
            chunk_encoding: 'snappy',
            chunk_retain_period: '5m',
            chunk_target_size: 2097152,
            wal+: {
              // loki.mixins set's a different directory, we might be able to remove this
              // once we move all componets to mixins
              dir: '/data/loki/wal',
              replay_memory_ceiling: '${LOKI_INGESTER_WAL_REPLAY_MEMORY_CEILING}',
            },
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#limits_config
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L214-L232
          // Some of the limits set here are the same as their defaults, we want to have them explicit to
          // facilitate reasoning about the service whenever we look at the config or try to debug issues.
          limits_config+: {
            // Although deletion_mode is by default false we want to make it explicit
            // because the Deletion API has caused numerous issues upstream.
            deletion_mode: 'disabled',
            max_line_size: 256000,
            // We want 721h (30 days) as that is our service agreement
            max_query_length: '721h',
            // Default is 168h if not configured, we want 24h.
            reject_old_samples_max_age: '24h',
            cardinality_limit: 100000,
            creation_grace_period: '10m',
            max_chunks_per_query: 2000000,
            max_entries_limit_per_query: 5000,
            max_label_name_length: 1024,
            max_label_names_per_series: 30,
            max_label_value_length: 2048,
            max_query_series: 500,
            per_stream_rate_limit: '3MB',
            per_stream_rate_limit_burst: '15MB',
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#memberlist_config
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/memberlist.libsonnet#L79-L96
          memberlist+: {
            // Both cluster_label + cluster_label_verification_disabled exist only for
            // backwards compatibility with 2.6.1, more info https://github.com/grafana/loki/blob/0030cafb167fd70375399599acd8568c9290746e/production/ksonnet/loki/memberlist.libsonnet#L20-L26
            cluster_label:: {},
            cluster_label_verification_disabled:: {},
            // loki.mixins doesn't support prefixes in resources names so we have to overwrite
            local gossipRingService = newGossipRingService(),
            join_members: ['%s.%s.svc.cluster.local:%s' % [gossipRingService.metadata.name, loki.config.namespace, loki.config.ports.gossip]],
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#query_range
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L191-L209
          query_range+: {
            results_cache+: {
              cache: (
                if loki.config.resultsCache == '' then {
                  // Only necessary for CI tests
                  embedded_cache: {
                    enabled: true,
                    // Default is 100 we want 500, review value
                    max_size_mb: 500,
                  },
                } else {
                  // Mixins still use host field where we want to use addresses so we
                  // overwrite
                  memcached_client: {
                    // Default is 100ms, we want 500ms, review value
                    timeout: '500ms',
                    addresses: loki.config.resultsCache,
                  },
                }
              ),
            },
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#ruler
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L348-L368
          ruler+: {
            // Alertmanager config will be set downstream and unhidden with :::
            alertmanager_url:: {},
            enable_alertmanager_v2:: {},
            // Mixins set's the rule path to /tmp/rules
            rule_path: '/data',
            // Mixins always configures the type to be a gcs bucket
            storage: {
              type: loki.config.rulesStorageConfig.type,
              s3: (
                // Mixins only supports gcp buckets for ruller, so we have to
                // build this structure ourselves
                local rsc = loki.config.rulesStorageConfig;
                if std.objectHas(rsc, 'endpointKey') then {
                  s3: '${RULER_S3_URL}',
                  s3forcepathstyle: true,
                } else {
                  bucketnames: '${RULER_S3_BUCKETS}',
                  region: '${RULER_S3_REGION}',
                  access_key_id: '${RULER_AWS_ACCESS_KEY_ID}',
                  secret_access_key: '${RULER_AWS_SECRET_ACCESS_KEY}',
                }
              ),
            },
            wal: {
              // Default is "ruler-wal", we want it to be /data/loki/wal because, we don't want
              // for PVC to change
              dir: '/data/loki/wal',
            },
          } + (
            if loki.config.ruler.externalUrl != '' then {
              external_url: loki.config.ruler.externalUrl,
            } else {}
          ),
          // Docs: https://grafana.com/docs/loki/latest/configuration/#schema_config
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L311-L323
          schema_config: {
            // Mixins configures for gcs, where we want s3
            configs: [
              {
                from: '2020-10-01',
                index: {
                  // Default is 168h we want 24h because, that's how we have managed
                  // indexes since the start of the service
                  period: '24h',
                  prefix: 'loki_index_',
                },
                object_store: 's3',
                schema: 'v11',
                store: 'boltdb-shipper',
              },
            ],
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#server
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L170-L180
          server+: {
            log_level: '${LOKI_LOG_LEVEL}',
          },
          // Docs: https://grafana.com/docs/loki/latest/configuration/#storage_config
          // Defaults: https://github.com/grafana/loki/blob/main/production/ksonnet/loki/config.libsonnet#L266-L295
          storage_config+: {
            aws: (
              // Mixins only support one s3 bucket, so we have to build this structure ourselves
              local osc = loki.config.objectStorageConfig;
              if std.objectHas(osc, 'endpointKey') then {
                s3: '${S3_URL}',
                s3forcepathstyle: true,
              } else {
                bucketnames: '${S3_BUCKETS}',
                region: '${S3_REGION}',
                access_key_id: '${AWS_ACCESS_KEY_ID}',
                secret_access_key: '${AWS_SECRET_ACCESS_KEY}',
              }
            ),
            boltdb_shipper+: {
              // Mixins sets a different directory
              active_index_directory: '/data/loki/index',
              cache_location: '/data/loki/index_cache',
              // Mixins don't support prefixes in resources names so we have to overwrite
              index_gateway_client+: {
                local indexService = newService('index_gateway'),
                server_address: '%s.%s.svc.cluster.local:%s' % [indexService.metadata.name, loki.config.namespace, loki.config.ports.grpc],
              },
            },
            // Since we only want to enable this when indexQueryCache is set
            // we have to disable it
            index_queries_cache_config:: {},
          } + (
            if loki.config.indexQueryCache != '' then {
              index_queries_cache_config::: {
                memcached: {
                  // Default is 1024, we want 100 because, review value
                  batch_size: 100,
                },
                memcached_client: {
                  // mixins assumes a service format for the cache, we want to overwrite it
                  addresses: loki.config.indexQueryCache,
                },
              },
            } else {}
          ),
          // Currently we don't want to configure the table manager
          table_manager:: {},
          // This line is important it's what allows us to specify configuration in defaults.config
          // that will take precedence vs the default loki config in this repo
        } + loki.config.config,
      },
      // Fix the query_scheduler headless service name on loki.mixins
      query_scheduler_headless_service_name:: 'query-scheduler-headless',
    },

  // newConfigMap will return a ConfigMap with the loki config and overrides.
  // Changes to the configuration should be introduced in loki.mixins._config
  // and not in here
  local newConfigMap() =
    // List of ENV vars that should show in the loki config without quotes
    local envVars = ['${LOKI_REPLICATION_FACTOR}', '${LOKI_QUERIER_MAX_CONCURRENCY}', '${LOKI_INGESTER_WAL_REPLAY_MEMORY_CEILING}'];
    local config = loki.mixins.config_file.data['config.yaml'];
    loki.mixins.config_file {
      metadata+: {
        name: loki.config.name,
        namespace: loki.config.namespace,
        labels: loki.config.commonLabels,
      },
      data+: {
        // Remove quotes from env vars in the array envVars and return
        'config.yaml': removeDoubleQuotes(config, envVars),
        'overrides.yaml': std.manifestYamlDoc(loki.config.overrides),
      },
    },


  // Loop over fromArray (of strings), running std.stdReplace() to remove "
  local removeDoubleQuotes(str, fromArray) = std.foldl(
    function(retStr, from) std.strReplace(retStr, '"' + from + '"', from),
    fromArray,
    str
  ),

  // buildEnvVars compiles a list of parameters whose values are only
  // known at rollout time. Loki config supports env variable expansion, we
  // leverage this feature to inject these config values at rollout time.
  // Some values are only specific to some component but since all components
  // consume the same ConfigMap we decided for all components share the same
  // of env variables.
  local buildConfigVars(component) =
    // Syntactic sugar
    local envVarFromValue(name, value) = { name: name, value: value };
    local envVarFromSecret(name, secretName, secretKey) = { name: name, valueFrom: { secretKeyRef: { name: secretName, key: secretKey } } };
    local osc = loki.config.objectStorageConfig;
    local rsc = loki.config.rulesStorageConfig;
    [
      envVarFromValue('LOKI_LOG_LEVEL', loki.config.logLevel),
      // The following values are added with '', because env only supports strings
      envVarFromValue('LOKI_REPLICATION_FACTOR', '' + loki.config.replicationFactor),
      envVarFromValue('LOKI_QUERIER_MAX_CONCURRENCY', '' + loki.config.query.concurrency),
      envVarFromValue('LOKI_INGESTER_WAL_REPLAY_MEMORY_CEILING', '' + loki.config.wal.replayMemoryCeiling),
    ] + (
      if loki.config.ruler.externalUrl != '' then [
        envVarFromValue('ALERTMANAGER_EXTERNAL_URL', loki.config.ruler.externalUrl),
      ] else []
    ) + (
      if std.objectHas(osc, 'endpointKey') then [
        envVarFromSecret('S3_URL', osc.secretName, osc.endpointKey),
      ] else [
        envVarFromSecret('S3_BUCKETS', osc.secretName, osc.bucketsKey),
        envVarFromSecret('S3_REGION', osc.secretName, osc.regionKey),
        envVarFromSecret('AWS_ACCESS_KEY_ID', osc.secretName, osc.accessKeyIdKey),
        envVarFromSecret('AWS_SECRET_ACCESS_KEY', osc.secretName, osc.secretAccessKeyKey),
      ]
    ) + (
      if component == 'ruler' && std.objectHas(rsc, 'endpointKey') then [
        envVarFromSecret('RULER_S3_URL', rsc.secretName, rsc.endpointKey),
      ] else if component == 'ruler' then [
        envVarFromSecret('RULER_S3_BUCKETS', rsc.secretName, rsc.bucketsKey),
        envVarFromSecret('RULER_S3_REGION', rsc.secretName, rsc.regionKey),
        envVarFromSecret('RULER_AWS_ACCESS_KEY_ID', rsc.secretName, rsc.accessKeyIdKey),
        envVarFromSecret('RULER_AWS_SECRET_ACCESS_KEY', rsc.secretName, rsc.secretAccessKeyKey),
      ] else []
    ),


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
    loki.mixins._config.loki.distributor.ring.kvstore.store == 'memberlist' &&
    loki.mixins._config.loki.ingester.lifecycler.ring.kvstore.store == 'memberlist' &&
    loki.mixins._config.loki.ruler.ring.kvstore.store == 'memberlist' &&
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

    local readinessProbe = {
      readinessProbe: {
        initialDelaySeconds: 15,
        timeoutSeconds: 1,
        httpGet: {
          scheme: 'HTTP',
          port: 3100,
          path: '/ready',
        },
      },
    };

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

    // Syntactic sugar
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
        // Necessary flag to support env variable expansion
        '-config.expand-env=true',
      ],
      env: buildConfigVars(component),
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

  local affinity(component, namespace) = {
    podAntiAffinity: {
      preferredDuringSchedulingIgnoredDuringExecution: [
        {
          weight: 100,
          podAffinityTerm: {
            topologyKey: 'kubernetes.io/hostname',
            namespaces: [namespace],
            labelSelector: {
              matchExpressions: [
                {
                  key: 'app.kubernetes.io/component',
                  operator: 'In',
                  values: [normalizedName(component)],
                },
              ],
            },
          },
        },
      ],
    },
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
          } + if config.withPodAntiAffinity then {
            affinity: affinity(component, loki.config.namespace),
          } else {},
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
          } + if config.withPodAntiAffinity then {
            affinity: affinity(component, loki.config.namespace),
          } else {},
        },
      },
    },

  // metadataFormat for a given k8s object add a prefix to name, set namespace
  // and set common labels
  local metadataFormat(component, object) = object {
    metadata+: {
      name: loki.config.name + '-' + object.metadata.name,
      namespace: loki.config.namespace,
      labels: newCommonLabels(component),
    },
  },

  // newService for a given component, generate its service using the loki mixins
  local newService(component) =
    metadataFormat(component, loki.mixins[component + '_service']) {
      spec+: {
        selector: newPodLabelsSelector(component),
      },
    },

  // newHttpService creates a http service for a component
  // TODO: delete this function once we have merged and deployed https://github.com/rhobs/configuration/pull/450
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

  // newHeadlessService for a given component, generate its headless service using the loki mixins
  // headless services are usually required for DNS SRV dicovery
  local newHeadlessService(component) =
    metadataFormat(component, loki.mixins[component + '_headless_service']) {
      spec+: {
        selector: newPodLabelsSelector(component),
      },
    },

  // newGossipRingService is a headless service that has in its selectors the
  // gossip_member_label that we want to preserve
  local newGossipRingService() =
    metadataFormat('gossip', loki.mixins.gossip_ring_service) {
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
    'config-map': newConfigMap(),
    'rules-config-map': loki.rulesConfigMap,
  } + {
    // Service generation for all the components except query_schedulers
    // see below comment
    [normalizedName(component) + '-service']: newService(component)
    for component in std.objectFields(loki.config.components)
    if !std.member(['query_scheduler'], component)
  } + {
    // Headless service generation for query_schedulers to allow frontend components
    // to discover them using DNS SRV records
    [normalizedName(component) + '-headless-service']: newHeadlessService(component)
    for component in std.objectFields(loki.config.components)
    if std.member(['query_scheduler'], component)
  } + {
    // Create the old http services
    // TODO: delete this function once we have merged and deployed https://github.com/rhobs/configuration/pull/450
    [normalizedName(component) + '-http-service']: newHttpService(component)
    for component in std.objectFields(loki.config.components)
    if std.member(['query_frontend', 'querier', 'distributor', 'ruler'], component)
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
  ) + {
    // Service Monitor generation for all the componets that enable it
    [normalizedName(name) + '-service-monitor']: loki.serviceMonitors[name]
    for name in std.objectFields(loki.config.components)
    if std.objectHas(loki.serviceMonitors, name) && loki.config.components[name].withServiceMonitor
  },
}
