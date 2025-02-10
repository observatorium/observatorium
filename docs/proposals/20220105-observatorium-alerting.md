# Observatorium Alerting

* **Owners:**
  * `@onprem` `@bwplotka`

* **Related Tickets:**
  * https://issues.redhat.com/browse/MON-1761

* **Other docs:*
  * None

> TL;DR: For monitoring use cases, especially tenants of the metric signal of Observatorium, users want to be able to configure Prometheus-based alerts that will be evaluated for a given interval. If triggered they should notify the desired target (receiver) (e.g PagerDuty) using Alerting Routing Configuration API. User secrets stored in Vault. Users should also be able to see the status of all alerts through Prometheus Rules HTTP API. This proposal specifies how such a system can be deployed within Observatorium.

## Why

Alerting is a critical component of reliable monitoring. We can build automation or notify humans in a reactive or proactive manner instead of watching dashboards.

### Pitfalls of the current solution

Observatorium does not have alerting support.

## Goals

Definition:
* Alert Routing: Configuration of Alertmanager related to receiver target config and its routing.

Goals:

* Tenants are able to create, edit and see alert routing configuration:
  * via new GET, DELETE and PUT HTTP API
* Tenants are able to see configuration presently loaded in AM
  * This API should support only GET method
  * Potentially work with Alertmanager Team on upstream API.
* AM routing API should be multi-tenant.
* Tenant's alerts will be evaluated and triggering proper receivers in AM.

## Non-Goals

