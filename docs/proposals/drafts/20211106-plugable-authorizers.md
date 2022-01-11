---
toc: true
title: 2021-11 Pluggable Authorizers
---

## Introduce Pluggable Authorizers

* **Owners**
  * @tareqmamari

## TL; DR

We propose a pluggable authorizers support to enable plugging in custom authorizers in a self-register manner besides the already supported authorizers in Observatorium API.

## Why

This proposed approach allows the adopters of Observatorium to introduce and use their own custom in-process authorizers' implementation. Currently, adopters can build their own OPA integration, which means introducing another component where OPA server/endpoints are running. In one hand, it allows adopters to use their own custom implementation, however, it also comes with additional efforts such as managing the new OPA-integration component's lifecycle. There are adopters, who prefer to have custom in-process authorizers other than using OPA-integration for different reasons, one important reason besides avoid managing a new component, in-process authorizers tend to work better for medium to large workloads.

## Goals

- Do not change the current common interface for the authorizers.
- Provide a self-registering mechanism of the authorizers so that the adopters do not need to patch the current codebase to introduce their own custom authorization logic.
- The new changes must be backward-compatible. In other words, the current authorizer implementation and configuration must not be broken by this change and should be supported to use the new pluggable based implementation.

## How

We propose to extend rego evaluation logic to load custom built-in functions in Go, in this way, adopters can implement their own authorization logic in go and then call it in Rego. In addition, we aim to provide a self-registering mechanism, so that introducing any new custom built-in function does not necessary require to do any patch/change in the existing codebase, instead, only the plugin's go source file is needed other than patching the other source files to add the plugin. For that, we propose to use importing for side-effect mechanism, where a protected map of authorizer's provider type and their custom built-in function can be used for such scenario.

## Implementation Options
1. Custom Rego Built-in Functions in Go
2. Importing for Side-Effect
3. Go Plugin Module

## 1. Custom Rego Built-in Functions in Go

Instead of totally introducing a new authorizer, OPA inprocess authorizer ( using Rego ) can be extended through [custom rego built-in functions in go](https://www.openpolicyagent.org/docs/latest/extensions/#custom-built-in-functions-in-go). In this approach, custom auhtorization go logic can be introduced as a custom built-in rego function in go, which then can be called in the rego logic. Once rego calls this function, the go implementation of it, would be executed. To make it pluggable, importing for side-effects is used, similar to what is proposed in the second option.

To onboard custom authorization with this approach, the following is needed:
1. Add a new go file with the custom authorization logic.
2. The new custom built-in function should register itself through an onboarding map.
3. All registered functions will be loaded into the rego env during the evaluation.

It worth mentioning, in this approach, we do not need to change the current configuraiton format for the `OPA` config.

## 2. Importing for Side-Effects

Each authorizer can register itself through a protected map of authorizer's provider types and their factories.

To onboard a new authorizer, the following is needed:

1. Implement the authorizer's provider interface.
2. Implement the new provider's factory.
3. Register the factory in the authorizers' factories map, this is the part where it registers itself.

The factory should receive three parameters: the authorizer's config, which is a map of strings and interfaces, tenant's name and logger instance.

The following is an example of how the config would look like:

```
- name: test-oidc
  id: 1610b0c3-c509-4592-a256-a1871353dbfa
  authorizer:
    type: opa
    config:
      query: data.observatorium.allow
      paths:
        - ./test/config/observatorium.rego
        - ./test/config/rbac.yaml
     rateLimits:
    - endpoint: "/api/metrics/v1/.+/api/v1/receive"
      limit: 100
      window: 1s
    - endpoint: "/api/logs/v1/.*"
      limit: 100
      window: 1s
```

## 3. Go Plugin Module

In this approach, the generic interface for authorization providers (Authorizers) is also used. Authorization providers then implement that interface in a go plugin. This plugin then will be built, and the compiled plugin is loaded at run time by Observatorium API, how and where to load the plugin module would be defined by the tenants’ config file.

This approach is easy to implement, very flexible in terms of adding more Authorizers without any additional implementation and integration complexity. However, there are few concerns when it comes to trust and security, since Observatorium API would consume a pre-built plugin binary as an authorizer.

**Reference**: [Go Plugin Package](https://pkg.go.dev/plugin)

## Conclusion

We propose to use the first approach (Custom Rego Built-in Functions in Go), as it appears to be the most flexible approach to follow and require no additional changes to onboard a new authorizer or additional resources to manage.

It worth mentioning, that the new changes must be backward compatible, in other words, it must not break neither the existing configuration nor the current authorization flows.

## Action plan
- [ ] Review and finalize this proposal document.
- [X] Provide a Proof of Concept (PoC) for pluggable authorizers.
- [ ] The community to review and provide feedback on the PoC Pull Requests.
- [ ] Apply any feedback and finalize the PRs to be merged.
- [ ] Merge the PRs and release a new Observatorium API release.
