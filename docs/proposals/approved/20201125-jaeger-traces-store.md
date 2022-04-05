# Jaeger traces store

* **Owners:**:
  * `@pavolloffay`

* **Related Tickets:**
  * [observatorium/issues/442](https://github.com/observatorium/observatorium/issues/442)
  * Depends on: [Traces ingestion API and OpenTelemetry collector](https://github.com/observatorium/observatorium/pull/443)

## Table of Contents

- [TLDR](#tldr)
- [Why](#why)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [How](#how)
  * [Jaeger configuration](#jaeger-configuration)
    + [Jaeger instance per tenant](#jaeger-instance-per-tenant)
      - [Architecture](#architecture)
    + [Single Jaeger instance - soft multitenancy](#single-jaeger-instance---soft-multitenancy)
  * [Jaeger Query](#jaeger-query)
    + [Expose only Jaeger Query API](#expose-only-jaeger-query-api)
    + [Expose Jaeger Query UI](#expose-jaeger-ui)
- [Alternatives](#alternatives)
  * [Grafana tempo](#grafana-tempo)
- [Action Plan](#action-plan)
- [References](#references)

<small>

<i>

<a href="http://ecotrust-canada.github.io/markdown-toc/">
Table of contents generated with markdown-toc
</a>

</i>

</small>

## TLDR

This proposal adds Jaeger deployment with a persistent storage (e.g. Elasticsearch) to store trace data and it as well defines query API to search for traces.

## Why

Trace store adds a crucial capability to Observatorium to persist and query tracing data.

## Goals

* Allow Observatorium to store and retrieve traces.
* Expose Jaeger query API for query trace data.

## Non-Goals

* Define trace ingestion API - it is done in [Traces ingestion API and OpenTelemetry collector](https://github.com/observatorium/observatorium/pull/443) proposal.

## How

The how is split into the following sections:
* Jaeger configuration - describes how Jaeger components are deployed.
* Jaeger Query - describes how users access data stored in Jaeger.

### Jaeger configuration

Currently, the most adopted and supported storage in Jaeger is Elasticsearch. Red Hat also supports Jaeger with Elasticsearch (6.x) as part of the OpenShift product. Therefore, this proposal assumes Elasticsearch will be used as a storage backend. Distributed tracing team at Red Hat is looking at alternative storages hence the storage technology can change in the future.

All deployment topologies assume that Jaeger will be deployed by the [Jaeger Operator](https://github.com/jaegertracing/jaeger-operator) which depends on [OpenShift Elasticserach operator](https://github.com/openshift/elasticsearch-operator) and [Strimzi operator](https://operatorhub.io/operator/strimzi-kafka-operator) if Kafka is used. Jaeger can be as well configured to use externally managed Elasticsearch or Kafka instances.

#### Jaeger instance per tenant

At the moment Jaeger supports only hard multitenancy. It means that each tenant uses a dedicated collector and query component. The data itself can be stored in a single storage. In Elasticsearch tenant data is isolated by using dedicated set of indices e.g. `tenant1-jaeger-*` for `tenant1` and `tenant2-jaeger-*` for `tenant2`.

Deploying dedicated set of components per tenant is not ideal from cost and operational standpoint, however at the moment it is the only possible solution. There are plans to add soft multitenancy to Jaeger to overcome this issue. See the next section for more details.

##### Architecture

Follows an architecture diagram showing Observatorium API with Jaeger deployment. A couple of high level notes:
* OpenTelemetry collector is used to translate OTLP to Jaeger gRCP and route the data to an appropriate Jaeger collector.
* The routing of data in OpenTelemetry collector is done by the [routing processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/routingprocessor#routing-processor). It can use HTTP header or attribute to make the decision. The HTTP header or attribute will be set in the Observatorium API.
* Routing to the Query service is done directly in the Observatorium API. For instance, it can be done by using tenant ID in Kubernetes service name `tenan1-jaeger-query.observatorium.svc.cluster.local`.

```
                                                                      +--------------------+
                                                                      |                    |
                                                +-------------------->|  Jaeger collector  |
                                                |                     |       tenant1      |
+-----------------------+         +-------------+-------------+       +----------+---------+        +-----------------+
|                       |         |                           |                  |                  |                 |
|   Observatorium API   +-------->|  OpenTelemetry collector  |                  +------------------>  Elasticsearch  |
|                       |         |                           |                  |                  |                 |
+-----------+-----------+         +-------------+-------------+       +----------+---------+        +--------^--------+
            |                                   |                     |                    |                 |
            |                                   +-------------------->|  Jaeger collector  |                 |
            |                                                         |       tenant2      |                 |
            |                                                         +--------------------+                 |
            |                                                                                                |
            |                                                           +----------------+                   |
            |                                                           |                |                   |
            +---------------------------------------------------------->|  Jaeger Query  +-------------------+
            |                                                           |     tenant1    |                   |
            |                                                           +----------------+                   |
            |                                                                                                |
            |                                                           +----------------+                   |
            |                                                           |                |                   |
            +---------------------------------------------------------->|  Jaeger Query  +-------------------+
                                                                        |     tenant2    |
                                                                        +----------------+
```

#### Single Jaeger instance - soft multitenancy

Soft multitenancy in Jaeger is not supported at the moment. The progress is tracked in [jaegertracing/jaeger/3427](https://github.com/jaegertracing/jaeger/issues/3427).

Here are few issues that needs to be addressed. The list might not be complete:
* Service and operation name API. The service query API nor storage does not expose/store labels.
* Dependency/service architecture diagram. To support this feature dependency schema, query, collector and Spark aggregation would have to change.
* Data retention/TTL configurable pre tenant. The data retention in Elasticsearch is configurable per index.

### Jaeger Query

There are two ways users could access data stored in Jaeger:
1. Expose only Jaeger Query API in the Observatorium API.
2. Expose Jaeger UI outside the cluster.

#### Expose only Jaeger Query API

Jaeger query API would be exposed in the Observatorium API and users would have to deploy Jaeger UI and use Observatorium API as a data source.

Jaeger query exposes multiple HTTP and gRPC APIs. The Jaeger UI uses HTTP API. This API does not have OpenAPI schema, and it is only defined in [http_handler.go](https://github.com/jaegertracing/jaeger/blob/0dd3e2da0579caed9e24ad2782a1f638ad63214d/cmd/query/app/http_handler.go#L119). The API uses this [data model](https://github.com/jaegertracing/jaeger/blob/9cd7a7ec1aa43b24a8a970eb5b393ca2ffd98a5d/model/json/model.go#L52).

#### Expose Jaeger UI

In addition to exposing Jaeger query API the Observatorium could deploy Jaeger UI as well. At the moment Observatorium does not expose any UI and it is expected that users deploy UI locally.

## Alternatives

### Grafana Tempo

[Grafana Tempo](https://github.com/grafana/tempo) is an alternative multitenant storage that Observatorium project could use to store traces. The project is inspired and similar to the Prometheus TSDB, Thanos and Grafana Loki, therefore could it could be a good fit for Obeservatorium. The storage is split into two parts: 1. low retention requiring fast local disk (e.g. SSD) to store recent data and 2. object store (S3 compatible) to store historical data. Tempo is a distributed system with [multiple components](https://grafana.com/docs/tempo/latest/operations/architecture/): distributor, ingester, compactor, storage, querier. Distributor uses parts of the OpenTelemetry collector codebase to receive data in multiple formats (OTLP, Jaeger, Zipkin).

The project at the moment supports [search](https://grafana.com/docs/tempo/latest/getting-started/tempo-in-grafana/#tempo-search) only for recent data held in memory with retention about 15 minutes, therefore it is not suitable as a storage for Jaeger that exposes richer search API that supports search by attributes, time range and service name. Pull request [search feature](https://github.com/grafana/tempo/pull/1174) adds search capability for data in the storage, therefore the statement in this paragraph will be revisited once the feature is production ready.

On top of the above, the current tracing team proficiency is around Elasticsearch and Jaeger development. Given that it makes sense to postpone this alternative and revisit it later once there are more arguments supporting this case.

## Action Plan

Given the issues with the single Jaeger instance topology, the best approach is to deploy a Jaeger instance per tenant. At the moment this is not an issue due to low number (2-3) of tenants used in deployed instances of Observatorium (e.g. RHOBS). Once the [soft-multitenancy](https://github.com/jaegertracing/jaeger/issues/3427) is implemented in Jaeger the Observatorium should make use of it.

* Iterate and finalise this design document.
* Change deployment manifests for OpenTelemetry collector.
* Add deployment manifests for Jaeger.
* Define a process for adding a new tenant.

## References

* Jaeger: https://github.com/jaegertracing/jaeger
* Jaeger operator: https://github.com/jaegertracing/jaeger-operator
* Jaeger Query proto: https://github.com/jaegertracing/jaeger-idl/tree/master/proto
* Strimzi Kafka operator: https://operatorhub.io/operator/strimzi-kafka-operator
* OpenShift Elasticsearch operator: https://github.com/openshift/elasticsearch-operator
* Grafana Tempo: https://github.com/grafana/tempo
