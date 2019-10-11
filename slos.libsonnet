{
  observatoriumQuery:: {
    errors: {
      metric: 'http_requests_total',
      selectors: ['job="thanos-querier"'],

      errorBudget: 1.0 - 0.99,
    },
  },
  telemeterServerUpload:: {
    errors: {
      metric: 'http_requests_total',
      selectors: [
        'service="telemeter-server"',
        'handler="upload"',
      ],
      errorBudget: 1.0 - 0.90,
    },
  },
}
