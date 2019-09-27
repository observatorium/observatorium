local k3 = import 'ksonnet/ksonnet.beta.3/k.libsonnet';
local k = import 'ksonnet/ksonnet.beta.4/k.libsonnet';
local list = import 'telemeter/lib/list.libsonnet';

{
  _config+:: {
    local namespace = '${NAMESPACE}',
    namespace: namespace,

    versions+:: {
      prometheus: 'v2.12.0',
      remoteWriteProxy: 'latest',
    },

    imageRepos+:: {
      prometheus: 'quay.io/prometheus/prometheus',
      remoteWriteProxy: 'quay.io/app-sre/observatorium-receive-proxy',
    },

    ams+:: {
      proxyPort: 8080,
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
            write_relabel_configs: {
              source_labels: ['__name__'],
              regex: 'subscription_labels',
              action: 'keep',
            },
          },
        ],
      },
    },
  },

  prometheusAms+:: {
    serviceAccount:
      local serviceAccount = k.core.v1.serviceAccount;

      serviceAccount.new('prometheus-' + $._config.ams.prometheus.name) +
      serviceAccount.mixin.metadata.withNamespace($._config.namespace),
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
    roleBindingSpecificNamespaces:
      local roleBinding = k.rbac.v1.roleBinding;

      local newSpecificRoleBinding(namespace) =
        roleBinding.new() +
        roleBinding.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name) +
        roleBinding.mixin.metadata.withNamespace(namespace) +
        roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
        roleBinding.mixin.roleRef.withName('prometheus-' + $._config.ams.prometheus.name) +
        roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
        roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.ams.prometheus.name, namespace: $._config.namespace }]);

      local roleBindingList = k3.rbac.v1.roleBindingList;
      roleBindingList.new([newSpecificRoleBinding(x) for x in $._config.ams.prometheus.namespaces]),
    clusterRole:
      local clusterRole = k.rbac.v1.clusterRole;
      local policyRule = clusterRole.rulesType;

      local metricsRule = policyRule.new() +
                          policyRule.withNonResourceUrls('/metrics') +
                          policyRule.withVerbs(['get']);

      local rules = [metricsRule];

      clusterRole.new() +
      clusterRole.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name) +
      clusterRole.withRules(rules),
    roleConfig:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;

      local configmapRule = policyRule.new() +
                            policyRule.withApiGroups(['']) +
                            policyRule.withResources([
                              'configmaps',
                            ]) +
                            policyRule.withVerbs(['get']);

      role.new() +
      role.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name + '-config') +
      role.mixin.metadata.withNamespace($._config.namespace) +
      role.withRules(configmapRule),
    roleBindingConfig:
      local roleBinding = k.rbac.v1.roleBinding;

      roleBinding.new() +
      roleBinding.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name + '-config') +
      roleBinding.mixin.metadata.withNamespace($._config.namespace) +
      roleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      roleBinding.mixin.roleRef.withName('prometheus-' + $._config.ams.prometheus.name + '-config') +
      roleBinding.mixin.roleRef.mixinInstance({ kind: 'Role' }) +
      roleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.ams.prometheus.name, namespace: $._config.namespace }]),
    clusterRoleBinding:
      local clusterRoleBinding = k.rbac.v1.clusterRoleBinding;

      clusterRoleBinding.new() +
      clusterRoleBinding.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name) +
      clusterRoleBinding.mixin.roleRef.withApiGroup('rbac.authorization.k8s.io') +
      clusterRoleBinding.mixin.roleRef.withName('prometheus-' + $._config.ams.prometheus.name) +
      clusterRoleBinding.mixin.roleRef.mixinInstance({ kind: 'ClusterRole' }) +
      clusterRoleBinding.withSubjects([{ kind: 'ServiceAccount', name: 'prometheus-' + $._config.ams.prometheus.name, namespace: $._config.namespace }]),
    roleSpecificNamespaces:
      local role = k.rbac.v1.role;
      local policyRule = role.rulesType;
      local coreRule = policyRule.new() +
                       policyRule.withApiGroups(['']) +
                       policyRule.withResources([
                         'services',
                         'endpoints',
                         'pods',
                       ]) +
                       policyRule.withVerbs(['get', 'list', 'watch']);

      local newSpecificRole(namespace) =
        role.new() +
        role.mixin.metadata.withName('prometheus-' + $._config.ams.prometheus.name) +
        role.mixin.metadata.withNamespace(namespace) +
        role.withRules(coreRule);

      local roleList = k3.rbac.v1.roleList;
      roleList.new([newSpecificRole(x) for x in $._config.ams.prometheus.namespaces]),
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
        resourceRequirements.withLimits($._config.ams.resourceLimits) +
        resourceRequirements.withRequests($._config.ams.resourceRequests);

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
          version: $._config.versions.prometheus,
          baseImage: $._config.imageRepos.prometheus,
          serviceAccountName: 'prometheus-' + $._config.ams.prometheus.name,
          serviceMonitorSelector: {},
          podMonitorSelector: {},
          serviceMonitorNamespaceSelector: {},
          nodeSelector: { 'kubernetes.io/os': 'linux' },
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
            container.new('remote-write-proxy', '%s:%s' % [$._config.imageRepos.remoteWriteProxy, $._config.versions.remoteWriteProxy]) +
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
          name: 'prometheus-ams',
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
          forward_host: $._config.ams.remoteWriteTarget,
          thanos_tenant: $._config.ams.receiveTenantId,
        }),
      }),
  },
} + {
  local prom = super.prometheusAms,
  prometheusAms+:: {
    template+:
      list.asList('prometheus-observatorium-ams', prom, []) +
      list.withNamespace($._config) +
      list.withPrometheusImage($._config) +
      list.withPrometheusResources($._config.ams.prometheus.resourceRequests, $._config.ams.prometheus.resourceLimits),
  },
}
