# Jaeger traces store

* **Owners:**:
    * `@pavolloffay`

* **Related Tickets:**
    * [observatorium/issues/442](https://github.com/observatorium/observatorium/issues/442)

* **Other docs:**

## Table of Contents

- [TLDR](#tldr)
- [Why](#why)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [How](#how)
    * [Jaeger configuration](#jaeger-configuration)
        + [Single Jaeger instance](#single-jaeger-instance)
            - [OpenTelemetry collector configuration](#opentelemetry-collector-configuration)
        + [Multiple Jaeger instances](#multiple-jaeger-instances)
            - [OpenTelemetry collector configuration](#opentelemetry-collector-configuration)
    * [Jaeger Query](#jaeger-query)
        + [Expose Jaeger Query API](#expose-jaeger-query-api)
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

This document depends on the [Traces ingestion API and OpenTelemetry collector](https://github.com/observatorium/observatorium/pull/443) proposal.

This proposal adds Jaeger deployment with a persistent storage (e.g. Elasticsearch) to store trace data and it as well defines a query API to search for traces.

## Why

Traces store adds a crucial capability to Observatorium to persist and query tracing data.

## Goals

* Configure OpenTelemetry collector to export data to Jaeger collector.
* Deploy Jaeger with a persistent storage.
* Define query API for trace data.

## Non-Goals

* Define trace ingestion api - it is done in [Traces ingestion API and OpenTelemetry collector](https://github.com/observatorium/observatorium/pull/443) proposal.

## How

The how is split into the following sections:
* Jaeger configuration - describes how Jaeger components are deployed.
* Jaeger Query - describes how users access data stored in Jaeger.

### Jaeger configuration

Currently, the most adopted and supported storage in Jaeger is Elasticsearch. Red Hat also supports Jaeger with Elasticsearch (6.x) as part of the OpenShift product. Therefore, this proposal assumes Elasticsearch will be used as a storage backend. Distributed tracing team at Red Hat is looking at alternative storages hence the storage technology can change in the future.

Follows possible multitenant Jaeger deployment topologies. All deployment topologies assume that Jaeger will be deployed by the [Jaeger Operator](https://github.com/jaegertracing/jaeger-operator).

#### Single Jaeger instance

This deployment topology deploys a single Jaeger instance for all tenants with a single Elasticsearch instance. The data for each tenant would contain a unique label identifying a tenant. The label would be dynamically injected in the Observatorium API service or in the OpenTelemetry collector.

This deployment strategy will not work with the current Jaeger components. Here are few issues, and the list might not be complete:
* Service and operation name API. The service query API nor storage does not expose/store labels.
* Dependency/service architecture diagram. To support this feature dependency schema, query, collector and Spark aggregation would have to change.
* Data retention/TTL configurable pre tenant. The data retention in Elasticsearch is configurable per index.

##### OpenTelemetry collector configuration

The OpenTelemetry collector is configured with a single exporter to export data to a single Jaeger collector.

#### Multiple Jaeger instances

This deployment topology deploys a Jaeger instance per tenant and all Jaeger instances use a single Elasticsearch instance. The data for each tenant is stored in a separate Elasticsearch index. Jaeger was designed to separate tenant data per index. Therefore, all Jaeger functionality works well in this deployment strategy.

##### OpenTelemetry collector configuration

The OpenTelemetry collector is configured with an exporter per tenant that exports data to Jaeger collector allocated per a single tenant.

### Jaeger Query

There are two ways users can access data stored in Jaeger:
1. Expose Jaeger Query API in the Observatorium API
2. Expose Jaeger UI outside the cluster

#### Expose Jaeger Query API

Jaeger query API would be exposed in the Observatorium API. The users would have to deploy Jaeger UI externally and use Observatorium as a data source.

There are multiple Jaeger query APIs, in this case we propose to expose [api_v3](https://github.com/jaegertracing/jaeger-idl/tree/master/proto/api_v3). This API returns OpenTelemetry compatible payload.

#### Expose Jaeger UI

Jaeger Query with UI would be deployed in the Observatorium cluster and exposed to outside the cluster. Each tenant would get a unique URL to access Jaeger Query UI.

## Action Plan

* Iterate and finalise this design document.
* Change deployment manifests for OpenTelemetry collector.
* Add deployment manifests for Jaeger.
* Define a process for adding a new tenant.

## References

* Jaeger https://github.com/jaegertracing/jaeger
* Jaeger operator: https://github.com/jaegertracing/jaeger-operator
* Jaeger Query proto: https://github.com/jaegertracing/jaeger-idl/tree/master/proto
