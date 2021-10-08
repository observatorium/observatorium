---
toc: true
title: 2021-10 Pluggable Authenticators and Authorizers
---

## Introduce Pluggable Authenticators and Authorizers

* **Owners**
* @tareqmamari

## TL; DR

We propose a pluggable authenticators and authorizers to support plugging in custom authenticators and authorizers in a self-register manner.

## Why

This proposed approach allows the adopters of Observatorium to use their own custom authenticators and authorizers’ implementation other than the current OIDC or mTLS authenticators, OPA or RBAC authorizers.

## Goals

- Provide a common interface for the authorizers and another one for authenticators, so that the adopters can use it to implement their own authenticators/authorizers.
- Provide a self-registering mechanism of the authenticators and authorizers so that the adopters do not need to patch the current codebase to introduce their own authenticators/authorizers.
- Adopters should be able to configure the tenants, in a way it can specify which authenticator/authorizer type to use for each tenant.
- The new changes must be backward-compatible. In other words, the current authentication/authorization implementation and configuration must not be broken by this change.

## How

Currently, Observatorium API supports only OIDC and mTLS based authentication/authorization, however, there are adopters who do not use standard OIDC or mTLS, instead, they use for example jwt-based access tokens with special handling that is done using private go modules owned by those adopters.

Like the authentication, only OPA and RBAC authorizers are currently supported, while there is the option to build a custom OPA component where Observatorium API is configured to use in order to do the authorization, it is introducing overhead in one way or another, and, if that OPA component resides within the same Observatorium API network, as a side-car in Kubernetes deployment for example, that have current issues, in particular, in istio-managed clusters.

We propose to refactor the current way of initializing the authenticators to introduce a common interface like the current authorizers’ implementation, which means all new authenticators have the same basic functions where the adopters need to implement for their custom authenticators. In addition, in this proposal, we aim to provide a self-registering mechanism, so that introducing any new authenticator does not necessary require to do any patch/change in the codebase, instead, only the plugin's go source file is needed other than patching the other source files to add the plugin. For that, we propose to use importing for side-effect mechanism, where a map of authenticator type and their factories can be used for such scenario.

## Implementation Options

1. Importing for side-effect
2. Go Plugin Module
3. Oauth2 Providers

## 1. Importing for side-effects

Like [Option 3](#3-oauth2-providers) except that we do not have a static list of supported authenticators/authorizers. Instead, each authenticator/authorizer can register themselves through a map of authenticator types and their factories.

To onboard a new authenticator/authorizer, the following is needed:

1. Implement the corresponding interface.
2. Implement a factory.
3. Register the factory in the authenticators/authorizers’ factories map, this is the part where it registers itself.

The factory should receive two parameters: 1) config map of strings and interfaces, and 2) a base authenticator/authorizer. The base authenticator/authorizer holds common objects such as a logger, tenant name and authenticator/authorizer type.

With this approach, an example of tenant configuration would like:

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
  authorizer:
    type: rbac
    config:
      rbacFilePath: "test/config/rbac.yaml"
  rateLimits:
    - endpoint: "/api/metrics/v1/.+/api/v1/receive"
      limit: 100
      window: 1s
    - endpoint: "/api/logs/v1/.*"
      limit: 100
      window: 1s
```

### 2. Go Plugin Module

In this approach, a generic interface for Authentication/authorization providers (Authenticators) is introduced. Authentication/authorization providers then implement that interface in a go plugin. This plugin then will be built, and the compiled plugin is loaded at run time by Observatorium API, how and where to load the plugin module is defined by the tenants’ config file.

This approach is easy to implement, very flexible in terms of adding more Authenticators without any additional implementation and integration complexity. However, there are few concerns when it comes to trust and security, since Observatorium API would consume a pre-built plugin binary as an authenticator/authorizer.

**Reference**: [Writing Modular Go Programs with Plugins](https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9)

## 3. Oauth2 Providers

Similar to the [generic Oauth2 providers in grafana](https://github.com/grafana/grafana/blob/main/pkg/login/social/social.go), Oauth2 providers can be introduced and injected into the supported Oauth2 besides to the current implementation of OIDC and mTLS to prevent any disruption for the current adopters. While this approach provides flexibility and it considers to be an elegant and clean way of introducing more Authenticators support, it requires a lot of changes in the Observatorium API.

However, introducing a new authenticator with private or custom implementation would require changing/patching EXISTING Go source files in the Observatorium API to achieve that. As a result, adopters need to re-apply their patch whenever migrating to a new Observatorium API release with risk to be affected by merge conflicts that must be resolved manually.

# Conclusion

We propose to use the last approach (importing for side-effects), as it appears to be the most flexible approach to follow as well as there are no additional changes are required to onboard a new authenticator/authorizer as well as no additional resources to manage.

It worth mentioning, that the new changes must be backward compatible, in other words, it must not break neither the existing configuration nor the current authentication/authorization flows.

## Migration plan

The current configuration must not be broken, however, to migrate the current tenants that are configured with OIDC or mTLS, the configuration should be changed:

### Authentication

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

### Authorization

#### OPA Authorization

Current OPA authorization configuration:

```yaml

```

Proposed new OPA authorizer configuration, for REST based OPA:

```yaml
  authorizer:
    type: opa
    config:
      url: http://127.0.0.1:9988/v1/data/observatorium/allow
      withAccessToken: true
```

or for in-process OPA:

```yaml
  authorizer:
    type: opa
    config:
      query: data.observatorium.allow
      paths:
        - ./test/config/observatorium.rego
        - ./test/config/rbac.yaml
```

#### RBAC static rules-based Authorization

The current implementation, passes the RBAC static rules as a file, instead, in the new proposed implementation, the new proposed RBAC configuration would look like:

```yaml
  authorizer:
    type: rbac
    config:
      rbacFilePath: "rbac.yaml"
```

## Action plan
- [ ] Review and finalize this proposal document.
- [X] Provide a Proof of Concept (PoC) for both pluggable authenticators and authorizers.
- [ ] The community to review and provide feedback on the PoC Pull Requests.
- [ ] Apply any feedback and finalize the PRs to be merged.
- [ ] Merge the PRs and release a new Observatorium API release.
