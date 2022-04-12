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

  commonLabels:: {
    'app.kubernetes.io/name': 'otelcol',
    'app.kubernetes.io/part-of': 'observatorium',
    'app.kubernetes.io/instance': defaults.name,
  },
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
      ['jaeger/' + tenant]: { endpoint: normalizedName(tracing.config.name + '-jaeger-' + tenant + '-collector:14250'), tls: { insecure: true } }
      for tenant in tracing.config.tenants
    },
    processors: {
      routing: {
        from_attribute: 'X-Tenant',
        table: [
          { value: tenant, exporters: ['jaeger/' + tenant] }
          for tenant in tracing.config.tenants
        ],
      },
    },
    service: {
      pipelines: {
        traces: {
          receivers: ['otlp'],
          processors: ['routing'],
          exporters: ['jaeger/' + tenant for tenant in tracing.config.tenants],
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

  local normalizedName(id) =
    std.strReplace(id, '_', '-'),

  local newCommonLabels(component) =
    tracing.config.commonLabels {
      'app.kubernetes.io/component': normalizedName(component),
    },

  local newJaeger(component, config) =
    local name = normalizedName(tracing.config.name + '-jaeger-' + component);
    {
      apiVersion: 'jaegertracing.io/v1',
      kind: 'Jaeger',
      metadata: {
        name: name,
        namespace: tracing.config.namespace,
        labels: newCommonLabels(component),
      },
      spec: {
        strategy: tracing.config.jaeger.strategy,
        ui: tracing.config.jaeger.ui,
      },
    },
  manifests: {
    otelcollector: tracing.otelcolcr,
  } + {
    [normalizedName('jaeger-' + tenant)]: newJaeger(tenant, tracing.config.components[tenant])
    for tenant in tracing.config.tenants
  },
}
