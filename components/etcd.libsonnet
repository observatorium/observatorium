{
  local etcd = self,

  config:: {
    name: error 'must provide name',
    namespace: error 'must provide namespace',
    version: error 'must provide version',
    image: error 'must provide image',
    replicas: error 'must provide replicas',

    commonLabels:: {
      'app.kubernetes.io/name': 'loki',
      'app.kubernetes.io/instance': etcd.config.name,
      'app.kubernetes.io/version': etcd.config.version,
      'app.kubernetes.io/component': 'ring-store',
    },

    podLabelSelector:: {
      [labelName]: etcd.config.commonLabels[labelName]
      for labelName in std.objectFields(etcd.config.commonLabels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },
  },

  configmap: {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: etcd.config.name,
      namespace: etcd.config.namespace,
      labels: etcd.config.commonLabels,
    },
    data: {
      'etcd-pre-stop.sh': |||
        EPS=""
        for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
            EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}:2379"
        done

        HOSTNAME=$(hostname)
        AUTH_OPTIONS=""

        member_hash() {
            etcdctl $AUTH_OPTIONS member list | grep http://${HOSTNAME}.${SET_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1
        }

        SET_ID=${HOSTNAME##*[^0-9]}

        if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
            echo "Removing ${HOSTNAME} from etcd cluster"
            ETCDCTL_ENDPOINT=${EPS} etcdctl $AUTH_OPTIONS member remove $(member_hash)
            if [ $? -eq 0 ]; then
                # Remove everything otherwise the cluster will no longer scale-up
                rm -rf /var/run/etcd/*
            fi
        fi
      |||,
      'etcd-server.sh': |||
        HOSTNAME=$(hostname)
        AUTH_OPTIONS=""
        # store member id into PVC for later member replacement
        collect_member() {
            while ! etcdctl $AUTH_OPTIONS member list &>/dev/null; do sleep 1; done
            etcdctl $AUTH_OPTIONS member list | grep http://${HOSTNAME}.${SET_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1 > /var/run/etcd/member_id
            exit 0
        }

        eps() {
            EPS=""
            for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}:2379"
            done
            echo ${EPS}
        }

        member_hash() {
            etcdctl $AUTH_OPTIONS member list | grep http://${HOSTNAME}.${SET_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1
        }

        # we should wait for other pods to be up before trying to join
        # otherwise we got "no such host" errors when trying to resolve other members
        for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
            while true; do
                echo "Waiting for ${SET_NAME}-${i}.${SET_NAME} to come up"
                ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME} > /dev/null && break
                sleep 1s
            done
        done

        # re-joining after failure?
        if [ -e /var/run/etcd/default.etcd ]; then
            echo "Re-joining etcd member"
            member_id=$(cat /var/run/etcd/member_id)

            # re-join member
            ETCDCTL_ENDPOINT=$(eps) etcdctl $AUTH_OPTIONS member update ${member_id} http://${HOSTNAME}.${SET_NAME}:2380 | true
            exec etcd --name ${HOSTNAME} \
                --listen-peer-urls http://0.0.0.0:2380 \
                --listen-client-urls http://0.0.0.0:2379\
                --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
                --data-dir /var/run/etcd/default.etcd

        fi

        # etcd-SET_ID
        SET_ID=${HOSTNAME##*[^0-9]}

        # adding a new member to existing cluster (assuming all initial pods are available)
        if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
            export ETCDCTL_ENDPOINT=$(eps)

            # member already added?
            MEMBER_HASH=$(member_hash)
            if [ -n "${MEMBER_HASH}" ]; then
                # the member hash exists but for some reason etcd failed
                # as the datadir has not be created, we can remove the member
                # and retrieve new hash
                etcdctl $AUTH_OPTIONS member remove ${MEMBER_HASH}
            fi

            echo "Adding new member"
            etcdctl $AUTH_OPTIONS member add ${HOSTNAME} http://${HOSTNAME}.${SET_NAME}:2380 | grep "^ETCD_" > /var/run/etcd/new_member_envs

            if [ $? -ne 0 ]; then
                echo "Exiting"
                rm -f /var/run/etcd/new_member_envs
                exit 1
            fi

            cat /var/run/etcd/new_member_envs
            source /var/run/etcd/new_member_envs

            collect_member &

            exec etcd --name ${HOSTNAME} \
                --listen-peer-urls http://0.0.0.0:2380 \
                --listen-client-urls http://0.0.0.0:2379 \
                --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
                --data-dir /var/run/etcd/default.etcd \
                --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}:2380 \
                --initial-cluster ${ETCD_INITIAL_CLUSTER} \
                --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}

        fi

        PEERS=""
        for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
            PEERS="${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}:2380"
        done

        collect_member &

        # join member
        exec etcd --name ${HOSTNAME} \
            --initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}:2380 \
            --listen-peer-urls http://0.0.0.0:2380 \
            --listen-client-urls http://0.0.0.0:2379 \
            --advertise-client-urls http://${HOSTNAME}.${SET_NAME}:2379 \
            --initial-cluster-token etcd-cluster-1 \
            --initial-cluster ${PEERS} \
            --initial-cluster-state new \
            --data-dir /var/run/etcd/default.etcd
      |||,
    },
  },

  service: {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: etcd.config.name,
      namespace: etcd.config.namespace,
      labels: etcd.config.commonLabels,
    },
    spec: {
      ports: [
        { name: 'client', targetPort: 2379, port: 2379, protocol: 'TCP' },
        { name: 'peer', targetPort: 2380, port: 2380, protocol: 'TCP' },
      ],
      selector: etcd.config.podLabelSelector,
    },
  },

  statefulSet:
    local c = {
      name: 'etch',
      image: etcd.config.image,
      command: [
        '/bin/sh',
        '-ec',
        '/scripts/etcd-server.sh',
      ],
      env: [
        { name: 'ETCDCTL_API', value: '3' },
        { name: 'INITIAL_CLUSTER_SIZE', value: '%d' % etcd.config.replicas },
        { name: 'SET_NAME', value: etcd.config.name },
      ],
      ports: [
        { name: port.name, containerPort: port.port }
        for port in etcd.service.spec.ports
      ],
      volumeMounts: [
        {
          name: 'scripts',
          mountPath: '/scripts',
        },
        {
          name: 'storage',
          mountPath: '/var/run/etcd',
        },
      ],
      resources: {
        requests: {
          cpu: '100m',
          memory: '128Mi',
        },
        limits: {
          cpu: '200m',
          memory: '256Mi',
        },
      },
      lifecycle: {
        preStop: {
          exec: {
            command: [
              '/bin/sh',
              '-ec',
              '/scripts/etcd-pre-stop.sh',
            ],
          },
        },
      },
    };

    {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: etcd.config.name,
        namespace: etcd.config.namespace,
        labels: etcd.config.commonLabels,
      },
      spec: {
        replicas: etcd.config.replicas,
        selector: { matchLabels: etcd.config.podLabelSelector },
        serviceName: etcd.service.metadata.name,
        template: {
          metadata: {
            labels: etcd.config.commonLabels,
          },
          spec: {
            containers: [c],
            volumes: [
              { name: 'scripts', configMap: { name: etcd.configmap.metadata.name, defaultMode: std.parseOctal('777') } },
              { name: 'storage', emptyDir: {} },
            ],
            volumeClaimTemplates:: null,
          },
        },
      },
    },

  manifests:: {
    'ring-store-service': etcd.service,
    'ring-store-configmap': etcd.configmap,
    'ring-store-statefulset': etcd.statefulSet,
  },
}