* Adding [Alerts API](https://prometheus.io/docs/prometheus/latest/querying/api/#alerts)
  * See [rationales](#alert-evaluation-status)
* Tenant should be able to create, edit and read status of alerting rules:
  * This is not the goal, since this is part of the [Rules API](https://github.com/observatorium/observatorium/blob/main/docs/proposals/20201019-prometheus-rules-in-observatorium-api.md) implementation.
* UI for Alerts/Rules, Alert Status, AM Routing, AM Alerts …
* Alertmanager Alerts API, Silences, Inhibits API
  * We should focus on that in next iteration.
* Scaling Rulers
* Different configuration than Alertmanager with 2 replicas.
* Testing routing configuration or UI allowing better experience for configuring those. It is a common issue that routing can be very misleading.

## Abbreviations

AM: Alertmanager

## Audience of this proposal

Observatorium Devs and Users.

## How

### Current Solution

Within work for [Rules API proposal](https://observatorium.io/docs/proposals/20201019-prometheus-rules-in-observatorium-api.md/) we already added Rules APIs for reading and configuring rules which include both recording and alerting rules.

For example, to save alerts we can already PUT the following object to `/api/metrics/v1/{tenant}/api/v1/rules/raw`

```
groups:
  - name: test-oidc
    interval: 5s
    rules:
      - alert: HighRequestLatency
        expr: job:request_latency_seconds:mean5m{job="second"} > 0.5
        for: 10m
        labels:
          team: alpha
```

Such a rule will be saved through Observatorium API, via [rule-objstore](https://github.com/observatorium/rules-objstore) to object storage for the specified tenant. It can be then listed through a similar endpoint and GET. Such configuration will be then eventually synchronized with tenant’s Rulers using [rule-syncer](https://github.com/observatorium/thanos-rule-syncer) (1m poll by default). This is what was agreed so far and is being tested in downstream RHOBS (Red Hat Observability Service) deployment of Observatorium.

### ConfigSync

We propose to move https://github.com/observatorium/thanos-rule-syncer to https://github.com/observatorium/config-sync (related ticket: https://issues.redhat.com/browse/MON-2129) that will understand Prometheus Rules configuration, Prometheus Alert Routing configuration and in future silences, inhibitors tpp etc. Similarly we propose to move https://github.com/observatorium/rules-objstore to be a generic configuration library, potentially in new github.com/observatorium/config-sync repo. It could be called a `config-client` module. Such library would need to support not only saving and retrieving Rules but also Alerting Routing configuration (and potentially silencer, inhibits and other user configuration in the future). Since alerting routing requires secrets we might need to implement various backends in this client (that's why no `objstore` suffix. ). See [chapter about secrets](#managing-user-secrets) for more details.

ConfigSync would then import such config-client library internally. This will allow us to have single focused codebase for all different user configurations we will need. It will also allow for consistency around using it, monitoring, debugging, using secrets etc.

We propose potential https://github.com/observatorium/config-sync to be deployed as a sidecar to each component needing configuration to be present and reloaded dynamically.

As a consequence of this decision I propose Observatorium API to also import https://github.com/observatorium/config-sync config-client library directly to allow User APIs to manage those items (Rules and Alert routing configuration so far).

### Alert Evaluation

The Ruler itself has to communicate on 4 levels. All connections are presented in the diagram below.

![ruler-connections-diagram](../assets/alerting-ruler.png)

1. Alerts will be evaluated based on periodic requests to any replica of HTTP Query API, potentially through Query-frontend.
2. Ruler evaluates Alerts that can be triggered. Triggered alerts will be pushed to all replicas of the Alertmanager (discussed later) through AM HTTP API.
3. Ruler exposes gRPC Rules API which is consumed by Queriers. Querier than can expose multi-tenant HTTP Rules API that can be consumed by Observatorium API.
4. It saves samples to any replica of Receiver.

### Rules API (non raw one)

We expect Rulers to be connected to Queriers through [gRPC Rules API](https://github.com/thanos-io/thanos/blob/main/pkg/rules/rulespb/rpc.proto#L29) (NOTE, not Store API as we have now). This allows queries to be able to present federated views of Recording and Alerting Rules loaded and active alerts through the [Promethus Rules HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/#rules).

This API is different but a has an overlap with `/api/metrics/v1/{tenant}/api/v1/rules/raw`, which represents what was configured (but not necessarily loaded).

In order to fulfill the full story we also need:

* Implement matching / filtering by label/tenant on Rules HTTP API of Querier.

This will allow users looking at the Thanos Rules API (or Query UI) to only see the alerts that belong to them.

### Alert Evaluation Status

On top of mentioned Rules API, Prometheus also exposes [HTTP Alert API](https://prometheus.io/docs/prometheus/latest/querying/api/#alerts). We never expose it within Observatorium and Thanos, but Rules API in fact already provides exactly that information (on top of other data).

Because of that we propose to not add this API, which will avoid a lot of work. In future iteration (e.g. if we would add UI), we could add Alert HTTP API in Observatorium which would extract data needed by this API from Rules API.

### Alert Routing

It is crucial that we are able to reliably forward alerts to correct notifiers as soon as they are triggered. We propose deploying 2-replica Alertmanager HA, through our configuration (without the help of the Prometheus Operator, which was usually used in our past deployments).

We propose defining an additional HTTP API for routing on `/api/metrics/v1/{tenant}/api/v1/routing/raw` path in Observatorium API. Path that supports GET, DELETE and PUT methods. This API's purpose will be to communicate with `config-client`, in order to update routing configurations on behalf of the tenants.

In addition to the 'raw' API, which operates on the AM configuration, we want to allow our tenants to see configuration which is currently in use in AM for them. This should improve tenants' ability to troubleshoot potential issue with their configuration. This aim might require further engagement in the AM upstream, although AM API already supports certain GET operations, [including `/status`](https://github.com/prometheus/alertmanager/blob/main/api/v1/api.go#L171), which is capable of returning the loaded AM config.

We propose to base the types on [Prometheus Operator AM Config](https://github.com/prometheus-operator/prometheus-operator/blob/9c0db5656f04e005de6a0413fd8eb8f11ec99757/pkg/apis/monitoring/v1alpha1/alertmanager_config_types.go#L69). Proposed adding this API to upstream Alertmanager too (TODO).

We could allow all read API as well write of Inhibits and Silences of Alertmanager by adding auth and multi-tenancy layer, but we propose to deal with that in a separate epic/proposal.

### Dealing with tenancy in routing configuration
TO-DO

### Managing User Secrets

Something new in Alertmanager Routing configuration is the fact that Alertmanager can integrate with other systemes ("receivers") to send alerts to. This requires TLS, basic auth or other type of authZ and authN to be configured within this configuration. AM supports it, but it expects secrets to be usually baked in file or available in local AM filesystem.

This means we have to enter into user secret management story which is not trivial. We cannot just store those files in object storage in plain text. Encrypting them in rest sounds easy, but it involves lots of complexity around rotating keys, revoking secrets, process when leaked etc.

We propose adding different backends depending on user preference to github.com/observatorium/config-client, but for a start we propose adding support to reading secrets via Vault. Unfortunately we have to delegate storing user credentials to 3rd party dependency. See alternatives for different ideas.

## Alternatives

### Alertmanager per tenant
TO-DO

### Different Backend for Secrets

#### Kubernetes secrets
In theory Vault is a heavy dependency. Since we run on Kubernetes, we could just write Kubernetes secrets and store it there. There are some disadvantages:

* Does it really solve secret management in practice? Those secrets are not encrypted
* We add strong dependency on Kubernetes.
* Can our deployment have access to writing and reading Kubernetes secrets.

This option has to be explored more.

#### Storing secrets in object store in plain-text
TO-DO

### Allow tenants to only use pre-specified receivers
If we could remove the need for tenants to provide us with their receiver credentials altogether, we would not need to deal with the burden of managing user secrets at all. One possibility would be to have a set of receivers managed by the Observatorium administrator and to allow tenants to use these receivers in their route configurations.

This has the major downside of limiting tenants in what receivers they can use. While this might work for few operational scenarios (e.g. where Observatorium administrator and tenants share same Slack or OpsGenie instance as a part of single organization), ultimately it disallows the tenants to use receiver of their own choosing, which does not suffice to make this feature truly multi-tenant.

## FAQ

### Why thanos-rule-syncer is not currently sufficient for our use-case?

Thanos rule syncer only support syncing Rules. Does not support AM Routing configuration. It also requires separate service `thanos-objstore` to be deployed. This unnecessary microservice architecture will become a problem and a fuss to operate and maintain. We don't need to scale those components separately, so I propose to keep it in a single binary for now and near future. 

## Action Plan

The tasks to do in order to migrate to the new idea.

* [ ] Start using stateless Ruler in Observatorium
* [ ] Implement matching / filtering by label/tenant on Rules HTTP API of Querier.
* [ ] Move https://github.com/observatorium/rules-objstore to be a library in repo github.com/observatorium/config-sync called config-client or so.
* [ ] Add ability to talk to Vault in `config-client`
* [ ] Develop github.com/observatorium/config-sync based on github.com/observatorium/config-sync
* [ ] Add Alertmanager deployment
* [ ] Add Alertmanager Routing configuration support to config-client and config-sync
* [ ] Add Alertmanager Routing configuration API to Observatorium API
* [ ] Enable Observatorium API to use config-client directly
