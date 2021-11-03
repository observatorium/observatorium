# Rules API Proposal

* **Owners:**:
  * `@squat` / `@ianbillett`

* **Related Tickets:**
  * `<JIRA, GH Issues>`

* **Other docs:**
  * Original [Prometheus Rules in Observatorium API design doc](https://docs.google.com/document/d/1F9Cw6I4qFs__0Dcm19xvxJqRBCAGQBBggYg_PQSV_-g/edit#heading=h.cp0jmcyfj3) (internal Red Hat link)

## Table of Contents

- [TLDR](#tldr)
- [Why](#why)
  * [Pitfalls of the current solution](#pitfalls-of-the-current-solution)
    + [Implicit deploy cycle dependency](#implicit-deploy-cycle-dependency)
    + [Shared configuration repositories](#shared-configuration-repositories)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [How](#how)
  * [How does Rule get tenant recording and alerting rules](#how-does-rule-get-tenant-recording-and-alerting-rules)
  * [How do we store tenant rules](#how-do-we-store-tenant-rules)
    + [Service](#service)
    + [Storage layer](#storage-layer)
      - [1. ETCD](#1-etcd)
      - [2. Object Storage](#2-object-storage)
      - [3. RDBMS](#3-rdbms)
      - [Storage Decision](#storage-decision)
  * [Sequence Diagram](#sequence-diagram)
    + [1. Store Tenant Rules](#1-store-tenant-rules)
    + [2. Sync Rules to Rule](#2-sync-rules-to-rule)
    + [3. Recording Rule Query](#3-recording-rule-query)
- [Alternatives](#alternatives)
  * [More frequent deploys](#more-frequent-deploys)
  * [Per-tenant GitOps](#per-tenant-gitops)
- [Action Plan](#action-plan)

<small>

<i>

<a href="http://ecotrust-canada.github.io/markdown-toc/">
Table of contents generated with markdown-toc
</a>

</i>

</small>

## TLDR

We propose implementing a multi-tenant API that allows tenants to create, read, update and delete Prometheus recording and alerting rules.

## Why

The single biggest source of frustration from internal Red Hat tenants of Observatorium is the time it takes for changes to their recording and/or alerting rules to appear in our production infrastructure.

### Pitfalls of the current solution

#### Implicit deploy cycle dependency

Currently, the only method of modifying Prometheus recording and/or alerting rules in Observatorium is via the jsonnet definition in our repository. This implicitly ties an update of the rule configuration to a rollout of the service. This implicitly ties the deploy-cycle of Prometheus rules to the deploy-cycle of Observatorium itself. If a new set of rules needs to be deployed, our team is required to roll-out our production infrastructure.

As the number of tenants we support and tenants we serve increases, the team responsible for the Observatorium installation is required to roll out production more frequently in order to satisfy tenant requests. This becomes an increasing impediment to tenant experience as the size of Observatorium increases.

#### Shared configuration repositories

For rules to be rolled out, tenants need the ability to raise pull requests against the Observatorium repositories that contain our source configuration. This works fine when the two teams co-exist within the same organisation, but it becomes a blocking issue when the two teams share a much more restricted (e.g. inter-company) trust boundary.

This is not a blocker for Red Hat's internal offering, however this is not inline with Observatorium's [stated goal](https://observatorium.io/docs/usage/getting-started.md/#whats-observatorium) of offering a SaaS-like monitoring solution.

## Goals

* Tenants can update their Prometheus recording/alerting rules without intervention from the operating team.
* Observatorium must be able to scale horizontally with the number of tenants.

## Non-Goals

* Rate-limiting or tenant resource accounting.
* Logging rules. While Loki supports the same format rules as Prometheus, we will explicitly not consider supporting these in this proposal.

## How

We propose to solve the above problem by implementing a multi-tenant API that allows tenants to create, read, update and delete Prometheus recording and alerting rules.

In Observatorium, [Thanos Rule](https://thanos.io/tip/components/rule.md/) evaluates recording and alerting rules. These are configured using a local file defined by Rule's `--rule.file` flag. `Rule` is configured as one of `Query`'s stores, so when rules are evaluated they are available to be queried.

This leads to two problems we need to solve:
1. How does `Rule` obtain tenant recording and alerting rules?
2. How do we store tenant recording and alerting rules?

### How does Rule get tenant recording and alerting rules

For this step, we will leverage the [thanos-ruler-syncer](https://github.com/observatorium/thanos-rule-syncer).

Periodically, this application does three things:
1. Calls the Observatorium API to retrieve tenant rules.
2. Writes these rules to a location defined by the `-file` flag.
3. Asks Rule to reload its rules by calling the `/-/reload` endpoint.

This will run as a sidecar to `Rule` and `-file` will be shared between the two containers.

### How do we store tenant rules

#### Service

We will implement the rule storage backend as a separate service within Observatorium (i.e. not part of the API):
* This follows the established pattern of specifying separate functional backends in the API.
* The API stays as a super-performant gateway that handles authentication and authorization but no application-specific logic (in the spirit of the [Unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy))
* Observatorium users are free to specify different storage backends, or no rule backend at all.

The rule backend will be specified via the `--metrics.rules.endpoint` flag, and will route requests via the `/api/metrics/v1/{tenant}/rules` endpoint.

NB: By scoping rules to the `/metrics` endpoint we are consciously creating room for future logging rules.

Ideally, we would define the rule storage backend API in OpenAPI format, so that it can easily be consumed by downstream services.

#### Storage layer

We will use a persistent storage mechanism that is external to the API deployment (i.e. no local storage). This enables us to satisfy our requirement of horizontally scaling the rule storage backend.

While we have explicitly de-coupled the rule storage backend from the API, we require a concrete implementaion.

We have a number of options for the storage layer backing the rules storage backend service.

##### 1. ETCD

One option would be to use the Kubernetes' control plane as the backing data store for tenant's recording and alerting rules.

Pros:
* Lowest possible overhead. No external storage is required to be orchestrated.
* ETCD data model well suited for recording and alerting rules i.e. YAML.
* Storage permissions are managed as Kubernetes objects.
* We could leverage Recording Rules CRDs and get validation for free (?)

Cons:
* Implicitly ties Observatorium's implementation into a Kubernetes environment. Hard to back out.
* Large blast radius. With a large number of tenants and recording and alerting rules, we could un-intentionally impact the Kubernetes control plane's performance.
* ETCD has a limit of 1.5MB per key / value pair so we can't store large files (but Prometheus operator seems to do this ok?).

##### 2. Object Storage

Pros:
* Users of Thanos project are likely to already have object storage configured and ready to use.
* Recording and alerting rule data model is a good fit for object storage i.e. YAML.

Cons:
* Unclear consistency semantics depending on object storage provider i.e. in the case of lots of writes, who wins?
* Requires tenants to manage another set of secret credentials.

##### 3. RDBMS

Another option we have considered is to provision and use a relational database i.e. Postgres.

Pros:
* Strong isolation and update semantics.
* Enables us to use multi-dimensional

Cons:
* We don't currently run RDBMS, and don't have any experience doing so.
* Significant infrastructure overhead - tenants required to BYO database.
* Requires tenants to manage another set of secret credentials.

##### Storage Decision

This was discussed in the [Observatorium Community Meeting](https://docs.google.com/document/d/1jLvOH0Lllt-ShXVgWeeHBGPb3ZuagirabD_hbq41EKs/edit#heading=h.wxenjcrf4iee) on 2021-07-31.

The consensus among the group was that object storage offered us the best combination of ease of use, familiarity for existing users, and also compatability with the required operations against tenants rules.

We discussed that we will be likely to employ a RDBMS relatively soon for other features (i.e. RBAC), but we decided to cross that bridge when we came to it.

### Sequence Diagram

How will this all work in practice?

![sequence-diagram](../assets/recording-rules-api.svg)

#### 1. Store Tenant Rules
1. Tenant sends request to the API using the path `/api/metrics/v1/{tenant}/rules` containing their rules data.
2. API performs authentication and authorization of the tenant then forwards the request to the rules storage backend.
3. Rules storage backend performs validation of the payload and stores the rule file in object storage at the path `metrics/rules/{tenant}/{file_name}.yaml`.
4. Object storage returns success to the Rule Backend.
5. Rule backend returns success to the API.
6. API returns success to the Tenant.

#### 2. Sync Rules to Rule
1. Periodically, thanos-ruler-syncer queries the API for all rules for all tenants.
   * In a trusted environment (i.e. with a kubernetes cluster) thanos-ruler-syncer could call the rule-storage-backend directly. Here we demonstrate the lower-trust environment and force thanos-ruler-syncer to authenticate via the API.
   * Do we need to expose an endpoint to return all of the rules? Unclear at the moment.
   * NB: We can definitely make some optimisations here. Suggestions welcome.
2. API requests data from the rule storage backend.
3. Rule storage backend requests data from Object storage.
4. Object storage returns data to the Rule storage backend.
5. Rule storage backend returns data to the API.
6. API returns data to thanos-ruler-syncer.
7. Thanos-ruler-syncer combines all of the data into one file, and writes this file to a directory that is shared with Thanos Rule.
8. Thanos-ruler-syncer calls Thanos Rule's `/-/reload` endpoint to reload newly created configuration.
9. Thanos Rule begins to evaluate the recording and alerting rules and stores the results in its local TSDB instance.

#### 3. Recording Rule Query

1. Tenant makes a query request to the API containing a recording or alerting rule metric
2. API performs authentication and authorization of the request then queries the data from Thanos Query.
3. Query has Rule configured as one of its Stores, and thus requests the data from Rule.
4. Rule returns the metric value of the recording or alerting rule to Query.
5. Query returns the result to API.
6. API returns the result to the tenant.

## Alternatives

### More frequent deploys

The above problem has manifested because our deploys are too infrequent for our tenants. Could we not use this as an opportunity to optimize the velocity at which we deploy to production? That way we have the benefit of solving the near-term problems of our tenants, and also rolling things out more frequently.

As it stands, any rollout to production requires coordination and management from a member of the operating team. As the number of tenants and tenants becomes larger, the operating team becomes more of a bottleneck to common tenant operations.

Our goal here is to remove our team entirely from the control loop of updating Prometheus rules. We are trading off immediate term technical complexity against longer-term scalability and happier multi-tenancy operation.

This option also does not address the trust-boundary issue identified above.

### Per-tenant GitOps

Another option would be for each tenant to commit their rule configuration into a git repository they control, and then we synchronise that into the right place in our infrastructure based on a URL they provide. This would eliminate the need for us to manage state on their behalf.

This is a neat idea that avoids deploy dependencies, but has some drawbacks:
* Validation - The tenant's rules would only be validated by us in the Ruler itself, meaning mis-configured rules do not get surfaced to tenants. We could encourage tenants to validate
* Secrets - We can assume that tenants may not want their production rules and alerts exposed to the world, this would then require us to exchange and store secrets with them to access their repository.
* Overheads - Managing an additional git repo & associated secrets imposes extra overheads on the tenants.

The overhead and complexity of managing sensitive tenant secrets is out-of-scope for this project.

## Action Plan

* Iterate and finalise this design document.
* Implement proof-of-concept Rules API [#138](https://github.com/observatorium/api/pull/138).
* Make the API available to tenants in staging.
* Promote API to production after ~1 month of running in staging.
