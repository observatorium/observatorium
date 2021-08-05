# Metrics

Observatorium makes extensive use of [Thanos](https://thanos.io/) to provide support for ingesting and querying metrics.

## Ingestion

Observatorium exposes an endpoint that supports the Prometheus [remote-write protocol](https://prometheus.io/docs/prometheus/latest/storage/#remote-storage-integrations). This is primarily implemented using the [Thanos receive](https://thanos.io/tip/proposals-done/201812-thanos-remote-receive.md/) component.

The Thanos receive component supports multi-tenancy out of the box: for each request it must be told the tenant of data being written through the use of the `THANOS-TENANT` HTTP header. In Observatorium, this is made possible by [the API aggregator](https://github.com/observatorium/api): after authenticating and authorizing a write request, the API aggregator sets the tenant identifier in the `THANOS-TENANT` HTTP header for the request. Depending on the tenant's configuration, the setup routes the write request to the appropriate hashring of Thanos receive instances, which then attaches the value of the `THANOS-TENANT` HTTP header as an external label to the metrics using the `tenant_id` label.

## Querying

To query metrics, the standard [Thanos querier](https://thanos.io/tip/thanos/design.md/#query-layer) is used. All metrics written by the Thanos receive component automatically have the tenant added as an external label. Upon querying metrics, a user specifies the tenant to be queried. The system forwards the tenant to query in form of the `THANOS-TENANT` HTTP header. The API aggregator authorizes, that the requesting user has query access to the tenant, and if so, it forces the tenant in the PromQL query to be `tenant_id`, using the [prom-label-proxy](https://github.com/openshift/prom-label-proxy) as a library.
