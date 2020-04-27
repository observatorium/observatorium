local g = import 'thanos-mixin/lib/thanos-grafana-builder/builder.libsonnet';

{
  grafanaDashboards+:: {
    'receive-controller.json':
      g.dashboard($._config.grafanaThanos.dashboardReceiveControllerTitle)
      .addRow(
        g.row('Reconcile Attempts')
        .addPanel(
          g.panel('Rate') +
          g.queryPanel(
            'thanos_receive_controller_reconcile_attempts_total{namespace="$namespace",%(thanosReceiveControllerSelector)s}' % $._config,
            'rate'
          )
        )
        .addPanel(
          g.panel('Errors') +
          g.queryPanel(
            'sum(rate(thanos_receive_controller_reconcile_errors_total{namespace="$namespace",%(thanosReceiveControllerSelector)s}[$interval])) by (type)' % $._config,
            '{{type}}'
          ) +
          { yaxes: g.yaxes('percentunit') } +
          g.stack
        )
      )
      .addRow(
        g.row('Configmap Changes')
        .addPanel(
          g.panel('Rate') +
          g.queryPanel(
            'thanos_receive_controller_configmap_change_attempts_total{namespace="$namespace",%(thanosReceiveControllerSelector)s}' % $._config,
            'rate',
          )
        )
        .addPanel(
          g.panel('Errors') +
          g.queryPanel(
            'sum(rate(thanos_receive_controller_configmap_change_errors_total{namespace="$namespace",%(thanosReceiveControllerSelector)s}[$interval])) by (type)' % $._config,
            '{{type}}'
          ) +
          { yaxes: g.yaxes('percentunit') } +
          g.stack
        )
      )
      .addRow(
        g.row('(Receive) Hashring Config Refresh')
        .addPanel(
          g.panel('Rate') +
          g.queryPanel(
            'sum(rate(thanos_receive_hashrings_file_changes_total{namespace="$namespace",%(thanosReceiveSelector)s}[$interval]))' % $._config,
            'all'
          )
        )
        .addPanel(
          g.panel('Errors') +
          g.qpsErrTotalPanel(
            'thanos_receive_hashrings_file_errors_total{namespace="$namespace",%(thanosReceiveSelector)s}' % $._config,
            'thanos_receive_hashrings_file_changes_total{namespace="$namespace",%(thanosReceiveSelector)s}' % $._config,
          )
        )
      )
      .addRow(
        g.row('Hashring Status')
        .addPanel(
          g.panel('Nodes per Hashring') +
          g.queryPanel(
            [
              'avg(thanos_receive_controller_hashring_nodes{namespace="$namespace",%(thanosReceiveControllerSelector)s}) by (name)' % $._config,
              'avg(thanos_receive_hashring_nodes{namespace="$namespace",%(thanosReceiveSelector)s}) by (name)' % $._config,
            ],
            [
              'receive controller {{name}}',
              'receive {{name}}',
            ]
          )
        )
        .addPanel(
          g.panel('Tenants per Hashring') +
          g.queryPanel(
            [
              'avg(thanos_receive_controller_hashring_tenants{namespace="$namespace",%(thanosReceiveControllerSelector)s}) by (name)' % $._config,
              'avg(thanos_receive_hashring_tenants{namespace="$namespace",%(thanosReceiveSelector)s}) by (name)' % $._config,
            ],
            [
              'receive controller {{name}}',
              'receive {{name}}',
            ],
          )
        )
      )
      .addRow(
        g.row('Hashring Config')
        .addPanel(
          g.panel('Last Updated') +
          g.statPanel(
            'time() - max(thanos_receive_controller_configmap_last_reload_success_timestamp_seconds{namespace="$namespace",%(thanosReceiveControllerSelector)s})' % $._config,
            's'
          ) +
          {
            postfix: 'ago',
            decimals: 0,
          }
        )
        .addPanel(
          g.panel('Last Updated') +
          g.statPanel(
            'time() - max(thanos_receive_config_last_reload_success_timestamp_seconds{namespace="$namespace",%(thanosReceiveSelector)s})' % $._config,
            's'
          ) +
          {
            postfix: 'ago',
            decimals: 0,
          }
        )
      )
      +
      g.template('namespace', 'kube_pod_info') +
      g.template('job', 'up', 'namespace="$namespace",%(thanosReceiveControllerSelector)s' % $._config, true, '%(thanosReceiveControllerJobPrefix)s.*' % $._config),
  },
} +
(import 'defaults.libsonnet')
