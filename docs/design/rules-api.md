# Rules API

Observatorium is featured with a multi-tenant API that enables tenants to write and read Prometheus recording and alerting rules.

The API is [defined](https://github.com/observatorium/api/blob/main/rules/spec.yaml) using the [OpenAPI](https://swagger.io/specification/)
specification format.

The goal is to enable tenants to create, modify and access their own rules.

## Usage

### List rules

```bash
GET /api/v1/rules/raw
```

List rules for a tenant.

#### Example request

```bash
curl http://<observatorium-api-url>/api/metrics/v1/<tenant>/api/v1/rules/raw
```

Where:

* `<observatorium-api-url>` is the URL where Observatorium API is hosted.
* `<tenant>` is the tenant name


#### Example response

The response format is in `application/yaml`.

| Status Code | Description                                              |
|-------------|----------------------------------------------------------|
| 200         | Successfully listed rules                                |
| 401         | Error finding tenant/tenant ID                           |
| 500         | A server side error happened while trying to list rules. |

### Create rules

```bash
PUT /api/v1/rules/raw
```

Set rules for a tenant.

#### Example request

```bash
curl -X PUT --data-binary @alerting-rule.yaml --header "Content-Type: application/yaml" http://<observatorium-api-url>/api/metrics/v1/<tenant>/api/v1/rules/raw
```

Where:

* `rule.yaml` is a YAML file containing the desired rule definition.
* `<observatorium-api-url>` is the URL where the Observatorium API is hosted.
* `<tenant>` is the tenant name


#### Example of a rule.yaml file

The `rule.yaml` should be defined following the [Observatorium OpenAPI specification](https://github.com/observatorium/api/blob/main/rules/spec.yaml). The syntax is based on the Prometheus
[recording](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) and [alerting](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/).

Example of a `rule.yaml` file containing an alerting rule:

```yaml
groups:
  - name: test-alerting-rule
    interval: 30s
    rules:
    - alert: HighRequestLatency
      expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
      for: 10m
      labels:
        severity: page
      annotations:
        summary: High request latency
```

#### Example response