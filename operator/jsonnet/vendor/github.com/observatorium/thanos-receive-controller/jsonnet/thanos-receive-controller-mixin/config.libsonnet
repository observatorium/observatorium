{
  _config+:: {
    thanosReceiveJobPrefix: 'thanos-receive',
    thanosReceiveSelector: 'job=~"%s.*"' % self.thanosReceiveJobPrefix,

    thanosReceiveControllerJobPrefix: 'thanos-receive-controller',
    thanosReceiveControllerSelector: 'job=~"%s.*"' % self.thanosReceiveControllerJobPrefix,

    grafanaThanos: {
      dashboardNamePrefix: 'Thanos / ',
      dashboardTags: ['thanos-receive-controller-mixin', 'observatorium'],

      dashboardReceiveControllerTitle: '%(dashboardNamePrefix)sReceive Controller' % $._config.grafanaThanos,
    },
  },
}
