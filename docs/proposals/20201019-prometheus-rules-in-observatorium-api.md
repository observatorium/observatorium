## Implement API for managing Prometheus Recording Rules

* **Owners:**:
  * `@squat` / `@ianbillett`

* **Related Tickets:**
  * `<JIRA, GH Issues>`

* **Other docs:**
  * Original [Prometheus Rules in Observatorium API design doc](https://docs.google.com/document/d/1F9Cw6I4qFs__0Dcm19xvxJqRBCAGQBBggYg_PQSV_-g/edit#heading=h.cp0jmcyfj3) (internal Red Hat link)

## TL;DR

We propose implementing a multi-tenant API that allows users to create, read, update and delete Prometheus recording rules.

## Why

The single biggest source of frustration from internal Red Hat users of Observatorium right now is the time it takes for changes to their recording rules to appear in our production infrastructure.

### Pitfalls of the current solution

#### Implicit deploy cycle dependency

Currently, the only method of modifying Prometheus recording rules in Observatorium is via the jsonnet definition in our repository. This implicitly ties an update of the rule configuration to a rollout of the entire infrastructure. This implicitly ties the deploy-cycle of Prometheus rules to the deploy-cycle of Observatorium itself. If a new set of rules needs to be deployed, our team is required to roll-out our production infrastructure.

As the number of tenants we support and users we serve increases, the team responsible for the Observatorium installation is required to roll out production more frequently in order to satisfy user requests. This becomes an increasing impediment to user experience as the size of Observatorium increases.

#### Shared configuration repositories

For rules to be rolled out, users need the ability to raise pull requests against the Observatorium repositories that contain our source configuration. This works fine when the two teams co-exist within the same organisation, but it becomes a blocking issue when the two teams share a much more restricted (e.g. inter-company) trust boundary.

This is not a blocker for Red Hat's internal offering, however this is not inline with Observatorium's stated goal (TODO: insert link) of offering a SaaS-like monitoring solution.

## Goals

* Users can update their Prometheus recording rules without intervention from the operating team.
* The Observatorium API must be able to scale horizontally scale with the number of users.

## Non-Goals

* Rate-limiting or accounting of user resources.
* Define or implement the underlying storage technology.

## How

We propose to solve the above problem by implementing a multi-tenant API that allows users to create, read, update and delete Prometheus alerting and recording rules.

Thanos Ruler operates on the rule configuration provided by users, which is consumed from a file defined by Ruler's `--rule.file` flag. Therefore, this configuration that we ingest via the API needs to be replicated into the local data directory of the Thanos Ruler instances. This is achieved with the [thanos-ruler-syncer](https://github.com/observatorium/thanos-rule-syncer), an application that will call the API above and write the results into a location that Thanos Rule can access.

### How should we persist user's rule configuration?

For the Observatorium API to satisfy our requirement of scaling horizontally with the number of user requests, we need to use a persistent storage layer for our user's recording rules.

TODO
* Object storage is already used in Thanos. Users will already have this infrastructure provisioned.

### How will users perform authentication and authorization?

TODO
* Probably the same way that Telemeter client and server authentication works?
* How will users provision new API keys / secret material?

## Alternatives

### More frequent deploys

The above problem has manifested because our deploys are too infrequent for our users. Could we not use this as an opportunity to optimize the velocity at which we deploy to production? That way we have the benefit of solving the near-term problems of our users, and also rolling things out more frequently.

As it stands, any rollout to production requires coordination and management from a member of the operating team. As the number of tenants and users becomes larger, the operating team becomes more of a bottleneck to common user operations.

Our goal here is to remove our team entirely from the control loop of updating Prometheus rules. We are trading off immediate term technical complexity against longer-term scalability and happier multi-tenancy operation.

This option also does not address the trust-boundary issue identified above (but that is not a pain point we currently experience internally at Red Hat ¯\_(ツ)_/¯).

### Store Rules in Kubernetes API

One option would be to store user's rule configuration in the Kubernetes API server, perhaps using a 'native' resource like `PrometheusRule`.

This is certainly attractive as we could leverage Kubernetes' in-built storage mechanisms to serve user's needs, and also leverage an operator to synchronise the rules into the right place.

This becomes problematic with a large number of users. The stress placed on the kubernetes api-server, and by extension the happy operation of our cluster, should not depend on the user's configuration. This option is a non-starter on operational grounds.

The backing etcd datastore imposes a limit of 1MB per file, so we could not support large rule files provided by the user.

### Per-tenant GitOps

Another option would be for each tenant to commit their rule configuration into a git repository they control, and then we synchronise that into the right place in our infrastructure based on a URL they provide. This would eliminate the need for us to manage state on their behalf.

This is a neat idea that avoids deploy dependencies, but has some drawbacks:
* Validation - The user's rules would only be validated by us in the Ruler itself, meaning mis-configured rules do not get surfaced to users. We could encourage users to validate
* Secrets - We can assume that users may not want their production rules and alerts exposed to the world, this would then require us to exchange and store secrets with them to access their repository.

The overhead and complexity of managing sensitive user secrets is out-of-scope for this project.

## Action Plan

* Iterate and finalise this design document.
* Implement proof-of-concept Rules API [#138](https://github.com/observatorium/api/pull/138).
* Make the API available to users in staging.
* Promote API to production after ~1 month of running in staging.
