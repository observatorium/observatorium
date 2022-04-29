{
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: 'hydra',
    namespace: 'hydra',
  },
  spec: {
    selector: {
      app: 'hydra',
    },
    ports: [
      {
        protocol: 'TCP',
        port: 4445,
        name: 'admin',
        targetPort: 'admin',
      },
      {
        protocol: 'TCP',
        port: 4444,
        name: 'public',
        targetPort: 'public',
      },
      {
        protocol: 'TCP',
        port: 5555,
        name: 'token',
        targetPort: 'token',
      },
    ],
  },
}
