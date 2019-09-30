local k3 = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local list = import 'telemeter/lib/list.libsonnet';

{
  _config+:: {
    namespace: 'observatorium',

    versions+:: {
      prometheusAms: 'v2.12.0',
      remoteWriteProxy: '14e844d',
    },

    imageRepos+:: {
      prometheusAms: 'quay.io/prometheus/prometheus',
      remoteWriteProxy: 'quay.io/app-sre/observatorium-receive-proxy',
    },

    ams+:: {
      proxyPort: 8080,
      dnsResolver: 'kube-dns.kube-system.svc.cluster.local',
      remoteWriteTarget: 'http://%s.%s.svc.cluster.local:%d/api/v1/receive' % [
        $.thanos.receive.service.metadata.name,
        $.thanos.receive.service.metadata.namespace,
        $.thanos.receive.service.spec.ports[1].port,
      ],
      receiveTenantId: 'FB870BF3-9F3A-44FF-9BF7-D7A047A52F43',

      prometheus+:: {
        namespaces: [$._config.namespace],
        name: 'ams',
        replicas: 1,
        rules: {},
        renderedRules: {},
        resourceLimits: {},
        resourceRequests: { memory: '400Mi' },
        remoteWrite: [
          {
            url: 'http://localhost:%d' % $._config.ams.proxyPort,
            writeRelabelConfigs: {
              sourceLabels: ['__name__'],
              regex: 'subscription_labels',
              action: 'keep',
            },
          },
        ],
      },
    },
  },

  prometheusAms+:: {
    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      local prometheusPort = servicePort.newNamed('web', 9090, 'web');

      service.new('prometheus-' + $._config.ams.prometheus.name, { app: 'prometheus', prometheus: $._config.ams.prometheus.name }, prometheusPort) +
      service.mixin.spec.withSessionAffinity('ClientIP') +
      service.mixin.metadata.withNamespace($._config.namespace) +
      service.mixin.metadata.withLabels({ prometheus: $._config.ams.prometheus.name }),
    [if $._config.ams.prometheus.rules != null && $._config.ams.prometheus.rules != {} then 'rules']:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'PrometheusRule',
        metadata: {
          labels: {
            prometheus: $._config.ams.prometheus.name,
            role: 'alert-rules',
          },
          name: 'prometheus-' + $._config.ams.prometheus.name + '-rules',
          namespace: $._config.namespace,
        },
        spec: {
          groups: $._config.ams.prometheus.rules.groups,
        },
      },
    prometheus:
      local statefulSet = k.apps.v1.statefulSet;
      local container = statefulSet.mixin.spec.template.spec.containersType;
      local resourceRequirements = container.mixin.resourcesType;
      local selector = statefulSet.mixin.spec.selectorType;
      local deployment = k.apps.v1.deployment;
      local container = deployment.mixin.spec.template.spec.containersType;
      local volume = k.apps.v1beta2.statefulSet.mixin.spec.template.spec.volumesType;
      local volumeMount = container.volumeMountsType;

      local resources =
        resourceRequirements.new() +
        resourceRequirements.withLimits({
          cpu: '${PROMETHEUS_AMS_CPU_LIMIT}',
          memory: '${PROMETHEUS_AMS_MEMORY_LIMIT}',
        }) +
        resourceRequirements.withRequests({
          cpu: '${PROMETHEUS_AMS_CPU_REQUEST}',
          memory: '${PROMETHEUS_AMS_MEMORY_REQUEST}',
        });

      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'Prometheus',
        metadata: {
          name: $._config.ams.prometheus.name,
          namespace: $._config.namespace,
          labels: {
            prometheus: $._config.ams.prometheus.name,
          },
        },
        spec: {
          replicas: $._config.ams.prometheus.replicas,
          version: '${PROMETHEUS_AMS_IMAGE_TAG}',
          baseImage: '${PROMETHEUS_AMS_IMAGE}',
          serviceAccount: 'prometheus-telemeter',
          serviceAccountName: 'prometheus-telemeter',
          serviceMonitorSelector: {},
          serviceMonitorNamespaceSelector: {},
          ruleSelector: selector.withMatchLabels({
            role: 'alert-rules',
            prometheus: $._config.ams.prometheus.name,
          }),
          resources: resources,
          securityContext: {
            runAsUser: 1000,
            runAsNonRoot: true,
            fsGroup: 2000,
          },
          remoteWrite: $._config.ams.prometheus.remoteWrite,
          containers: [
            container.new('remote-write-proxy', '${PROMETHEUS_AMS_REMOTE_WRITE_PROXY_IMAGE}:${PROMETHEUS_AMS_REMOTE_WRITE_PROXY_VERSION}') +
            container.withArgs([]) +
            container.withPorts([{ name: 'http', containerPort: $._config.ams.proxyPort }]) +
            container.withVolumeMounts([volumeMount.new('prometheus-%s' % $.prometheusAms.configmap.metadata.name, '/nginx.conf', true)]),
          ],
          volumes+: [
            volume.fromConfigMap('prometheus-%s' % $.prometheusAms.configmap.metadata.name, $.prometheusAms.configmap.metadata.name, ['nginx.conf']),
          ],
        },
      },
    serviceMonitor:
      {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus-' + $._config.ams.prometheus.name,
          namespace: $._config.namespace,
          labels: {
            'k8s-app': 'prometheus',
          },
        },
        spec: {
          selector: {
            matchLabels: {
              prometheus: $._config.ams.prometheus.name,
            },
          },
          endpoints: [
            {
              port: 'web',
              interval: '30s',
            },
          ],
        },
      },
    configmap:
      local configmap = k.core.v1.configMap;

      configmap.new() +
      configmap.mixin.metadata.withName('remote-write-proxy-config') +
      configmap.mixin.metadata.withNamespace($._config.namespace) +
      configmap.mixin.metadata.withLabels({ prometheus: $._config.ams.prometheus.name }) +
      configmap.withData({
        local f = importstr './prometheus/remote_write_proxy.conf',

        'nginx.conf': std.format(f, {
          listen_port: $._config.ams.proxyPort,
          dns_resolver: $._config.ams.dnsResolver,
          forward_host: $._config.ams.remoteWriteTarget,
          thanos_tenant: $._config.ams.receiveTenantId,
        }),
      }),
  },
} + {
  local prom = super.prometheusAms,
  prometheusAms+:: {
    template+:
      list.asList('prometheus-observatorium-ams', prom, [
        {
          name: 'PROMETHEUS_AMS_REMOTE_WRITE_PROXY_IMAGE',
          value: $._config.imageRepos.remoteWriteProxy,
        },
        {
          name: 'PROMETHEUS_AMS_REMOTE_WRITE_PROXY_VERSION',
          value: $._config.versions.remoteWriteProxy,
        },
        {
          name: 'PROMETHEUS_AMS_IMAGE',
          value: $._config.imageRepos.prometheusAms,
        },
        {
          name: 'PROMETHEUS_AMS_IMAGE_TAG',
          value: $._config.versions.prometheusAms,
        },
        {
          name: 'PROMETHEUS_AMS_CPU_REQUEST',
          value: if std.objectHas($._config.ams.prometheus.resourceRequests, 'cpu') then $._config.ams.prometheus.resourceRequests.cpu else '0',
        },
        {
          name: 'PROMETHEUS_AMS_CPU_LIMIT',
          value: if std.objectHas($._config.ams.prometheus.resourceLimits, 'cpu') then $._config.ams.prometheus.resourceLimits.cpu else '0',
        },
        {
          name: 'PROMETHEUS_AMS_MEMORY_REQUEST',
          value: if std.objectHas($._config.ams.prometheus.resourceRequests, 'memory') then $._config.ams.prometheus.resourceRequests.memory else '0',
        },
        {
          name: 'PROMETHEUS_AMS_MEMORY_LIMIT',
          value: if std.objectHas($._config.ams.prometheus.resourceLimits, 'memory') then $._config.ams.prometheus.resourceLimits.memory else '0',
        },
      ]) +
      list.withNamespace($._config),
  } + {
    local setNamespace(object) =
      if std.objectHas(object, 'metadata') && std.objectHas(object.metadata, 'namespace') then {
        metadata+: {
          namespace: '${NAMESPACE}',
        },
      },
    local setSubjectNamespace(object) =
      if std.endsWith(object.kind, 'Binding') then {
        subjects: [
          s { namespace: '${NAMESPACE}' }
          for s in super.subjects
        ],
      }
      else {},
    template+: {
      objects: [
        if std.objectHas(o, 'items') then o {
          items: [i + setNamespace(i) + setSubjectNamespace(i) for i in super.items],
        } else o
        for o in super.objects
      ],
    },
  },
}
