(import '../../components/observatorium.libsonnet') + {
  config+:: (import 'default-config.libsonnet'),
} + (import '../../components/observatorium-configure.libsonnet')
