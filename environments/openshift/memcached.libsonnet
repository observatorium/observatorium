local memcached = import '../kubernetes/memcached.libsonnet';
local list = import 'telemeter/lib/list.libsonnet';

{
  memcached+:: {
    list: list.asList('memcached', memcached.memcached, [
            {
              name: 'MEMCACHED_IMAGE',
              value: memcached.memcached.image,
            },
          ]) +
          list.withNamespace($._config),
  },
}
