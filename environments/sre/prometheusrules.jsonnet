local thanos = (import 'thanos-mixin/mixin.libsonnet') + (import 'thanos-mixin/defaults.libsonnet');
local thanosReceiveController = (import 'thanos-receive-controller-mixin/mixin.libsonnet');
local jaeger = (import 'jaeger-mixin/mixin.libsonnet');
local slo = import 'slo-libsonnet/slo.libsonnet';
local observatoriumSLOs = import '../../slos.libsonnet';
local tenants = import '../../tenants.libsonnet';

local capitalize(str) = std.asciiUpper(std.substr(str, 0, 1)) + std.asciiLower(std.substr(str, 1, std.length(str)));

{
  'observatorium-thanos-stage.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-stage',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos + thanosReceiveController {
      compactor+:: {
        jobPrefix: 'thanos-compactor',
        selector: 'job="%s", namespace="telemeter-stage"' % self.jobPrefix,
      },
      querier+:: {
        jobPrefix: 'thanos-querier',
        selector: 'job=~"%s.*", namespace="telemeter-stage"' % self.jobPrefix,
      },
      receiver+:: {
        jobPrefix: 'thanos-receive',
        selector: 'job=~"%s-.*", namespace="telemeter-stage"' % self.jobPrefix,
      },
      store+:: {
        jobPrefix: 'thanos-store',
        selector: 'job="%s", namespace="telemeter-stage"' % self.jobPrefix,
      },
      ruler+:: {
        jobPrefix: 'thanos-ruler',
        selector: 'job=~"%s.*", namespace="telemeter-stage"' % self.jobPrefix,
      },

      // We build alerts for the presence of all these jobs.
      jobs+: {
        ['ThanosReceive' + capitalize(tenant.hashring)]: 'job="thanos-receive-%s", namespace="telemeter-stage"' % tenant.hashring
        for tenant in tenants
      },

      _config+:: {
        thanosReceiveControllerJobPrefix: 'thanos-receive-controller',
        thanosReceiveControllerSelector: 'job=~"%s.*",namespace="telemeter-stage"' % self.thanosReceiveControllerJobPrefix,
      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },

    spec: alerts.prometheusAlerts,
  },
  'observatorium-thanos-production.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-production',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos + thanosReceiveController {
      compactor+:: {
        jobPrefix: 'thanos-compactor',
        selector: 'job="%s", namespace="telemeter-production"' % self.jobPrefix,
      },
      querier+:: {
        jobPrefix: 'thanos-querier',
        selector: 'job=~"%s.*", namespace="telemeter-production"' % self.jobPrefix,
      },
      receiver+:: {
        jobPrefix: 'thanos-receive',
        selector: 'job=~"%s-.*", namespace="telemeter-production"' % self.jobPrefix,
      },
      store+:: {
        jobPrefix: 'thanos-store',
        selector: 'job="%s", namespace="telemeter-production"' % self.jobPrefix,
      },
      ruler+:: {
        jobPrefix: 'thanos-ruler',
        selector: 'job=~"%s.*", namespace="telemeter-production"' % self.jobPrefix,
      },

      // We build alerts for the presence of all these jobs.
      jobs+: {
        ['ThanosReceive' + capitalize(tenant.hashring)]: 'job="thanos-receive-%s", namespace="telemeter-production"' % tenant.hashring
        for tenant in tenants
      },

      _config+:: {
        thanosReceiveControllerJobPrefix: 'thanos-receive-controller',
        thanosReceiveControllerSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosReceiveControllerJobPrefix,
      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },

    spec: alerts.prometheusAlerts,
  },
}
