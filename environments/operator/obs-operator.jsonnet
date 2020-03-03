(import '../../components/observatorium.libsonnet') + {
  config+:: (import 'operator-config.libsonnet'),
} + (import '../../components/observatorium-configure.libsonnet')
