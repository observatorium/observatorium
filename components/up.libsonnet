{
  local up = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    endpointType: error 'must provide endpoint type',
    replicas: 1,
    queryConfig: {},
    readEndpoint: '',
    writeEndpoint: '',
    logs: '',
    resources: {
      requests: {},
      limits: {},
    },
    serviceMonitor: false,

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-up',
      'app.kubernetes.io/instance': up.config.name,
      'app.kubernetes.io/version': up.config.version,
      'app.kubernetes.io/component': 'blackbox-prober',
    },

    podLabelSelector:: {
      [labelName]: up.config.commonLabels[labelName]
      for labelName in std.objectFields(up.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: up.config.name,
      namespace: up.config.namespace,
      labels: up.config.commonLabels,
    },
    spec: {
      ports: [
        { name: 'http', targetPort: 8080, port: 8080 },
      ],
      selector: up.config.podLabelSelector,
    },
  },

  deployment:
    local c = {
      name: 'observatorium-up',
      image: up.config.image,
      args: [
        '--duration=0',
        '--log.level=debug',
        '--endpoint-type=' + up.config.endpointType,
      ] + (
        if up.config.queryConfig != {} then
          ['--queries-file=/etc/up/queries.yaml']
        else []
      ) + (
        if up.config.readEndpoint != '' then
          ['--endpoint-read=' + up.config.readEndpoint]
        else []
      ) + (
        if up.config.writeEndpoint != '' then
          ['--endpoint-write=' + up.config.writeEndpoint]
        else []
      ) + (
        if up.config.logs != '' then
          ['--logs=' + up.config.logs]
        else []
      ),
      ports: [
        { name: port.name, containerPort: port.port }
        for port in up.service.spec.ports
      ],
      volumeMounts: if up.config.queryConfig != {} then [{
        mountPath: '/etc/up/',
        name: 'query-config',
        readOnly: false,
      }] else [],
      resources: up.config.resources,
    };

    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: up.config.name,
        namespace: up.config.namespace,
        labels: up.config.commonLabels,
      },
      spec: {
        replicas: up.config.replicas,
        selector: { matchLabels: up.config.podLabelSelector },
        template: {
          metadata: {
            labels: up.config.commonLabels,
          },
          spec: {
            containers: [c],
            volumes: if up.config.queryConfig != {} then
              [
                {
                  configMap: {
                    name: up.config.name,
                  },
                  name: 'query-config',
                },
              ] else [],
          },
        },
      },
    },

  configmap:
    if up.config.queryConfig != {} then {
      apiVersion: 'v1',
      data: {
        'queries.yaml': std.manifestYamlDoc(up.config.queryConfig),
      },
      kind: 'ConfigMap',
      metadata: {
        labels: up.config.commonLabels,
        name: up.config.name,
        namespace: up.config.namespace,
      },
    } else null,

  serviceMonitor:
    if up.config.serviceMonitor == true then
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata+: {
          name: up.config.name,
          namespace: up.config.namespace,
        },
        spec: {
          selector: {
            matchLabels: up.config.podLabelSelector,
          },
          endpoints: [
            { port: 'http' },
          ],
        },
      } else null,

  manifests+:: {
    ['up-' + name]: up[name]
    for name in std.objectFields(up)
    if up[name] != null
  },
}
