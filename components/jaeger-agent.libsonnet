local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local deployment = k.apps.v1.deployment;
local container = deployment.mixin.spec.template.spec.containersType;
local containerEnv = container.envType;

{
  container(image)::
    container.new('jaeger-agent', image) +
    container.withArgs([
      '--reporter.grpc.host-port=dns:///jaeger-collector-headless.$(NAMESPACE).svc:14250',
      '--reporter.type=grpc',
      '--jaeger.tags=pod.namespace=$(NAMESPACE),pod.name=$(POD)',
    ]) +
    container.withEnv([
      containerEnv.fromFieldPath('NAMESPACE', 'metadata.namespace'),
      containerEnv.fromFieldPath('POD', 'metadata.name'),
    ]) +
    container.mixin.resources.withRequests({ cpu: '32m', memory: '16Mi' }) +
    container.mixin.resources.withLimits({ cpu: '128m', memory: '64Mi' }) +
    container.mixin.livenessProbe.withFailureThreshold(5) +
    container.mixin.livenessProbe.httpGet.withPath('/').withPort(self.metricsPort).withScheme('HTTP') +
    container.withPorts([
      container.portsType.newNamed(6831, 'jaeger-thrift'),
      container.portsType.newNamed(5778, 'configs'),
      container.portsType.newNamed(self.metricsPort, 'metrics'),
    ]),

  metricsPort:: 14271,
  serviceLabels:: {
    'app.kubernetes.io/name': 'jaeger-agent',
  },
  labels:: {
    'app.kubernetes.io/tracing': 'jaeger-agent',
  },
  thanosFlag:: |||
    --tracing.config=
      type: JAEGER
      config:
        service_name: %s
        sampler_type: ratelimiting
        sampler_param: 2
  |||,
}
