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

* Tenants should be able to read the status of active alerts 
  * via ([Prometheus Rules API GET](https://prometheus.io/docs/prometheus/latest/querying/api/#rules))
* Tenants are able to configure alert routing:
  * via new GET, DELETE and PUT HTTP API
  * Potentially work with Alertmanager Team on upstream API.
* Both Rules API and AM routing API should be multi-tenant.
* Tenant's alerts will be evaluated and triggering proper receivers in AM.

## Non-Goals

* Adding [Alerts API](https://prometheus.io/docs/prometheus/latest/querying/api/#alerts)
  * See [rationales](#alert-evaluation-status)
* Tenant should be able to create and edit alerting rules:
  * Alerting Rules via GET and PUT via HTTP (Rules API)
  * This is not the goal, because it should be done when Rules API is done.
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

We propose to move https://github.com/observatorium/thanos-rule-syncer to https://github.com/observatorium/config-sync (related ticket: https://issues.redhat.com/browse/MON-2129) that will understand Prometheus Rules configuration, Prometheus Alert Routing configuration and in future silences, inhibitors tpp etc. Similarly we propose to move https://github.com/observatorium/rules-objstore to be a generic configuration library, potentially in new https://github.com/observatorium/config-sync repo. It could be called a `config-client` module. Such library would need to support not only saving and retrieving Rules but also Alerting Routing configuration (and potentially silencer, inhibits and other user configuration in the future). Since alerting routing requires secrets we might need to implement various backends in this client (that's why no `objstore` suffix. ). See [chapter about secrets](#managing-user-secrets) for more details.

ConfigSync would then import such config-client library internally. This will allow us to have single focused codebase for all different user configurations we will need. This will allow us to maintain consistency around using it, monitoring, debugging, using secrets etc.

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

This API is different but a bit overlapping with `/api/metrics/v1/{tenant}/api/v1/rules/raw` which represents what was configured (not loaded).

In order to fulfill the full story we also need:

* Implement matching / filtering by label/tenant on Rules HTTP API of Querier.

This will allow users looking at the Thanos Rules API (or Query UI) to only see the alerts that belong to them.

### Alert Evaluation Status

On top of mentioned Rules API, Prometheus also exposes [HTTP Alert API](https://prometheus.io/docs/prometheus/latest/querying/api/#alerts). We never expose it within Observatorium and Thanos, but Rules API in fact already provides exactly that information (on top of other data).

Because of that we propose to not add this API, which will avoid a lot of work. In future iteration (e.g. if we would add UI), we could add Alert HTTP API in Observatorium which would extract data needed by this API from Rules API.

### Alert Routing

It is crucial that we are able to reliably forward alerts to correct notifiers as soon as they are triggered. We propose deploying 2-replica Alertmanager HA, through our configuration (without the help of the Prometheus Operator, which was usually used in our past deployments).

We propose defining an additional HTTP API for routing on `/api/metrics/v1/{tenant}/api/v1/routing/raw` path in Observatorium API. Path that supports both GET, DELETE and PUT. 

We propose to base the type on [Prometheus Operator AM Config](https://github.com/prometheus-operator/prometheus-operator/blob/9c0db5656f04e005de6a0413fd8eb8f11ec99757/pkg/apis/monitoring/v1alpha1/alertmanager_config_types.go#L69). Proposed adding this API to upstream Alertmanager too (TODO).

We could allow all read API as well write of Inhibits and Silences of Alertmanager by adding auth and multi-tenancy layer, but we propose to deal with that in a separate epic/proposal.

### Managing User Secrets

Something new in Alertmanager Routing configuration is the fact that Alertmanager can integrate with other systemes ("receivers") to send alerts to. This requires TLS, basic auth or other type of authZ and authN to be configured within this configuration. AM supports it, but it expects secrets to be usually baked in file or available in local AM filesystem.

This means we have to enter into user secret management story which is not trivial. We cannot just store those files in object storage in plain text. Encrypting them in rest sounds easy, but it involves lots of complexity around rotating keys, revoking secrets, process when leaked etc.

We propose adding different backends depending on user preference to https://github.com/observatorium/config-client, but for a start we propose adding support to reading secrets via Vault. Unfortunately we have to delegate storing user credentials to 3rd party dependency. See alternatives for different ideas.

## Alternatives

### Alertmanager per tenant

### Different Backend for Secrets

In theory Vault is a heavy dependency. Since we run on Kubernetes, we could just write Kubernetes secrets and store it there. There are some disadvantages:

* Does it really solve secret management in practice? Those secrets are not encrypted
* We add strong dependency on Kubernetes.
* Can our deployment have access to writing and reading Kubernetes secrets.

This option has to be explored more.

## FAQ

### Why thanos-rule-syncer is not currently sufficient for our use-case?

Thanos rule syncer only support syncing Rules. Does not support AM Routing configuration. It also requires separate service `thanos-objstore` to be deployed. This unnecessary microservice architecture will become a problem and a fuss to operate and maintain. We don't need to scale those components separately, so I propose to keep it in a single binary for now and near future. 

## Action Plan

The tasks to do in order to migrate to the new idea.

* [ ] Start using stateless Ruler in Observatorium
* [ ] Implement matching / filtering by label/tenant on Rules HTTP API of Querier.
* [ ] Move https://github.com/observatorium/rules-objstore to be a library in repo https://github.com/observatorium/config-sync called config-client or so.
* [ ] Add ability to talk to Vault in `config-client`
* [ ] Develop https://github.com/observatorium/config-sync based on https://github.com/observatorium/config-sync
* [ ] Add Alertmanager deployment
* [ ] Add Alertmanager Routing configuration support to config-client and config-sync
* [ ] Add Alertmanager Routing configuration API to Observatorium API
* [ ] Enable Observatorium API to use config-client directly
