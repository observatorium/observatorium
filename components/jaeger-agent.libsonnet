local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local deployment = k.apps.v1.deployment;
local container = deployment.mixin.spec.template.spec.containersType;
local containerEnv = container.envType;

{
  container::
    container.new('jaeger-agent', 'jaegertracing/jaeger-agent:1.14.0') +
    container.withArgs([
      '--reporter.grpc.host-port=dns:///jaeger-collector-headless.$(NAMESPACE).svc:14250',
      '--reporter.type=grpc',
      '--jaeger.tags=pod.namespace=$(NAMESPACE),pod.name=$(POD)',
    ]) +
    container.withEnv([
      containerEnv.fromFieldPath('NAMESPACE', 'metadata.namespace'),
      containerEnv.fromFieldPath('POD', 'metadata.name'),
    ]) +
    container.withPorts([
      container.portsType.newNamed(6831, 'jaeger-thrift'),
      container.portsType.newNamed(5778, 'configs'),
    ]),

  thanosFlag:: |||
    --tracing.config=
      type: JAEGER
      config:
        service_name: %s
        sampler_type: ratelimiting
        sampler_param: 2
  |||,
}
