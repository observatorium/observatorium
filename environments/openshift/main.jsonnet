local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'kube-thanos.libsonnet') +
  (import 'prometheus.libsonnet') +
  (import 'telemeter.libsonnet') +
  {
    local thanos = super.thanos,

    template:
      list.asList('observatorium', {}, []) + {
        objects:
          $.thanos.template.objects +
          $.telemeterServer.list.objects +
          $.prometheusAms.template.objects,

        parameters:
          $.thanos.template.parameters +
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
