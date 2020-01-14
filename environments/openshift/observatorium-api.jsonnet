local list = import 'telemeter/lib/list.libsonnet';

// This entire file takes what we have for Kubernetes and
// generates an OpenShift specific Template on top of that.

local app =
  (import 'observatorium.libsonnet') + {
    template:
      list.asList('observatorium-api', {}, []) + {
        objects: app.objects,
        parameters: app.parameters,
      },
  } + {
    template+: {
      parameters:
        std.filter(function(param) !(param.name == 'NAMESPACE' && param.value == 'observatorium'), super.parameters),
    },
  };

// Output only the template
app.template
