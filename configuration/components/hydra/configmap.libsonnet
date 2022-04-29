{
  apiVersion: 'v1',
  kind: 'ConfigMap',
  metadata: {
    name: 'hydra-config',
    namespace: 'hydra',
  },
  data: {
    'config.yaml': 'dsn: sqlite:///var/lib/sqlite/hydra.sqlite?_fk=true\nstrategies:\n  access_token: jwt\nurls:\n  self:\n    issuer: http://hydra.hydra.svc.cluster.local:4444/',
  },
}
