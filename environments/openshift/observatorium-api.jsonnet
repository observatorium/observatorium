local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'observatorium.libsonnet') + {
    apiVersion: 'v1',
    kind: 'Template',
    metadata: {
      name: 'observatorium-api',
    },
    objects: [
      $.observatorium.api[name]
      for name in std.objectFields($.observatorium.api)
    ] + [
      $.observatorium.querier[name]
      for name in std.objectFields($.observatorium.querier)
    ],
    parameters: [
      { name: 'NAMESPACE', value: 'telemeter' },
      { name: 'IMAGE', value: '' },
      { name: 'IMAGE_TAG', value: '' },
      { name: 'OBSERVATORIUM_API_IMAGE', value: 'quay.io/observatorium/observatorium' },
      { name: 'OBSERVATORIUM_API_IMAGE_TAG', value: 'master-2020-01-14-d076eab' },
      { name: 'PROXY_IMAGE', value: 'openshift/oauth-proxy' },
      { name: 'PROXY_IMAGE_TAG', value: 'v1.1.0' },
      { name: 'OBSERVATORIUM_API_REPLICAS', value: '3' },
      { name: 'OBSERVATORIUM_API_CPU_REQUEST', value: '100m' },
      { name: 'OBSERVATORIUM_API_CPU_LIMIT', value: '1' },
      { name: 'OBSERVATORIUM_API_MEMORY_REQUEST', value: '256Mi' },
      { name: 'OBSERVATORIUM_API_MEMORY_LIMIT', value: '1Gi' },
      { name: 'OBSERVATORIUM_API_PROXY_CPU_REQUEST', value: '100m' },
      { name: 'OBSERVATORIUM_API_PROXY_MEMORY_REQUEST', value: '100Mi' },
      { name: 'OBSERVATORIUM_API_PROXY_CPU_LIMITS', value: '200m' },
      { name: 'OBSERVATORIUM_API_PROXY_MEMORY_LIMITS', value: '200Mi' },
      { name: 'OBSERVATORIUM_API_EXTERNAL_URL', value: 'https://observatorium.api' },
      { name: 'OBSERVATORIUM_API_THANOS_QUERIER_CPU_REQUEST', value: '100m' },
      { name: 'OBSERVATORIUM_API_THANOS_QUERIER_CPU_LIMIT', value: '1' },
      { name: 'OBSERVATORIUM_API_THANOS_QUERIER_MEMORY_REQUEST', value: '256Mi' },
      { name: 'OBSERVATORIUM_API_THANOS_QUERIER_MEMORY_LIMIT', value: '1Gi' },
    ],
    template:
      list.asList('observatorium-api', {}, []) + {
        objects: $.objects,
        parameters: $.parameters,
      },
  } + {
    template+: {
      parameters:
        std.filter(function(param) !(param.name == 'NAMESPACE' && param.value == 'observatorium'), super.parameters),
    },
  };

// Output only the template
app.template
