local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';

{
  local api = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',

    metrics: {
      readEnpoint: error 'must provide metrics readEnpoint',
      writeEndpoint: error 'must provide metrics writeEndpoint',
    },

    ports: {
      public: 8080,
      internal: 8081,
    },

    logs: {},
    rbac: {},
    tenants: {},
    tls: {},

    commonLabels:: {
      'app.kubernetes.io/name': 'observatorium-api',
      'app.kubernetes.io/instance': api.config.name,
      'app.kubernetes.io/version': api.config.version,
      'app.kubernetes.io/component': 'api',
    },

    podLabelSelector:: {
      [labelName]: api.config.commonLabels[labelName]
      for labelName in std.objectFields(api.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  service:
    local service = k.core.v1.service;
    local ports = service.mixin.spec.portsType;

    service.new(
      api.config.name,
      api.config.podLabelSelector,
      [
        {
          name: name,
          port: api.config.ports[name],
          targetPort: api.config.ports[name],
        }
        for name in std.objectFields(api.config.ports)
      ],
    ) +
    service.mixin.metadata.withNamespace(api.config.namespace) +
    service.mixin.metadata.withLabels(api.config.commonLabels),

  deployment:
    local deployment = k.apps.v1.deployment;
    local container = deployment.mixin.spec.template.spec.containersType;
    local containerPort = container.portsType;

    local c =
      container.new('observatorium-api', api.config.image) +
      container.withArgs([
        '--web.listen=0.0.0.0:%s' % api.config.ports.public,
        '--web.internal.listen=0.0.0.0:%s' % api.config.ports.internal,
        '--metrics.read.endpoint=' + api.config.metrics.readEndpoint,
        '--metrics.write.endpoint=' + api.config.metrics.writeEndpoint,
        '--log.level=warn',
      ] + (
        if api.config.logs != {} then
          [
            '--logs.read.endpoint=' + api.config.logs.readEndpoint,
            '--logs.tail.endpoint=' + api.config.logs.tailEndpoint,
            '--logs.write.endpoint=' + api.config.logs.writeEndpoint,
          ]
        else []
      ) + (
        if api.config.rbac != {} then
          ['--rbac.config=/etc/observatorium/rbac.yaml']
        else []
      ) + (
        if api.config.tenants != {} then
          ['--tenants.config=/etc/observatorium/tenants.yaml']
        else []
      ) + (
        if api.config.tls != {} then
          [
            '--web.healthchecks.url=https://127.0.0.1:%s' % api.config.ports.public,
            '--tls.server.cert-file=/var/run/tls/' + api.config.tls.certKey,
            '--tls.server.key-file=/var/run/tls/' + api.config.tls.keyKey,
          ] + (
            if std.objectHas(api.config.tls, 'caKey') then [
              '--tls.healthchecks.server-ca-file=/var/run/tls/' + api.config.tls.caKey,
            ]
            else []
          ) + (
            if std.objectHas(api.config.tls, 'reloadInterval') then
              [
                '--tls.reload-interval=' + api.config.tls.reloadInterval,
              ]
            else []
          ) + (
            if std.objectHas(api.config.tls, 'serverName') then
              [
                '--tls.healthchecks.server-name=' + api.config.tls.serverName,
              ]
          )
        else []
      )) +
      container.withPorts([
        containerPort.newNamed(api.config.ports[name], name)
        for name in std.objectFields(api.config.ports)
      ]) +
      container.mixin.livenessProbe +
      container.mixin.livenessProbe.withPeriodSeconds(30) +
      container.mixin.livenessProbe.withFailureThreshold(10) +
      container.mixin.livenessProbe.httpGet.withPort(api.config.ports.internal) +
      container.mixin.livenessProbe.httpGet.withScheme('HTTP') +
      container.mixin.livenessProbe.httpGet.withPath('/live') +
      container.mixin.readinessProbe +
      container.mixin.readinessProbe.withPeriodSeconds(5) +
      container.mixin.readinessProbe.withFailureThreshold(12) +
      container.mixin.readinessProbe.httpGet.withPort(api.config.ports.internal) +
      container.mixin.readinessProbe.httpGet.withScheme('HTTP') +
      container.mixin.readinessProbe.httpGet.withPath('/ready') +
      container.withVolumeMounts(
        (if api.config.rbac != {} then [
           {
             name: 'rbac',
             mountPath: '/etc/observatorium/rbac.yaml',
             subPath: 'rbac.yaml',
             readOnly: true,
           },
         ] else []) +
        (if api.config.tenants != {} then [
           {
             name: 'tenants',
             mountPath: '/etc/observatorium/tenants.yaml',
             subPath: 'tenants.yaml',
             readOnly: true,
           },
         ] else []) +
        (if api.config.tls != {} then [
           {
             name: 'tls-secret',
             mountPath: '/var/run/tls/' + api.config.tls.certKey,
             subPath: api.config.tls.certKey,
             readOnly: true,
           },
           {
             name: 'tls-secret',
             mountPath: '/var/run/tls/' + api.config.tls.keyKey,
             subPath: api.config.tls.keyKey,
             readOnly: true,
           },
         ] + (
           if std.objectHas(api.config.tls, 'caKey') then [
             {
               name: 'tls-configmap',
               mountPath: '/var/run/tls/' + api.config.tls.caKey,
               subPath: api.config.tls.caKey,
               readOnly: true,
             },
           ] else []
         ) else [])
      );

    deployment.new(api.config.name, api.config.replicas, c, api.config.commonLabels) +
    deployment.mixin.metadata.withNamespace(api.config.namespace) +
    deployment.mixin.metadata.withLabels(api.config.commonLabels) +
    deployment.mixin.spec.selector.withMatchLabels(api.config.podLabelSelector) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxSurge(0) +
    deployment.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
    deployment.mixin.spec.template.spec.withVolumes(
      (if api.config.rbac != {} then [
         {
           configMap: {
             name: api.config.name,
           },
           name: 'rbac',
         },
       ] else []) +
      (if api.config.tenants != {} then [
         {
           secret: {
             secretName: api.config.name,
           },
           name: 'tenants',
         },
       ] else []) +
      (if api.config.tls != {} then [
         {
           secret: {
             secretName: api.config.tls.secretName,
           },
           name: 'tls-secret',
         },
       ] + (
         if std.objectHas(api.config.tls, 'caKey') then [
           {
             configMap: {
               name: api.config.tls.configMapName,
             },
             name: 'tls-configmap',
           },
         ] else []
       ) else [])
    ),

  configmap:
    if api.config.rbac != {} then {
      apiVersion: 'v1',
      data: {
        'rbac.yaml': std.manifestYamlDoc(api.config.rbac),
      },
      kind: 'ConfigMap',
      metadata: {
        labels: api.config.commonLabels,
        name: api.config.name,
        namespace: api.config.namespace,
      },
    } else null,

  secret:
    if api.config.tenants != {} then {
      apiVersion: 'v1',
      stringData: {
        'tenants.yaml': std.manifestYamlDoc(api.config.tenants),
      },
      kind: 'Secret',
      metadata: {
        labels: api.config.commonLabels,
        name: api.config.name,
        namespace: api.config.namespace,
      },
    } else null,

  withServiceMonitor:: {
    local api = self,

    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata+: {
        name: api.config.name,
        namespace: api.config.namespace,
      },
      spec: {
        selector: {
          matchLabels: api.config.commonLabels,
        },
        endpoints: [
          { port: 'internal' },
        ],
      },
    },
  },

  withResources:: {
    local api = self,

    config+:: {
      resources: error 'must provide resources',
    },

    deployment+: {
      spec+: {
        template+: {
          spec+: {
            containers: [
              if c.name == 'observatorium-api' then c {
                resources: api.config.resources,
              } else c
              for c in super.containers
            ],
          },
        },
      },
    },
  },
}
