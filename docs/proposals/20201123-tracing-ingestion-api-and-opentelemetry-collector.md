# Traces ingestion API and OpenTelemetry collector

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
  * [Expose tracing ingestion API in the Observatorium API service](#expose-tracing-ingestion-api-in-the-observatorium-api-service)
    + [Multitenancy](#multitenancy)
  * [Add deployments manifests for OpenTelemetry collector](#add-deployments-manifests-for-opentelemetry-collector)
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

This document proposes to expose tracing ingestion API  backed by the OpenTelemetry collector. This is a first step to support ingestion and storage for tracing data. The OpenTelemetry collector gives flexibility to forward tracing data to variety of tracing platforms, both on-prem or SaaS - Jaeger, Zipkin, Tempo, SigNoz or DataDog, Dynatrace etc.

## Why

At the moment Observatorium does not expose API to ingest distributed tracing data.

## Goals

* Expose API to ingest tracing data.
* Deploy a component that receives tracing data and is able to forward it to tracing platforms with storage.

## Non-Goals

* Store data in a persistent storage.
* Provide read API for tracing data.

## How

The how is split into two sections:
1. Expose tracing ingestion API
2. Add deployment manifests for OpenTelemetry collector

### Expose tracing ingestion API in the Observatorium API service

There are a couple of open-source tracing protocols out there - Zipkin, Jaeger and OpenTelemetry. At the moment the OpenTelemetry seems to be the most popular, and it is projected to have the biggest adoption in the future. Hence, using OpenTelemetry protocol seems to be the most appropriate choice.

The OpenTelemetry protocol primarily supports gRPC for sending traces (OTLP gRPC) with proto encoding. The HTTP protocol with protobuf encoding is supported as well, but the JSON encoding is still in experimental mode. The majority of users use OTLP gRPC and the HTTP is used in environments where gRPC cannot be used (e.g. mobile clients). Because the Observatorium already supports HTTP and the final state is to support both OTLP HTTP and gRPC, the Observatorium could initially support HTTP and in parallel start working on gRPC.

#### Multitenancy

Multitenancy could be handled the same way as it is handled for metrics and logs. The API service could add an HTTP header or attribute/label to data. This label would then be used in the collector to identify the tenant. Then the collector could route the data to a specific exporter per tenant or export the data to platform/storage by using a single exporter.

### Add deployments manifests for OpenTelemetry collector

The OpenTelemetry collector community already support deploying the collector via [OpenTelemetry operator](https://github.com/open-telemetry/opentelemetry-operator) or [HELM chart](https://github.com/open-telemetry/opentelemetry-helm-charts). A plain Kubernetes manifests could be as well used given the stateless nature and low complexity of the collector.

## Action Plan

* Iterate and finalise this design document.
* Add deployment manifests for OpenTelemetry collector.
* Add OTLP HTTP support to the Observatorium API.
* Add OTLP gRPC support to the Observatorium API.

## References

* OTLP HTTP: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md#otlphttp
* OTEL operator: https://github.com/open-telemetry/opentelemetry-operator
* OTEL Helm chart: https://github.com/open-telemetry/opentelemetry-helm-charts
* OTLP proto: https://github.com/open-telemetry/opentelemetry-proto/tree/main/opentelemetry/proto/trace/v1
