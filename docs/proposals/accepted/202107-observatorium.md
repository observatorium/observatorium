# Observatorium

Initial Observatorium Design Proposal

* **Owners:**
  * Observatorium Maintainers

* **Other docs:**
  * RH internal: [1](https://docs.google.com/document/d/1lX0Tl77NFp9m1ZhV3ya1iOQSLZUI0fYbliTt0H_WGJA/edit), [2](https://docs.google.com/document/d/15MZfUPXvOMc9rUlXOVmveQMgiWVzgzNsRwxR8tZmx4s/edit)

## TL;DR

Observatorium brings multiple backends for each observability signal into one consistent and comprehensive platform. It allows the easier deployment and operational model with consistent APIs and tenancy for metrics, logging, tracing and more.

The idea is to develop an easy to operate stack that can run as a SaaS offering and for on-premise and self-service installations.

## Why

Undoubtedly, observability is critical for developing and running software, especially given the Cloud Native application's complexities and scale. Traditional Observability signals like metrics, logging, tracing, and profiles are key to enable monitoring (with alerting), debugging, billing, incident handling, reporting, analytics, and more. All that is required to make data-driven decisions.

Each observability signal requires a different type of storage and APIs for efficiency purposes. That's why we have many projects focused on each signal, e.g. Thanos, Prometheus and Cortex for metrics; Loki, Elasticsearch for logging and Tempo, Jaeger + Elasticsearch/Cassandra for tracing. This is required but there are many lessons we learned that can work across all signals:

* Schemaless metadata enables innovation and feature reusability. Prometheus and Loki data model relies on labels, which are just string names and values, no other type. Admins can build whatever schema they want on top of this model, yet it brings enormous simplification to the underlying system.
* Object Storage as the storage it's often undervalued. It's cheap, available everywhere and has easy APIs. We believe you can build all monitoring stories on top of it, which allows a cheap and easy operational model (no NoSQL/Postgres to maintain and scale!)
* Multi-tenancy can be often built based on data labels, but cross-tenancy and cross instance federation is often required to gather data from multiple backend instances/locations.

There is a power in embedding storage that allows us to do those things under a single umbrella and have a unifying API in front. Plus one of the biggest benefits of observability is when we [deploy multiple signals together into one coherent and consistent system](https://www.bwplotka.dev/2021/correlations-exemplars/).

## Goals

**Target Audience**: Personas who want to provide observability services for monitoring, debugging and SLOs assurance for themselves, their teams or even SaaS customers.

The goal of Observatorium is to be able to deploy an observability system that has the following characteristics:

* Open Source first
  * Apache v2 license, free to use, all (in-scope) contributions welcome.
* Ability to scrape/ingest (write), store, query (read), and retrieve metrics, logs and traces from multiple clusters and tenants, using consistent APIs:
  * Metrics
    * Write: Prometheus Remote Write API (replication model)
    * Read: Compatible with the Prometheus/Thanos query and metadata APIs (e.g. query, query_range, series, labels), making it possible to use any dashboard solution (e.g. Grafana).
  * Logs:  (Loki API)
  * Tracing: Not defined
* Support recording rules and alerting based on data from multiple clusters and tenants (and signals).
  * Allow users to configure alerts and recording rules with immediate effect (minutes)
* Opinionated installation and deployment model for Kubernetes using Operator, OLM and jsonnet templates.
* Support multi-tenancy for all APIs allowing to choose between soft and hard tenancy.
  * Multiple tenants can access the same metrics (cross tenancy)
* Horizontally Scalable: supporting smaller as well as larger scale.
  * It should be possible to deploy the solution in small environments such as single-node clusters.
  * Scaling up is adding more instances rather than throwing more resources at the same instance.
* Relatively cheap, easy to operate, and ubiquitous storage even for long retention:
  * Object storage is the main storage solution for longer retention.
  * Block storage is only for short term data or local buffers.
* Ability to define data retention policies per tenant and signals.
* Control over the granularity of the data access.
  * Allowing a tenant to read a subset of data.
  * Allowing a tenant to send a subset of the data off the cluster.
* Secure by default
  * All external-facing APIs support TLS for encryption
  * All external-facing APIs support OpenID for authentication.
  * All external-facing APIs support [OPA](https://www.openpolicyagent.org/) for authorization.
* High availability and solid failover handling.
  * For example, for the right path, the client-side can survive the unavailability of a central cluster for e.g. 30m without data loss.
* Built-in mechanisms to support production setups out of the box.
  * Rate-limiting.
  * Retries.
  * Sane defaults for metrics ingestion (remote write configuration, sharding, queues).

## Soft Goals

Currently not in scope due to time constraints, but possible if you want to contribute. Those all can be build on top of Observatorium as extensions.

* Ability to maintain alert routing (e.g. Alertmanager).
  * Alerts API:
    * Alertmanager v2 API
* APIs for managing Rules (self-service)
* APIs for managing Tenancy (self-service)
* The ability for users to manage dashboards (e.g. Grafana)
* Provide a single access point with access to all signals
  *  Discovery
* Support per tenant API quota and usage reports (e.g. show-back, billing).
* Support cold and hot storage.
* Ability to ingest (write), store, query (read), and visualize traces.
* Ability to ingest (write), store, query (read), and visualize profiles.
* Standard out of box alerts and monitoring rules for the components that Observatorium will deploy
* Support Blackbox Probes.
* Support [OLAP](https://en.wikipedia.org/wiki/Online_analytical_processing) Analytics Queries.
* Ability to correlate signals and quickly navigate from one to another.
* Ability to ingest metrics via pure push method (push gateway, OLTP metrics protocol).
* Deploy the components on non-Kubernetes platforms.
* Pluggability: Supporting more than one project per signal

## Non-Goals

Currently not in scope due to project complexity, but we may revisit:

* Extensibility in Authn/Authz protocols.
* Non-observability Events.
* Collection and Instrumentation pipeline (e.g. things that OpenTelemetry. Prometheus, Prometheus Operator, [MonitoringStack Operator](https://github.com/rhobs/monitoring-stack-operator) etc. are focusing on)

## How

TBD

## Alternatives

1. Don't do it. Prometheus with Operator, Thanos, Loki, Jaeger, Tempo, Grafana exist, why not deploy and configure independently?

That's a very valid point. You can already, and those projects already collaborate. However, to have a solid infrastructure and observability using those, you need to spend many months of engineering time in understanding:

*   **What project to choose for each signal?**

Investing in a single project takes time and effort, so you want to choose correctly. But how to choose? In Observatorium, we chose for you based on Red Hat and open source community experience and the best of our knowledge.

*   **How to deploy it using a single way? **

    There is no single way. Each of the projects prefers different things. For example, one prefers Ansible or Operator model more, second Jsonnet, other Helm. And you probably use something different. In most cases, your organization has to reinvent all of it, and it takes time and huge knowledge to develop and maintain.

*   **Where to store data?**

Object Storage, Disk, Other DB?

*   **How to operate and run it?**

What are the alerting rules to deploy, what performance limits to set, what are the dashboards?

*   **How to have a unified API, rate limiting, quota, billing, and auth story?**

    Each project can use different APIs, protocols and most likely enforce adding your auth, rate limit, etc., layers. Observatorium adds all of this.

*   **Lots of data is duplicated. How can we reduce running costs?**

    In many cases, the data ingested by such an observability system is duplicated like, e.g. Prometheus, Loki, and Jaeger. All of them have to store the index up to the workload container. What if we can reuse and store only one not three, replicas of this information? Observatorium in principle, would allow such a solution.

Currently, we are aware of one open-source alternative: [opstrace](https://github.com/opstrace/opstrace). It was announced recently by ex-Red Hat employees, and we don't know much about it. From a quick look, it provides Cortex and Loki operators written in TypeScript. The community is very fresh as well. It sounds like we could collaborate on some aspects, but Observatorium aims for a more flexible deployment model. That's why the Thanos project was chosen for metrics. Observatorium is also very deeply rooted in all open source projects it uses, allowing smooth experience and community efforts.

2. Focus on a single Observability signal instead of many e.g. rely on traces only.

There is a huge debate ([example](https://twitter.com/el_bhs/status/1349406398388400128)) in open source what Observability signal is the most important and which one we can drop to reduce the complexity and cost of achieving observability goals. The popular statement is to trace everything and deduce metrics or even logs from traces. Or log only and deduce metrics and traces from log lines. All of those are novel proposals, but we believe in **no fit-all solution**. From our experience, healthy infrastructure should have, first of all, a metric collection pipeline for real-time debugging and reliable alerting. Then, only if metrics are achievable, which solves ~90% of your monitoring and observability use cases, one can focus on installing logging infrastructure, which solves further ~90% of the remaining 10% of use cases. Only if that is met, tracing and profiling might help for more advanced use cases and might be required for the remaining ~1% of Observability cases. So does it make sense to focus on tracing only? It might make sense for some, might be over-engineering and over-expensive for others.

In our opinion, the biggest value is in a system that gathers all metrics, logging, tracing, and profiles (if you choose to) and allows for a correlation between those.

3. Don't do it. Again, paid, close source solutions exist.

Many players are aiming to solve similar goals, but in the close source: Tanzai, Splunk, DataDog, ElasticCloud, StackDriver (GCP), CloudWatch (AWS), AWS Prometheus Managed, Grafana Enterprise, Logz.io, Honeycomb.io (there are more, those are just examples). Why investing time in contributing, developing, and running Observatorium if there are good paid services?

There are a few reasons:

*   **Cost: **

All of the above are paid. Of course, running Observatorium is far from free. However, it allows you to make the decision if you want to run yourself or allow others to manage Observatorium as SaaS for you. (While currently, no SaaS option emerged, because we announced the project now, we expect many will emerge). Additionally, one of the Observatorium's main goals is to limit operational (maintenance) and running costs to a minimum.

*   **Missing features or relying on one signal: **

Most of the mentioned solutions focus only on the part of the observability story. This is great, however, it might mean inconsistency and further complexity if you have to pay for logging to service X, use metrics with project Y, and pay to service Z to have traced. See

*   **Lock-in due to close source: **

Most of those solutions are top-notch and state of the engineering art (e.g. Monarch used in part of StackDriver). However, they require tons of engineering knowledge, so it's hard to use even if installable on-premise.

Since it's a closed source, you can't help in maintaining this. Of course, you pay for support to NOT do it, but that does not mean a solution to all problems as SaaS teams has also limited resources. Again, not mentioning the inability to assess the solution viability for your use cases properly. Last but not least proprietary code usually means proprietary standards and data formats. This means huge problems in extendability and future integrations.

This is why we believe in open source code and development and it's why the whole Observatorium project is fully open source with open governance.

*   **Limitation due to Saas model**:

Because someone has to manage a service for you, it's extremely hard to control and run software within your data center or network. Operators help, but in the end, this also needs a huge amount of attention and control which you don't want on the client-side. Ideally, you want to run as simple a stack as possible. This is why the **only** way to solve observability using those services is to ship data as soon as possible to remote endpoints. While it might be the best solution for some cases, it might be an over-engineering and too big a trade-off for others. We found that many use cases, especially reliable monitoring, can be satisfied while keeping the majority of all the data within a single cluster or network infrastructure, which improves reliability, reduces cost and complexity.

Observatorium aims to allow both SaaS or single network cases (or a mix!)
