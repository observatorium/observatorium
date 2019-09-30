{
  observatoriumRangeQuery:: {
    errors: {
      metric: 'http_requests_total',
      selectors: ['job="thanos-querier"'],

      errorBudget: 1.0 - 0.99,
    },
  },
}
