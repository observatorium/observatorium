---
toc: true
title: 2021-10 Pluggable Authenticators
---

## Introduce Pluggable Authenticators

* **Owners**
  * @tareqmamari

## TL; DR

We propose a pluggable authenticator approach to support plugging in custom authenticators in a self-registering manner.

## Why

This proposed approach allows the adopters of Observatorium to use their own custom authenticator's implementation other than the current OIDC, mTLS or Openshift authenticators.

## Goals

- Provide a common interface for the authenticators, so that the adopters can use it to implement their own.
- Provide a self-registering mechanism for authenticators so that the adopters do not need to patch the current codebase to introduce their own authenticators.
- Adopters should be able to configure the tenants such that they can specify which authenticator type to use for each tenant.
- The new changes must be backward-compatible. In other words, the current authentication implementation and configuration must not be broken by this change.

## How

Currently, Observatorium API supports only OIDC and mTLS based authentication, however, there are adopters who do not use standard OIDC or mTLS, instead, they use for example JWT-based access tokens with special handling that is done using private Go modules owned by those adopters.

We propose to refactor the current way of initializing the authenticators to introduce a common interface, which means all new authenticators have the same basic functions, which adopters need to implement for their custom authenticators. In addition, in this proposal, we aim to provide a self-registering mechanism, so that introducing any new authenticator does not necessarily require any additional patch/change to the existing codebase, instead, only the plugin's Go source file is needed other than patching the other source files to add the plugin. For that, we propose to use Go's importing for side-effects, where a protected map of authenticator types and their factories can be used.

## Implementation Options

1. Importing for Side-Effects
2. Go Plugin Module
3. OAuth2 Providers

## 1. Importing for Side-Effects

In this approach, importing for Go's side-effects is used to enable each authenticator to register itself through a map of authenticator types and their factories.

To onboard a new authenticator, the following is needed:

1. Implement the corresponding interface.
2. Implement a factory.
3. Register the factory in the authenticators' factories map.

With this approach, an example tenant configuration would look like:

```
- name: test-oidc
  id: 1610b0c3-c509-4592-a256-a1871353dbfa
  authenticator:
    type: oidc
    config:
      clientID: test
      clientSecret: xyz
      issuerCAPath: ./tmp/certs/ca.pem
      issuerURL: http://127.0.0.1:5556/dex
      redirectURL: https://localhost:8443/oidc/test-oidc/callback
      usernameClaim: email
  rateLimits:
    - endpoint: "/api/metrics/v1/.+/api/v1/receive"
      limit: 100
      window: 1s
    - endpoint: "/api/logs/v1/.*"
      limit: 100
      window: 1s
```

## 2. Go Plugin Module

In this approach, a generic interface for authentication providers (authenticators) is introduced. authentication providers then implement that interface in a Go plugin. This plugin will be built and loaded at run time by the Observatorium API. How and where to load the plugin module is defined by the tenants configuration file.

This approach is easy to implement, very flexible in terms of adding more authenticators without any additional implementation and integration complexity. However, there are few concerns when it comes to trust and security, since the Observatorium API would consume a pre-built plugin binary as an authenticator.

**Reference**: [Writing Modular Go Programs with Plugins](https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9)

## 3. OAuth2 Providers

Similar to the [generic OAuth2 providers in grafana](https://github.com/grafana/grafana/blob/main/pkg/login/social/social.go), OAuth2 providers are [pre-registered in a static list of supported providers](https://github.com/grafana/grafana/blob/e73cd2fdeb3a08db32139f5ce4da4accf162737e/pkg/login/social/social.go#L254) alongside the current implementation of OIDC and mTLS. While this approach provides flexibility and can be considered to be an elegant and clean way of introducing more authenticators support, it requires a lot of changes in the Observatorium API.

However, introducing a new authenticator with private or custom implementation would require changing/patching existing Go source files in the Observatorium API. As a result, adopters need to re-apply their patch whenever migrating to a new Observatorium API release with risk of being affected by merge conflicts that must be resolved manually.

## Conclusion

We propose to use the first approach (importing for side-effects), as it appears to be the most flexible approach to follow and requires no additional changes to onboard a new authenticator or additional resources to manage.

It worth mentioning that the new changes must be backward-compatible, in other words, they must not break neither the existing configuration nor the current authentication flows.

## Migration Plan

The current configuration format will not be broken but will be deprecated, so immediate migration is not necessary. Tenants configured to use the old authenticators, namely OIDC, mTLS, and OpenShift, should eventually be migrated to use the new configuration format.

Current OIDC authentication configuration:

```yaml
  oidc:
    clientID: test
    clientSecret: ZXhhbXBsZS1hcHAtc2VjcmV0
    issuerCAPath: ./tmp/certs/ca.pem
    issuerURL: http://127.0.0.1:5556/dex
    redirectURL: https://localhost:8443/OIDC/test-OIDC/callback
    usernameClaim: email
```

Proposed new OIDC authenticator configuration:

```yaml
  authenticator:
    type: oidc
    config:
      clientID: test
      clientSecret: ZXhhbXBsZS1hcHAtc2VjcmV0
      issuerCAPath: ./tmp/certs/ca.pem
      issuerURL: http://127.0.0.1:5556/dex
      redirectURL: https://localhost:8443/oidc/test-oidc/callback
      usernameClaim: email
```

Current mTLS authentication configuration:

```yaml
  mtls:
    caPath: ./tmp/certs/ca.pem
```

Proposed new mTLS authenticator configuration:

```yaml
  authenticator:
    type: mtls
    config:
      caPath: ./tmp/certs/ca.pem
```

## Action plan
- [X] Review and finalize this proposal document.
- [X] Provide a Proof of Concept (PoC) for both pluggable authenticators.
- [X] The community to review and provide feedback on the PoC Pull Requests.
- [X] Apply any feedback and finalize the PRs to be merged.
- [X] Merge the PRs and release a new Observatorium API release.
