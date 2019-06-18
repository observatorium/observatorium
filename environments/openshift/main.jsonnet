local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'kube-thanos.libsonnet') +
  {
    _config+:: {
      namespace: '{$NAMESPACE}',

      images+: {
        thanos: '$IMAGE',
      },

      thanos+: {
        objectStorageConfig: {
          name: 'telemeter-thanos-config',
          key: 'thanos.yaml',
        },
      },
    },

    // This generates the Template kind that OpenShift requires

    local t = super.thanos,
    thanos+:: {
      template:
        local objects = {
          ['querier-' + name]: t.querier[name]
          for name in std.objectFields(t.querier)
        } + {
          ['store-' + name]: t.store[name]
          for name in std.objectFields(t.store)
        };

        list.asList('thanos', objects, [
          {
            name: 'NAMESPACE',
            value: 'telemeter',
          },
          {
            name: 'IMAGE',
            value: 'improbable/thanos:v0.5.0',
          },
          {
            name: 'THANOS_QUERIER_REPLICAS',
            value: 3,
          },
          {
            name: 'THANOS_STORE_REPLICAS',
            value: 5,
          },
        ]),
    },
  };

// Output only the template
app.thanos.template
