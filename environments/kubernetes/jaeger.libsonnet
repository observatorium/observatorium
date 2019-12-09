(import '../../components/jaeger-collector.libsonnet') + {
  jaeger+:: {
    namespace:: 'observatorium',
    image:: 'jaegertracing/all-in-one:1.15.0',
  },
}
