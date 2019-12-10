local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'prometheus.jsonnet') +
  (import 'telemeter.jsonnet') +
  {
    local thanos = (import 'thanos.jsonnet'),

    template:
      list.asList('observatorium', {}, []) + {
        objects:
          [thanos.objects[name] for name in std.objectFields(thanos.objects)] +
          $.telemeterServer.list.objects +
          $.prometheusAms.template.objects,

        parameters:
          thanos.parameters +
          $.telemeterServer.list.parameters + [
            { name: 'TELEMETER_FORWARD_URL', value: '' },
          ] +
          $.prometheusAms.template.parameters,
      },
  } + {
    template+: {
      parameters:
        std.filter(function(param) !(param.name == 'NAMESPACE' && param.value == 'observatorium'), super.parameters),
    },
  };

// Output only the template
app.template
