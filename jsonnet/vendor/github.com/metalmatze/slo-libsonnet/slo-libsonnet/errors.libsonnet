local util = import '_util.libsonnet';

{
  errors(param):: {
    local slo = {
      metric: error 'must set metric for errors',
      selectors: error 'must set selectors for errors',
      labels: [],
      rate: '5m',
      codeSelector: 'code',
    } + param,

    local labels =
      util.selectorsToLabels(slo.selectors) +
      util.selectorsToLabels(slo.labels),

    local recordingrule = {
      expr: |||
        sum by (status_class) (
          label_replace(
            rate(%s{%s}[%s]
          ), "status_class", "${1}xx", "%s", "([0-9])..")
        )
      ||| % [
        slo.metric,
        std.join(',', slo.selectors),
        slo.rate,
        slo.codeSelector,
      ],
      record: 'status_class:%s:rate%s' % [
        slo.metric,
        slo.rate,
      ],
      labels: labels,
    },
    recordingrule: recordingrule,

    alertWarning: {
      expr: |||
        (
          sum(%s{status_class="5xx"})
        /
          sum(%s)
        ) > %s
      ||| % [
        recordingrule.record,
        recordingrule.record,
        slo.warning,
      ],
      'for': '5m',
      labels: labels {
        severity: 'warning',
      },
    },

    alertCritical: {
      expr: |||
        (
          sum(%s{status_class="5xx"})
        /
          sum(%s)
        ) > %s
      ||| % [
        recordingrule.record,
        recordingrule.record,
        slo.critical,
      ],
      'for': '5m',
      labels: labels {
        severity: 'critical',
      },
    },

    grafana: {
      graph: {
        span: 12,
        aliasColors: {
          '1xx': '#EAB839',
          '2xx': '#7EB26D',
          '3xx': '#6ED0E0',
          '4xx': '#EF843C',
          '5xx': '#E24D42',
          success: '#7EB26D',
          'error': '#E24D42',
        },
        datasource: '$datasource',
        legend: {
          avg: false,
          current: false,
          max: false,
          min: false,
          show: true,
          total: false,
          values: false,
        },
        targets: [
          {
            expr: '%s' % recordingrule.record,
            format: 'time_series',
            intervalFactor: 2,
            legendFormat: '{{status_class}}',
            refId: 'A',
            step: 10,
          },
        ],
        title: 'HTTP Response Codes',
        tooltip: {
          shared: true,
          sort: 0,
          value_type: 'individual',
        },
        type: 'graph',
      },
    },
  },
}
