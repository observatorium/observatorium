local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'kube-thanos.libsonnet') +
  {
    // This generates the Template kind that OpenShift requires
    local thanos = super.thanos,

    template:
      local objects = {
        ['querier-' + name]: thanos.querier[name]
        for name in std.objectFields(thanos.querier)
      } + {
        ['store-' + name]: thanos.store[name]
        for name in std.objectFields(thanos.store)
      } + {
        ['receive-' + name]: thanos.receive[name]
        for name in std.objectFields(thanos.receive)
      };

      list.asList('observatorium', objects, [
        {
          name: 'NAMESPACE',
          value: 'telemeter',
        },
        {
          name: 'IMAGE',
          value: 'improbable/thanos:v0.5.0',
        },
        {
          name: 'IMAGE_TAG',
          value: 'dummy',  // We don't actually use this, but need it for OpenShift.
        },
        {
          name: 'THANOS_QUERIER_REPLICAS',
          value: '3',
        },
        {
          name: 'THANOS_STORE_REPLICAS',
          value: '5',
        },
        {
          name: 'THANOS_RECEIVE_REPLICAS',
          value: '5',
        },
        {
          name: 'THANOS_CONFIG_SECRET',
          value: 'thanos-objectstorage',
        },
        {
          name: 'THANOS_S3_SECRET',
          value: 'telemeter-thanos-stage-s3',
        },
      ]),
  };

// Output only the template
app.template
