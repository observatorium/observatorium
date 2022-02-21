# Rules API

Observatorium is featured with a multi-tenant API that enables tenants to write and read Prometheus recording and alerting rules.

The API is [defined](https://github.com/observatorium/api/blob/main/rules/spec.yaml) using the [OpenAPI](https://swagger.io/specification/) specification format.

The goal is to enable tenants to create, modify and access their own rules.

## Usage

### Create rules

```
PUT /api/v1/rules/raw
```

Set rules for a tenant.

#### Example request

```bash
curl -X PUT --data-binary @alerting-rule.yaml --header "Content-Type: application/yaml" http://<observatorium-api-url>/api/metrics/v1/<tenant>/api/v1/rules/raw
```

Where:

* `alerting-rule.yaml` is a YAML file containing the desired rule definition. It can contain recording rules, alerting rules or both.
  * *Note: Each time a PUT request is made to the `/api/v1/rules/raw` endpoint, the rules contained in the request will overwrite all other rules for that tenant.*
* `<observatorium-api-url>` is the URL where the Observatorium API is hosted.
* `<tenant>` is the tenant name

#### Example of a rule.yaml file

The `rule.yaml` should be defined following the [Observatorium OpenAPI specification](https://github.com/observatorium/api/blob/main/rules/spec.yaml). The syntax is based on the Prometheus [recording](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) and [alerting](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) rules syntax.

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

```
successfully updated rules file
```

| Status Code | Description                                                                                           |
|-------------|-------------------------------------------------------------------------------------------------------|
| 200         | Successfully listed rules.                                                                            |
| 401         | Error finding tenant/tenant ID.                                                                       |
| 500         | A server side error happened while trying to create rules or while trying to write the response body. |

### List rules

```
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

```yaml
groups:
- interval: 30s
  name: test-alerting-rule
  rules:
  - alert: HighRequestLatency
    annotations:
      summary: High request latency
    expr: job:request_latency_seconds:mean5m{job="myjob",tenant_id="1234"}
      > 0.5
    for: 10m
    labels:
      severity: page
      tenant_id: 1234
```

The response format is in `application/yaml`.

*Note: In the current implementation from the Rules OpenAPI specification in Observatorium API, to better validate tenant rules, the label `tenant_id` was enforced in the read path:*

* *In the `labels` field and*
* *In the metrics that are present in the expression defined in the `expr` field.*

| Status Code | Description                                              |
|-------------|----------------------------------------------------------|
| 200         | Successfully listed rules.                               |
| 401         | Error finding tenant/tenant ID.                          |
| 500         | A server side error happened while trying to list rules. |

## Difference between /api/v1/rules/raw and /api/v1/rules endpoints

Note that the `/api/v1/rules/raw` endpoint differs from `/api/v1/rules` endpoint:

* `/api/v1/rules/raw` supports `GET` and `PUT` requests, as described above. It refers only to the rules that were defined by the tenant that was authorized to use the endpoint.
* `/api/v1/rules` supports `GET` requests and is the endpoint that is proxied by the Observatorium API to the read endpoint (in this case, Thanos Querier). It contains all rules from all tenants.
