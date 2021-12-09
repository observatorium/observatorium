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
        + [Single Jaeger instance](#single-jaeger-instance)
            - [OpenTelemetry collector configuration](#opentelemetry-collector-configuration)
        + [Jaeger instance per tenant](#jaeger-instance-per-tenant)
            - [OpenTelemetry collector configuration](#opentelemetry-collector-configuration)
    * [Jaeger Query](#jaeger-query)
        + [Expose only Jaeger Query API](#expose-only-jaeger-query-api)
        + [Expose Jaeger Query UI](#expose-jaeger-ui)
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

This proposal adds Jaeger deployment with a persistent storage (e.g. Elasticsearch) to store trace data and it as well defines a query API to search for traces.

## Why

Traces store adds a crucial capability to Observatorium to persist and query tracing data.

## Goals

* Configure OpenTelemetry collector to export data to Jaeger collector.
* Allow Observatorium to store and retrieve traces.
* Define query API for trace data.

## Non-Goals

* Define trace ingestion API - it is done in [Traces ingestion API and OpenTelemetry collector](https://github.com/observatorium/observatorium/pull/443) proposal.

## How

The how is split into the following sections:
* Jaeger configuration - describes how Jaeger components are deployed.
* Jaeger Query - describes how users access data stored in Jaeger.

### Jaeger configuration

Currently, the most adopted and supported storage in Jaeger is Elasticsearch. Red Hat also supports Jaeger with Elasticsearch (6.x) as part of the OpenShift product. Therefore, this proposal assumes Elasticsearch will be used as a storage backend. Distributed tracing team at Red Hat is looking at alternative storages hence the storage technology can change in the future.

Follows possible multitenant Jaeger deployment topologies. All deployment topologies assume that Jaeger will be deployed by the [Jaeger Operator](https://github.com/jaegertracing/jaeger-operator) which depends on [OpenShift Elasticserach operator](https://github.com/openshift/elasticsearch-operator) and [Strimzi operator](https://operatorhub.io/operator/strimzi-kafka-operator) if Kafka is used. Jaeger can be as well configured to use externally managed Elasticsearch or Kafka instances.

#### Single Jaeger instance

This deployment topology deploys a single Jaeger instance for all tenants with a single Elasticsearch instance. The data for each tenant would contain a unique attribute to identify a tenant. The label would be dynamically injected in the Observatorium API service or in the OpenTelemetry collector.

This deployment strategy will not work with the current Jaeger components. Here are few issues, and the list might not be complete:
* Service and operation name API. The service query API nor storage does not expose/store labels.
* Dependency/service architecture diagram. To support this feature dependency schema, query, collector and Spark aggregation would have to change.
* Data retention/TTL configurable pre tenant. The data retention in Elasticsearch is configurable per index.

Proposal to add soft multitenancy to Jaeger - [jaegertracing/jaeger/3427](https://github.com/jaegertracing/jaeger/issues/3427)

##### OpenTelemetry collector configuration

The OpenTelemetry collector is configured with a single exporter to export data to a single Jaeger collector.

#### Jaeger instance per tenant

This deployment topology deploys a Jaeger instance per tenant and all Jaeger instances use a single Elasticsearch instance. The data for each tenant is stored in a separate Elasticsearch index. Jaeger was designed to separate tenant data per index. Therefore, all Jaeger functionality works well with this deployment strategy.

##### OpenTelemetry collector configuration

The OpenTelemetry collector is configured with an exporter per tenant that exports data to Jaeger collector allocated per a single tenant.

### Jaeger Query

There are two ways users could access data stored in Jaeger:
1. Expose only Jaeger Query API in the Observatorium API.
2. Expose Jaeger UI outside the cluster.

#### Expose only Jaeger Query API

Jaeger query API would be exposed in the Observatorium API and users would have to deploy Jaeger UI and use Observatorium API as a data source.

Jaeger query exposes multiple HTTP and gRPC APIs. The Jaeger UI uses HTTP API. This API does not have OpenAPI schema, and it is only defined in [http_handler.go](https://github.com/jaegertracing/jaeger/blob/0dd3e2da0579caed9e24ad2782a1f638ad63214d/cmd/query/app/http_handler.go#L119). The API uses this [data model](https://github.com/jaegertracing/jaeger/blob/9cd7a7ec1aa43b24a8a970eb5b393ca2ffd98a5d/model/json/model.go#L52).

#### Expose Jaeger UI

In addition to exposing Jaeger query API the Observatorium could deploy Jaeger UI as well. At the moment Observatorium does not expose any UI and it is expected that users deploy UI locally.

## Action Plan

Given the issues with the single Jaeger instance topology, the best approach is to deploy a Jaeger instance per tenant that separates tenant's data by using different indices. This approach requires adding new Jaeger instance if a new tenant is added to the system.

* Iterate and finalise this design document.
* Change deployment manifests for OpenTelemetry collector.
* Add deployment manifests for Jaeger.
* Define a process for adding a new tenant.

## References

* Jaeger https://github.com/jaegertracing/jaeger
* Jaeger operator: https://github.com/jaegertracing/jaeger-operator
* Jaeger Query proto: https://github.com/jaegertracing/jaeger-idl/tree/master/proto
* Strimzi Kafka operator: https://operatorhub.io/operator/strimzi-kafka-operator
* OpenShift Elasticsearch operator: https://github.com/openshift/elasticsearch-operator
