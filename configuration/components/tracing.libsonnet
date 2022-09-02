// These are the defaults for this components configuration.
// When calling the function to generate the component's manifest,
// you can pass an object structured like the default to overwrite default values.
local defaults = {
  local defaults = self,

  name: 'observatorium-xyz',
  namespace: error 'must provide namespace',
  tenants: [],
  enabled: false,
  otelcolVersion: '0.46.0',
  // The core distribution does not contain routing processor, therefore using contrib distribution
  // https://github.com/orgs/open-telemetry/packages?repo_name=opentelemetry-collector-releases
  otelcolImage: 'ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib',
  otelcolTLS: {
    insecure: true,
  },

  commonLabels:: {
    'app.kubernetes.io/name': 'otelcol',
    'app.kubernetes.io/part-of': 'observatorium',
    'app.kubernetes.io/instance': defaults.name,
  },
  monitoring: false,
};

function(params) {
  local tracing = self,

  // Combine the defaults and the passed params to make the component's config.
  config:: defaults + params,

  otelcolcfg:: {
    receivers: {
      otlp: {
        protocols: {
          grpc: {},
        },
      },
    },
    exporters: {
      ['jaeger/' + tenantName]: { endpoint: normalizedName(tracing.config.name + '-jaeger-' + tenantName + '-collector-headless.' + tracing.config.namespace + '.svc.cluster.local:14250'), tls: tracing.config.otelcolTLS }
      for tenantName in tracing.config.tenants
    },
    processors: {
      routing: {
        from_attribute: 'X-Tenant',
        table: [
          { value: tenantName, exporters: ['jaeger/' + tenantName] }
          for tenantName in tracing.config.tenants
        ],
      },
    },
    service: {
      telemetry: {
        metrics: {
          level: 'basic',
        },
      },
      pipelines: {
        traces: {
          receivers: ['otlp'],
          processors: ['routing'],
          exporters: ['jaeger/' + tenantName for tenantName in tracing.config.tenants],
        },
      },
    },
  },

  otelcolcr:: {
    apiVersion: 'opentelemetry.io/v1alpha1',
    kind: 'OpenTelemetryCollector',
    metadata: {
      name: normalizedName(tracing.config.name + '-otel'),
      namespace: tracing.config.namespace,
      labels: newCommonLabels('jaeger'),
    },
    spec: {
      image: tracing.config.otelcolImage + ':' + tracing.config.otelcolVersion,
      mode: 'deployment',
      config: std.manifestYamlDoc(tracing.otelcolcfg, indent_array_in_object=false, quote_keys=false),
    },
  },

  otelServiceMonitor: {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    metadata+: {
      name: 'otel-' + tracing.config.name,
      namespace: tracing.config.namespace,
    },
    spec: {
      selector: {
        matchLabels: {
          'app.kubernetes.io/instance': tracing.config.namespace + '.' + normalizedName(tracing.config.name + '-otel'),
          'app.kubernetes.io/component': 'opentelemetry-collector',
          'app.kubernetes.io/name': normalizedName(tracing.config.name + '-otel') + '-collector-monitoring',
        },
      },
      endpoints: [
        { port: 'monitoring' },
      ],
    },
  },

  local newPodMonitor(tenantName) =
    local name = normalizedName(tracing.config.name + '-jaeger-' + tenantName);
    {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'PodMonitor',
      metadata: {
        name: name + '-admin',
        namespace: tracing.config.namespace,
        labels: newCommonLabels(tenantName),
      },
      spec: {
        namespaceSelector: {},
        podMetricsEndpoints: [{
          port: 'admin-http',
          interval: '2s',
        }],
        selector: {
          matchLabels: {
            'app.kubernetes.io/instance': name,
            app: 'jaeger',
          },
        },
      },
    },

  local normalizedName(id) =
    std.strReplace(id, '_', '-'),

  local newCommonLabels(tenantName) =
    tracing.config.commonLabels {
      'app.kubernetes.io/component': normalizedName(tenantName),
    },

  local newJaeger(tenantName, config) =
    local name = normalizedName(tracing.config.name + '-jaeger-' + tenantName);
    {
      apiVersion: 'jaegertracing.io/v1',
      kind: 'Jaeger',
      metadata: {
        name: name,
        namespace: tracing.config.namespace,
        labels: newCommonLabels(tenantName),
      },
      spec: tracing.config.jaegerSpec,
    } + {
      spec+: (if tracing.config.jaegerSpec.strategy == 'production' &&
                 tracing.config.jaegerSpec.storage.type == 'elasticsearch' then {
                storage+: {
                  options+: {
                    es+: {
                      'index-prefix'+: tenantName,
                    },
                  },
                },
              } else {}),
    },
  manifests: {
    otelcollector: tracing.otelcolcr,
  } + {
    [normalizedName('jaeger-' + tenantName)]: newJaeger(tenantName, tracing.config.components[tenantName])
    for tenantName in tracing.config.tenants
  } + (
    if tracing.config.monitoring == true then {
      [normalizedName('jaeger-podmonitor-' + tenantName)]: newPodMonitor(tenantName)
      for tenantName in tracing.config.tenants
    } else {}
  ) + (if tracing.config.monitoring == true then {
         'servicemonitor-otel': tracing.otelServiceMonitor,
       } else {}),
}
