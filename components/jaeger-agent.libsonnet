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
            {
              name: 'jaeger-agent',
              image: sm.config.jaegerAgent.image,
              args: [
                '--reporter.grpc.host-port=' + sm.config.jaegerAgent.collectorAddress,
                '--reporter.type=grpc',
                '--jaeger.tags=pod.namespace=$(NAMESPACE),pod.name=$(POD)',
              ],
              env: [
                {
                  name: 'NAMESPACE',
                  valueFrom: { fieldRef: { fieldPath: 'metadata.namespace' } },
                },
                {
                  name: 'POD',
                  valueFrom: { fieldRef: { fieldPath: 'metadata.name' } },
                },
              ],
              ports: [
                { name: 'jaeger-thrift', containerPort: 6831 },
                { name: 'configs', containerPort: 5778 },
                { name: 'metrics', containerPort: 14271 },
              ],
              readinessProbe: {
                failureThreshold: 5,
                httpGet: {
                  scheme: 'HTTP',
                  port: 14271,
                  path: '/',
                },
              },
              resources: {
                requests: { cpu: '32m', memory: '64Mi' },
                limits: { cpu: '128m', memory: '128Mi' },
              },
            },
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
