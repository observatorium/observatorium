local errors = import 'errors.libsonnet';
local util = import '_util.libsonnet';
{
  errorburn(param):: {
    local slo = {
      metric: error 'must set metric for errorburn',
      selectors: error 'must set selectors for errorburn',
      errorBudget: error 'must set errorBudget for errorburn',
      labels: [],
      codeSelector: 'code',
    } + param,

    local rates = ['5m', '30m', '1h', '2h', '6h', '1d', '3d'],

    local labels =
      util.selectorsToLabels(slo.selectors) +
      util.selectorsToLabels(slo.labels),

    local errorRatesWithRate = [
      errors.errors({
        metric: slo.metric,
        selectors: slo.selectors,
        rate: rate,
        codeSelector: slo.codeSelector,
      }).recordingrule {
        // We need to communicate the rate to the errorPercentage step
        // They will be remove after that again
        labels+: { __tmpRate__: rate },
      }
      for rate in rates
    ],

    local errorPercentages = [
      {
        expr: |||
          sum(%s{%s})
          /
          sum(%s{%s})
        ||| % [
          err.record,
          std.join(',', slo.selectors + ['status_class="5xx"']),
          err.record,
          std.join(',', slo.selectors),
        ],
        record: 'status_class_5xx:%s:ratio_rate%s' % [slo.metric, err.labels.__tmpRate__],
        labels: labels,
      }
      for err in errorRatesWithRate
    ],

    // Remove __tmpRate__ label from errorRates rules again
    local errorRates = std.map(
      function(rule) rule {
        local ls = super.labels,
        labels: {
          [k]: ls[k]
          for k in std.objectFields(ls)
          if !std.setMember(k, ['__tmpRate__'])
        },
      },
      errorRatesWithRate,
    ),

    recordingrules: errorRates + errorPercentages,

    local multiBurnRate30d = [
      {
        alert: 'ErrorBudgetBurn',
        expr: |||
          (
            %s{%s} > (14.4*%f)
            and
            %s{%s} > (14.4*%f)
          )
          or
          (
            %s{%s} > (6*%f)
            and
            %s{%s} > (6*%f)
          )
        ||| % [
          errorPercentages[2].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[0].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[4].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[1].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
        ],
        labels: labels {
          severity: 'critical',
        },
      },
      {
        alert: 'ErrorBudgetBurn',
        expr: |||
          (
            %s{%s} > (3*%f)
            and
            %s{%s} > (3*%f)
          )
          or
          (
            %s{%s} > (%f)
            and
            %s{%s} > (%f)
          )
        ||| % [
          errorPercentages[5].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[3].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[6].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
          errorPercentages[4].record,
          std.join(',', slo.selectors),
          slo.errorBudget,
        ],
        labels: labels {
          severity: 'warning',
        },
      },
    ],

    alerts: multiBurnRate30d,
  },
}
