local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local deployment = k.apps.v1.deployment;
local container = deployment.mixin.spec.template.spec.containersType;
local containerEnv = container.envType;

{
  local jm = self,
  config+:: {
    jaegerAgent: {
      image: error 'must provide image',
      collectorAddress: error 'must provide collectorAddress',
    },
  },

  specMixin:: {
    local sm = self,
    config+:: {
      jaegerAgent: {
        image: error 'must provide image',
        collectorAddress: error 'must provide collectorAddress',
      },
    },
    spec+: {
      template+: {
        metadata+: {
          labels+: {
            'app.kubernetes.io/tracing': 'jaeger-agent',
          },
        },
        spec+: {
          containers+: [
            container.new('jaeger-agent', sm.config.jaegerAgent.image) +
            container.withArgs([
              '--reporter.grpc.host-port=' + sm.config.jaegerAgent.collectorAddress,
              '--reporter.type=grpc',
              '--jaeger.tags=pod.namespace=$(NAMESPACE),pod.name=$(POD)',
            ]) +
            container.withEnv([
              containerEnv.fromFieldPath('NAMESPACE', 'metadata.namespace'),
              containerEnv.fromFieldPath('POD', 'metadata.name'),
            ]) +
            container.mixin.livenessProbe.withFailureThreshold(5) +
            container.mixin.livenessProbe.httpGet.withPath('/').withPort(14271).withScheme('HTTP') +
            container.mixin.resources.withRequests({ cpu: '32m', memory: '64Mi' }) +
            container.mixin.resources.withLimits({ cpu: '128m', memory: '128Mi' }) +
            container.withPorts([
              container.portsType.newNamed(6831, 'jaeger-thrift'),
              container.portsType.newNamed(5778, 'configs'),
              container.portsType.newNamed(14271, 'metrics'),
            ]),
          ],
        },
      },
    } + {
      template+: {
        spec+: {
          containers: [
            if std.startsWith(c.name, 'thanos-') then c {
              args+: [
                |||
                  --tracing.config=
                    type: JAEGER
                    config:
                      service_name: %s
                      sampler_type: ratelimiting
                      sampler_param: 2
                ||| % c.name,
              ],
            } else c
            for c in super.containers
          ],
        },
      },
    },
  },

  deploymentMixin:: {
    local dm = self,
    config+:: {
      jaegerAgent: {
        image: error 'must provide image',
        collectorAddress: error 'must provide collectorAddress',
      },
    },
    deployment+: jm.specMixin {
      config+:: {
        jaegerAgent+: {
          image: dm.config.jaegerAgent.image,
          collectorAddress: dm.config.jaegerAgent.collectorAddress,
        },
      },
    },
  },

  statefulSetMixin:: {
    local sm = self,
    config+:: {
      jaegerAgent: {
        image: error 'must provide image',
        collectorAddress: error 'must provide collectorAddress',
      },
    },
    statefulSet+: jm.specMixin {
      config+:: {
        jaegerAgent+: {
          image: sm.config.jaegerAgent.image,
          collectorAddress: sm.config.jaegerAgent.collectorAddress,
        },
      },
    },
  },
}
