local config = import 'operator-config.libsonnet';
local manifests = ((import '../../components/observatorium.libsonnet') + {
                     config+:: config,
                   } + (import '../../components/observatorium-configure.libsonnet')).manifests;

{
  manifests: std.mapWithKey(function(k, v) v {
    metadata+: {
      ownerReferences: [{
        apiVersion: config.apiVersion,
        blockOwnerdeletion: true,
        controller: true,
        kind: config.kind,
        name: config.name,
        uid: config.uid,
      }],
    },
  }, manifests),
}
