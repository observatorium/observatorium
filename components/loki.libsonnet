{
  local loki = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',
    objectStorageConfig: error 'must provide object storage config',
    queryConcurrency: 32,
    queryParallelism: 32,  // Defaults to queryConcurrency because single query-frontend replica

    commonLabels:: {
      'app.kubernetes.io/name': 'loki',
      'app.kubernetes.io/instance': 'loki',
      'app.kubernetes.io/version': loki.config.version,
    },

    podLabelSelector:: {
      [labelName]: loki.config.commonLabels[labelName]
      for labelName in std.objectFields(loki.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  configmap:: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: loki.config.name,
      namespace: loki.config.namespace,
      labels: loki.config.commonLabels,
    },
    data: {
      'config.yaml': std.manifestYamlDoc(loki.defaultConfig),
      'overrides.yaml': std.manifestYamlDoc(loki.defaultOverrides),
    },
  },

  local normalizedName(id) =
    std.strReplace(id, '_', '-'),

  local newPodLabelsSelector(component) =
    loki.config.podLabelSelector {
      'app.kubernetes.io/component': normalizedName(component),
    },

  local newCommonLabels(component) =
    loki.config.commonLabels {
      'app.kubernetes.io/component': normalizedName(component),
    },

  local joinGossipRing(component) =
    loki.defaultConfig.distributor.ring.kvstore.store == 'memberlist' &&
    loki.defaultConfig.ingester.lifecycler.ring.kvstore.store == 'memberlist' &&
    std.member(['distributor', 'ingester', 'querier'], component),

  local isStatefulSet(component) =
    std.member(['compactor', 'ingester', 'querier'], component),

  local newLokiContainer(name, component, config) =
    local osc = loki.config.objectStorageConfig;
    local replicas = loki.config.replicas[component];

    local readinessProbe = { readinessProbe: {
      initialDelaySeconds: 15,
      timeoutSeconds: 1,
      httpGet: {
        scheme: 'HTTP',
        port: 3100,
        path: '/ready',
      },
    } };

    local resources = { resources: config.resources };

    {
      name: name,
      image: loki.config.image,
      args: [
        '-target=' + normalizedName(component),
        '-config.file=/etc/loki/config/config.yaml',
        '-limits.per-user-override-config=/etc/loki/config/overrides.yaml',
        '-log.level=error',
      ] + if std.objectHas(osc, 'endpointKey') then [
        '-s3.url=$(S3_URL)',
        '-s3.force-path-style=true',
      ] else [
        '-s3.buckets=$(S3_BUCKETS)',
        '-s3.region=$(S3_REGION)',
        '-s3.access-key-id=$(AWS_ACCESS_KEY_ID)',
        '-s3.secret-access-key=$(AWS_SECRET_ACCESS_KEY)',
      ],
      env: if std.objectHas(osc, 'endpointKey') then [
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
      ],
      ports: [
        { name: 'metrics', containerPort: 3100 },
        { name: 'grpc', containerPort: 9095 },
      ] + if joinGossipRing(component) then
        [{ name: 'gossip-ring', containerPort: 7946 }]
      else [],
      volumeMounts: [
        { name: 'config', mountPath: '/etc/loki/config/', readOnly: false },
        { name: 'storage', mountPath: '/data', readOnly: false },
      ],
    } + {
      [name]: readinessProbe[name]
      for name in std.objectFields(readinessProbe)
      if config.withReadinessProbe
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
            ],
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

  components:: {
    compactor: {
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
    },
    distributor: {
      withReadinessProbe: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
    },
    ingester: {
      withReadinessProbe:: true,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
    },
    querier: {
      withReadinessProbe:: true,
      resources:: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
    },
    query_frontend: {
      withReadinessProbe: false,
      resources: {
        requests: { cpu: '100m', memory: '100Mi' },
        limits: { cpu: '200m', memory: '200Mi' },
      },
    },
  },

  defaultConfig+:: {
    local grpcServerMaxMsgSize = 104857600,
    local querierConcurrency = 32,
    local indexPeriodHours = 24,

    auth_enabled: true,
    chunk_store_config: {
      max_look_back_period: '0s',
    },
    // Warning: Do not add join_members here use withMemberList
    // memberlist: {},
    compactor: {
      compaction_interval: '2h',
      shared_store: 's3',
      working_directory: '/data/loki/compactor',
    },
    distributor: {
      ring: {
        kvstore: {
          store: 'inmemory',
        },
      },
    },
    frontend: {
      compress_responses: true,
      max_outstanding_per_tenant: 200,
    },
    frontend_worker: {
      frontend_address: '%s.%s.svc.cluster.local:9095' % [
        loki.config.name + '-query-frontend-grpc',
        loki.config.namespace,
      ],
      grpc_client_config: {
        max_send_msg_size: grpcServerMaxMsgSize,
      },
      parallelism: loki.config.queryParallelism,
    },
    ingester: {
      chunk_block_size: 262144,
      chunk_encoding: 'snappy',
      chunk_idle_period: '2h',
      chunk_retain_period: '1m',
      chunk_target_size: 1.572864e+06,
      lifecycler: {
        heartbeat_period: '5s',
        interface_names: [
          'eth0',
        ],
        join_after: '30s',
        num_tokens: 512,
        ring: {
          heartbeat_timeout: '1m',
          kvstore: {
            store: 'inmemory',
          },
          replication_factor: 1,
        },
      },
      max_transfer_retries: 0,
    },
    ingester_client: {
      grpc_client_config: {
        max_recv_msg_size: 1024 * 1024 * 64,
      },
      remote_timeout: '1s',
    },
    limits_config: {
      enforce_metric_name: false,
      ingestion_burst_size_mb: 20,
      ingestion_rate_mb: 10,
      ingestion_rate_strategy: 'global',
      max_cache_freshness_per_query: '10m',
      max_global_streams_per_user: 10000,
      max_query_length: '12000h',
      max_query_parallelism: 32,
      max_streams_per_user: 0,
      reject_old_samples: true,
      reject_old_samples_max_age: '%dh' % indexPeriodHours,
    },
    querier: {
      query_timeout: '1h',
      tail_max_duration: '1h',
      extra_query_delay: '0s',
      query_ingesters_within: '2h',
      engine: {
        timeout: '3m',
        max_look_back_period: '5m',
      },
    },
    query_range: {
      align_queries_with_step: true,
      cache_results: true,
      max_retries: 5,
      split_queries_by_interval: '30m',
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
      grpc_server_max_concurrent_streams: 1000,
      grpc_server_max_recv_msg_size: grpcServerMaxMsgSize,
      grpc_server_max_send_msg_size: grpcServerMaxMsgSize,
      http_listen_port: 3100,
      http_server_idle_timeout: '120s',
      http_server_write_timeout: '1m',
    },
    storage_config: {
      boltdb_shipper: {
        active_index_directory: '/data/loki/index',
        cache_location: '/data/loki/index_cache',
        cache_ttl: '24h',
        resync_interval: '5m',
        shared_store: 's3',
      },
    },
  },

  defaultOverrides:: {},

  withConfig:: {
    local l = self,
    config+:: {
      config: error 'must provide loki config',
    },

    assert l.defaultConfig.auth_enabled == true : 'Disabling auth not allowed in multi-tenancy',
    assert l.defaultConfig.distributor.ring.kvstore.store != 'memberlist' : 'Use withMemberList to configure memberlist store',
    assert l.defaultConfig.ingester.lifecycler.ring.kvstore.store != 'memberlist' : 'Use withMemberList to configure memberlist store',

    defaultConfig+:: l.config.config,
  },

  withOverrides:: {
    local l = self,
    config+:: {
      overrides: error 'must provide loki config overrides',
    },
    defaultOverrides+:: l.config.overrides,
  },

  withReplicas:: {
    local l = self,
    config+:: {
      replicas: error 'must provide replicas per component',
    },

    assert l.config.replicas.compactor == 1 : 'Only one replica for compactor component allowed',

    manifests+:: {
      [normalizedName(name) + '-deployment']+: {
        spec+: {
          replicas: l.config.replicas[name],
        },
      }
      for name in std.objectFields(l.config.replicas)
      if !isStatefulSet(name)
    } + {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          replicas: l.config.replicas[name],
        },
      }
      for name in std.objectFields(l.config.replicas)
      if isStatefulSet(name)
    },
  },

  withChunkStoreCache:: {
    local l = self,
    config+:: {
      chunkCache: error 'must provide addresses for chunk store cache',
    },

    defaultConfig+:: {
      chunk_store_config+: {
        chunk_cache_config: {
          memcached: {
            batch_size: 100,
            parallelism: 100,
          },
          memcached_client: {
            addresses: l.config.chunkCache,
            timeout: '100ms',
            max_idle_conns: 100,
            update_interval: '1m',
            consistent_hash: true,
          },
        },
      },
    },
  },

  withIndexQueryCache:: {
    local l = self,
    config+:: {
      indexQueryCache: error 'must provide addresses for index query cache',
    },

    defaultConfig+:: {
      storage_config+: {
        index_queries_cache_config: {
          memcached: {
            batch_size: 100,
            parallelism: 100,
          },
          memcached_client: {
            addresses: l.config.indexQueryCache,
            consistent_hash: true,
          },
        },
      },
    },
  },

  withResultsCache:: {
    local l = self,
    config+:: {
      resultsCache: error 'must provide addresses for frontend results cache',
    },

    defaultConfig+:: {
      query_range+: {
        split_queries_by_interval: '30m',
        align_queries_with_step: true,
        cache_results: true,
        max_retries: 5,
        results_cache: {
          cache: {
            memcached_client: {
              timeout: '500ms',
              consistent_hash: true,
              addresses: l.config.resultsCache,
              update_interval: '1m',
              max_idle_conns: 16,
            },
          },
        },
      },
    },
  },

  withEtcd:: {
    local l = self,

    config+:: {
      etcdEndpoints: error 'must provide etcd endpoints list',
    },

    defaultConfig+:: {
      distributor+: {
        ring: {
          kvstore: {
            store: 'etcd',
            etcd: {
              endpoints: l.config.etcdEndpoints,
            },
          },
        },
      },
      ingester+: {
        lifecycler+: {
          ring+: {
            kvstore: {
              store: 'etcd',
              etcd: {
                endpoints: l.config.etcdEndpoints,
              },
            },
          },
        },
      },
    },
  },

  withMemberList:: {
    local l = self,
    local gossipRingName = 'gossip-ring',
    local gossipPort = 7946,
    local gossipSvc =
      {
        apiVersion: 'v1',
        kind: 'Service',
        metadata: {
          name: l.config.name + '-' + gossipRingName,
          namespace: l.config.namespace,
          labels: l.config.commonLabels,
        },
        spec: {
          ports: [
            { name: gossipRingName, targetPort: gossipPort, port: gossipPort, protocol: 'TCP' },
          ],
          selector: l.config.podLabelSelector { 'loki.grafana.com/gossip': 'true' },
          clusterIP: 'None',
        },
      },

    defaultConfig+:: {
      distributor+: {
        ring: {
          kvstore: {
            store: 'memberlist',
          },
        },
      },
      ingester+: {
        lifecycler+: {
          ring+: {
            kvstore: {
              store: 'memberlist',
            },
          },
          join_after: '60s',
        },
      },
      memberlist+: {
        bind_port: gossipPort,
        abort_if_cluster_join_fails: false,
        min_join_backoff: '1s',
        max_join_backoff: '1m',
        max_join_retries: 10,
        join_members: [
          '%s.%s.svc.cluster.local:%d' % [
            gossipSvc.metadata.name,
            gossipSvc.metadata.namespace,
            gossipSvc.spec.ports[0].port,
          ],
        ],
      },
    },

    manifests+:: {
      [gossipRingName]: gossipSvc,
    },
  },

  withVolumeClaimTemplate:: {
    local l = self,
    config+:: {
      volumeClaimTemplate: error 'must provide volumeClaimTemplate for ingesters and queriers',
    },
    manifests+:: {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          template+: {
            spec+: {
              volumes: std.filter(function(v) v.name != 'storage', super.volumes),
            },
          },
          volumeClaimTemplates: [l.config.volumeClaimTemplate {
            metadata+: {
              name: 'storage',
              labels+: l.config.podLabelSelector,
            },
          }],
        },
      }
      for name in std.objectFields(l.components)
      if isStatefulSet(name)
    },
  },

  withServiceMonitor:: {
    local l = self,
    serviceMonitors: {},

    manifests+:: {
      [normalizedName(name) + '-service-monitor']: {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata+: {
          name: l.config.name + '-' + name,
          namespace: l.config.namespace,
          labels: l.config.commonLabels,
        },
        spec: {
          selector: {
            matchLabels: l.config.podLabelSelector {
              'app.kubernetes.io/component': normalizedName(name),
            },
          },
          endpoints: [
            { port: 'metrics' },
          ],
        },
      } + if std.objectHas(l.serviceMonitors, name) then l.serviceMonitors[name] else {}
      for name in std.objectFields(loki.components)
      if std.member(['compactor', 'distributor', 'query_frontend', 'querier', 'ingester'], name)
    },
  },

  withResources:: {
    local l = self,
    config+:: {
      resources: error 'must provide resources per component',
    },

    manifests+:: {
      [normalizedName(name) + '-deployment']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  resources: l.config.resources[name],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(l.config.resources)
      if !isStatefulSet(name)
    } + {
      [normalizedName(name) + '-statefulset']+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  resources: l.config.resources[name],
                }
                for c in super.containers
              ],
            },
          },
        },
      }
      for name in std.objectFields(l.config.resources)
      if isStatefulSet(name)
    },
  },

  manifests+:: {
    'config-map': loki.configmap,
  } + {
    [normalizedName(name) + '-deployment']: newDeployment(name, loki.components[name])
    for name in std.objectFields(loki.components)
    if !isStatefulSet(name)
  } + {
    [normalizedName(name) + '-statefulset']: newStatefulSet(name, loki.components[name])
    for name in std.objectFields(loki.components)
    if isStatefulSet(name)
  } + {
    [normalizedName(name) + '-grpc-service']: newGrpcService(name)
    for name in std.objectFields(loki.components)
  } + {
    [normalizedName(name) + '-http-service']: newHttpService(name)
    for name in std.objectFields(loki.components)
    if std.member(['compactor', 'distributor', 'query_frontend', 'querier', 'ingester'], name)
  },
}
