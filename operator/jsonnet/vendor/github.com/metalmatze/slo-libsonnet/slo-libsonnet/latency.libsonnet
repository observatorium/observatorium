local util = import '_util.libsonnet';

{
  latency(param):: {
    local slo = {
      metric: error 'must set metric for latency',
      selectors: error 'must set selectors for latency',
      quantile: error 'must set quantile for latency',
      labels: [],
    } + param,

    local labels =
      util.selectorsToLabels(slo.selectors) +
      util.selectorsToLabels(slo.labels),

    recordingrule(quantile=slo.quantile):: {
      expr: |||
        histogram_quantile(%.2f, sum(%s_bucket{%s}) by (le))
      ||| % [
        quantile,
        slo.metric,
        slo.selectors,
      ],
      record: '%s:histogram_quantile' % slo.metric,

      labels: labels {
        quantile: '%.2f' % quantile,
      },
    },

    local _recordingrule = self.recordingrule(),

    alertWarning: {
      expr: |||
        %s > %.3f
      ||| % [_recordingrule.record, slo.warning],
      'for': '5m',
      labels: labels {
        severity: 'warning',
      },
    },

    alertCritical: {
      expr: |||
        %s > %.3f
      ||| % [_recordingrule.record, slo.critical],
      'for': '5m',
      labels: labels {
        severity: 'critical',
      },
    },

    grafana: {
      gauge: {
        type: 'gauge',
        title: 'P99 Latency',
        datasource: '$datasource',
        options: {
          maxValue: '1.5',  // TODO might need to be configurable
          minValue: 0,
          thresholds: [
            {
              color: 'green',
              index: 0,
              value: null,
            },
            {
              color: '#EAB839',
              index: 1,
              value: slo.warning,
            },
            {
              color: 'red',
              index: 2,
              value: slo.critical,
            },
          ],
          valueOptions: {
            decimals: null,
            stat: 'last',
            unit: 'dtdurations',
          },
        },
        targets: [
          {
            expr: '%s{quantile="%.2f"}' % [
              _recordingrule.record,
              slo.quantile,
            ],
            format: 'time_series',
          },
        ],
      },
      graph: {
        type: 'graph',
        title: 'Request Latency',
        datasource: '$datasource',
        targets: [
          {
            expr: 'max(%s) by (quantile)' % _recordingrule.record,
            legendFormat: '{{ quantile }}',
          },
        ],
        yaxes: [
          {
            show: true,
            min: '0',
            max: null,
            format: 's',
            decimals: 1,
          },
          {
            show: false,
          },
        ],
        xaxis: {
          show: true,
          mode: 'time',
          name: null,
          values: [],
          buckets: null,
        },
        yaxis: {
          align: false,
          alignLevel: null,
        },
        lines: true,
        fill: 2,
        span: 12,
        linewidth: 1,
        dashes: false,
        dashLength: 10,
        paceLength: 10,
        points: false,
        pointradius: 2,
        thresholds: [
          {
            value: slo.warning,
            colorMode: 'warning',
            op: 'gt',
            fill: true,
            line: true,
            yaxis: 'left',
          },
          {
            value: slo.critical,
            colorMode: 'critical',
            op: 'gt',
            fill: true,
            line: true,
            yaxis: 'left',
          },
        ],
      },
    },
  },
}
