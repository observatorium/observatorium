local app =
  (import '../kubernetes/jaeger.libsonnet') + {
    jaeger+:: {
      namespace:: '${NAMESPACE}',
      image:: '${IMAGE}:${IMAGE_TAG}',
      replicas:: '${{REPLICAS}}',
    },
  };

{
  apiVersion: 'v1',
  kind: 'Template',
  metadata: {
    name: 'jaeger',
  },
  objects: [
    app.jaeger[name]
    for name in std.objectFields(app.jaeger)
  ],
  parameters: [
    { name: 'NAMESPACE', value: 'telemeter' },
    { name: 'IMAGE', value: 'jaegertracing' },
    { name: 'IMAGE_TAG', value: 'all-in-one:1.14.0' },
    { name: 'REPLICAS', value: 1 },
  ],
}
