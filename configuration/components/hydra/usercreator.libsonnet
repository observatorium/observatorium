{
  apiVersion: 'batch/v1',
  kind: 'Job',
  metadata: {
    name: 'usercreator',
    namespace: 'hydra',
  },
  spec: {
    ttlSecondsAfterFinished: 120,
    template: {
      spec: {
        restartPolicy: 'OnFailure',
        containers: [
          {
            name: 'usercreator',
            image: 'alpine/curl',
          },
        ],
      },
    },
  },
}
