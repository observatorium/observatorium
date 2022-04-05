# Traces ingestion API and OpenTelemetry collector

* **Owners:**:
  * `@pavolloffay`

* **Related Tickets:**
  * [observatorium/issues/442](https://github.com/observatorium/observatorium/issues/442)

## Table of Contents

- [TLDR](#tldr)
- [Why](#why)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [How](#how)
  * [Expose tracing ingestion API in the Observatorium API service](#expose-tracing-ingestion-api-in-the-observatorium-api-service)
    + [Multitenancy](#multitenancy)
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

This document proposes to expose tracing ingestion API. This is a first step to support ingestion and storage for tracing data.

## Why

At the moment Observatorium does not expose API to ingest distributed tracing data.

## Goals

* Expose API in the Observatorium API service to ingest tracing data.

## Non-Goals

* Store data in a persistent storage.
* Provide read API for tracing data.

## How

### Expose tracing ingestion API in the Observatorium API service

There are a couple of open-source tracing protocols out there - Zipkin, Jaeger and OpenTelemetry. At the moment the OpenTelemetry seems to be the [most popular](https://opentelemetry.io/vendors/), and it is projected to have the biggest adoption in the future. Hence, using OpenTelemetry protocol seems to be the most appropriate choice.

The OpenTelemetry protocol primarily supports gRPC for sending traces (OTLP gRPC) with proto encoding. The HTTP protocol with protobuf encoding is supported as well, but the JSON encoding is still in experimental mode. The majority of users use OTLP gRPC and the HTTP is used in environments where gRPC cannot be used (e.g. mobile clients).

Observatorium should expose OTLP gRPC protocol and later add support for OTLP HTTP if there will be use-case to require it.

### Multitenancy

Multitenancy will be handled the same way as it is handled for metrics and logs. The API service will add an HTTP header or attribute/label to data. This label would then be used in the collector to identify the tenant and store data accordingly.

## Action Plan

* Iterate and finalise this design document.
* Add OTLP gRPC support to the Observatorium API.

## References

* OTLP HTTP: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md#otlphttp
* OTEL operator: https://github.com/open-telemetry/opentelemetry-operator
* OTEL Helm chart: https://github.com/open-telemetry/opentelemetry-helm-charts
* OTLP proto: https://github.com/open-telemetry/opentelemetry-proto/tree/main/opentelemetry/proto/trace/v1
* OpenTelemetry vendors: https://opentelemetry.io/vendors/
* OpenTelemetry adopters: https://github.com/open-telemetry/community/blob/main/ADOPTERS.md
* OpenTelemetry is the second most active CNCF project: https://twitter.com/smflanders/status/1262401649248739331
