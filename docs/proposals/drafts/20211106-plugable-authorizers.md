---
toc: true
title: 2021-11 Pluggable Authorizers
---

## Introduce Pluggable Authorizers

* **Owners**
  * @tareqmamari

## TL; DR

We propose a pluggable authorizers support to enable plugging in custom authorizers in a self-register manner besides the already supported OPA and static RBAC-rules authorizers.

## Why

This proposed approach allows the adopters of Observatorium to introduce and use their own custom in-process authorizers' implementation other than the current static RBAC-rules or OPA-based authorizers. Currently, adopters can build their own OPA integration, which means introducing another component, in one hand, it allows adopters to use their own custom implementation, however, it also comes with additional efforts such as managing the new OPA-integration component's lifecycle. There are adopters, who prefer to have custom in-process authorizers other than using OPA-integration for different reasons, one important reason besides avoid managing a new component, in-process authorizers tend to work better for medium to large load.

## Goals

- Exploit the current common interface for the authorizers, so that the adopters can use it to implement their own.
- Provide a self-registering mechanism of the authorizers so that the adopters do not need to patch the current codebase to introduce their own authorization providers.
- Adopters should be able to configure the tenants, in a way it can specify which authorizer provider to use for each tenant.
- The new changes must be backward-compatible. In other words, the current authorizer implementation and configuration must not be broken by this change and should be supported to use the new pluggable based implementation.

## How

We propose to refactor the current way of initializing the authorizers to exploit the existing common interface. In addition, we aim to provide a self-registering mechanism, so that introducing any new authorizer does not necessary require to do any patch/change in the codebase, instead, only the plugin's go source file is needed other than patching the other source files to add the plugin. For that, we propose to use importing for side-effect mechanism, where a protected map of authorizer's provider type and their factories can be used for such scenario.

## Implementation Options

1. Importing for side-effect
2. Go Plugin Module

## 1. Importing for side-effects

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

### 2. Go Plugin Module

In this approach, the generic interface for authorization providers (Authorizers) is also used. Authorization providers then implement that interface in a go plugin. This plugin then will be built, and the compiled plugin is loaded at run time by Observatorium API, how and where to load the plugin module would be defined by the tenants’ config file.

This approach is easy to implement, very flexible in terms of adding more Authorizers without any additional implementation and integration complexity. However, there are few concerns when it comes to trust and security, since Observatorium API would consume a pre-built plugin binary as an authorizer.

**Reference**: [Writing Modular Go Programs with Plugins](https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9)

# Conclusion

We propose to use the first approach (importing for side-effects), as it appears to be the most flexible approach to follow and require no additional changes to onboard a new authorizer or additional resources to manage.

It worth mentioning, that the new changes must be backward compatible, in other words, it must not break neither the existing configuration nor the current authorization flows.

## Migration plan

The current configuration must not be broken, however, to migrate the current tenants that are configured with OPA or static RBAC-rules, the configuration should be changed but not required:

### Authentication

Current OPA authorization configuration:

```yaml
  opa:
    query: data.observatorium.allow
    paths:
      - ./test/config/observatorium.rego
      - ./test/config/rbac.yaml
```

Proposed new OIDC authenticator configuration:

```yaml
  authorizer:
    type: opa
    config:
      query: data.observatorium.allow
      paths:
        - ./test/config/observatorium.rego
        - ./test/config/rbac.yaml
```

Currently, the static RBAC-rules based authorizer is configured as a fallback if OPA is not used, however, this proposal would make it possible to configure separate static rules per tenant. Example of such config:

```
  authorizer:
    type: rbac
    config:
      
```

## Action plan
- [ ] Review and finalize this proposal document.
- [X] Provide a Proof of Concept (PoC) for pluggable authorizers.
- [ ] The community to review and provide feedback on the PoC Pull Requests.
- [ ] Apply any feedback and finalize the PRs to be merged.
- [ ] Merge the PRs and release a new Observatorium API release.
