---
toc: true
title: 2021-07 Authz Matchers API
---

## Implement Authz Matchers API for OPA authorizers

* **Owners**
  * @periklis

* **Related Tickets**
  * [LOG-1513: Provide an opa-openshift authorizer for OpenShift Cluster Logging using Loki](https://issues.redhat.com/browse/LOG-1513)

* **Other Docs**
  * Cross-Tenancy proposal???

## TL;DR;

We propose an API that allows OPA authorizers to return label matchers which get enforced read endpoints by the observatorium api.

## Why

The proposed API addresses the issue to use domain-specific authorization knowledge to read queries. This enables a much richer isolation of read queries for the observatroium api.

## Goals

* Users can configure OPA authorizers to pass a set of label matchers back to the observatorium api.
* The observatorium api can extend the label sets automatically for queries against the read endpoints for all supported observability signals.

## Non-Goals

* Provide a generic label matching engine inside the observatorium api for query result sets.

## How

OPA authorizers are the single source of through to check if a given subject (e.g. identified by a bearer token) has access to a requested resource (e.g. read/write on metrics and/or logs). Querying either the in-process via rego or any REST OPA authorizer relays this task from the observatorium-api to a separate isolated component. In addition an OPA authorizer can relay the request further to a third-party system (e.g. RedHat AMS, OpenShift/Kubernetes API server). Upon successful authorization the OPA authorizer can request further information (e.g. list of sub-resources the subject has access) and pass it back to the observatorium api. Such further information can be used to further refine the read endpoints (e.g. return logs for kubernetes namespaces the subject has access to).

We propose to extend and further standardize the API between the observatorium-api and OPA authorizers. The current implementation includes only a boolean flag to signal success or failure for the authorization request against an OPA endpoint. This proposal introduces a second optional field to pass label matchers back to the API, e.g.:

Current implementation payload:

```json
{
    "result": true|false
}
```

Proposed implementation payload:

```json
{
    "result": true|false,
    "matchers": [
        {
            "name":  "label key",
            "type":  "MatchEqual|MatchNotEqual|MatchRegex|MatchNotReqex",
            "value": "label value"
        }
    ]
}
```

Furthermore the observatorium api will handle an empty or missing `matchers` field as obligatory. Thus a successful authorization request with an empty or missing `matchers` field is accepted as successful authorization.

## Migration plan

The present implementations of OPA authorizers (e.g. [opa-ams](https://github.com/observatorium/opa-ams)) do not require any migration. The observatorium-api will handle the response payload following the missing `matchers` approach. Thus these authorizers will not put further constraints on the read endpoints.

## Action plan

- [] Iterate and finalize this design document.
- [] Ensure authz matchers API is applicable to other signals (e.g. tracing).
- [] Implement a proof-of-concept Authz Matchers API [observatorium/api#151](https://github.com/observatorium/api/pull/151) for PromQL and LogQL.
- [] Make this API GA.
