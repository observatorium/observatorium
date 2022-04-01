![Observatorium](logo/Logo-Observatorium-Full.png)

[![Build Status](https://circleci.com/gh/observatorium/observatorium.svg?style=svg)](https://circleci.com/gh/observatorium/observatorium) [![Slack](https://img.shields.io/badge/join%20slack-%23observatorium-brightgreen.svg)](https://slack.cncf.io/)

### Configuration for Multi-Tenant, Flexible, Scalable, Observability Backend

Observatorium allows you to run and operate effectively a multi-tenant, easy to operate, scalable open source observability system on Kubernetes. This system will allow you to ingest, store and use common observability signals like metrics, logging and tracing. Observatorium is a "meta project" allows you to manage, integrate and combine multiple well-established existing projects like [Thanos](https://thanos.io), Loki, Tempo/Jaeger, Open Policy Agent etc under a single consistent system with well-defined tenancy APIs and signal correlation capabilities.

As active maintainers and contributors to the underlying projects, we created a reference configuration, with extra software that connects those open source solutions into one unified and easy to use service. It adds missing gaps between those projects like consistency, multi-tenancy, security and resiliency pieces that are needed for a robust backend.

![](/docs/design/Observatorium-High-Level.png)

Read more on [High Level Architecture](/docs/design/architecture.md) docs.

### Context

As the Red Hat Monitoring Team, we were focusing on the Observability software and concepts since the CoreOS acquisition. From the beginning, one of our main goals was to establish a stable in-cluster metric collection, querying, and alerting for OpenShift clusters. With the growth of managed OpenShift (OSD) clusters, the scope of the team goal has extended: we had to develop a scalable, global, metric stack that can be run in local as well as a central location for monitoring and telemetry purposes. We also worked together with Red Hat Logging and Tracing teams to implement something similar for logging and tracing. We’re also working on Continuous Profiling aspects.

From the very beginning our teams were leveraging Open Source to accomplish all those goals. We believe that working with the communities is the best way to have long term, successful systems, share knowledge and establish solid APIs. You might have not seen us, but members of our teams have been actively maintaining and contributing to major Open Source standards and projects like Prometheus, Thanos, Loki, Grafana, kube-state-metrics (KSM), prometheus-operator, kube-prometheus, Alertmanager, cluster-monitoring-operator (CMO), OpenMetrics, Jaeger, ConProf, Cortex, SIG CNCF Observability, SIG K8s Instrumentation and more.

## What's Included

* [Observatorium](https://github.com/observatorium/observatorium) is primarily defined in [jsonnet](https://jsonnet.org/), which allows great flexibility and reusability. The main configuration resources are stored in [components](https://github.com/observatorium/observatorium/tree/main/configuration/components) directory, and they import further official resources like [kube-thanos](https://github.com/thanos-io/kube-thanos). Some Examples:
  * You can see examples of how it can be used in different variations/environments [here](https://github.com/observatorium/observatorium/tree/main/configuration/examples).
  * Our [Red Hat Observability Service](https://github.com/rhobs/configuration) is also build on Observatorium.

* We are aware that not everybody speaks jsonnet, and not everybody have it's own GitOps pipeline, so we designed alternative deployments based on the main jsonnet resources. [Operator](https://github.com/observatorium/operator) project delivers Kubernetes plain Operator that operates Observatorium.

> NOTE: Observatorium is set of cloud native, mostly stateless components that mostly does not special operating logic. For those operations that required automation, specialized controllers were designed. Use Operator only if this is your primary installation logic or if you don't have CI pipeline.

> NOTE2: Operator is in heavy progress. There are already plans to streamline its usage and redesign current CustomResourceDefinition in next version. Yet, it's currently used in production by many bigger users, so any changes will be done with care.

* The [Thanos Receive Controller](https://github.com/observatorium/thanos-receive-controller) is a Kubernetes controller written in Go that distributes essential tenancy configuration to the desired pods.

* [The `API`](https://github.com/observatorium/api) is the facet of Observatorium service. It's a lightweight proxy written in Go that helps with multi-tenancy, tenancy (isolation, cross tenancy requests, rate-limiting, roles, tracing). This proxy should be used for all external traffic with Observatorium.

* [OPA-AMS](https://github.com/observatorium/opa-ams) is our Go library for integrating Open Policy Agent with Red Hat authorization service for smooth OpenShift experience.

* [up](https://github.com/observatorium/up) is a useful Go service that periodically queries Observatorium and outputs vital metrics on the Observatorium read path healthiness and performance over time.

* [token-refresher](https://github.com/observatorium/token-refresher) is a simple Go CLI allowing to perform OIDC refresh flow.

### Getting Started

* See [Getting Started](docs/usage/getting-started.md) and our Katacoda Tutorials: https://katacoda.com/observatorium
* See [Contributing](docs/community/README.md) on how to start contributing.

### Status: Work In Progress

While metric and logging part using Thanos and Loki is used in production at Red Hat,documentation, full design, user guides, different configurations support are in progress.

Stay Tuned!

### Missing something or not sure?

Let us know! Visit our Slack channel or put a GitHub issue!
