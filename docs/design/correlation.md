---
title: "Correlation"
description: "Correlation between metrics, logs, and traces."
lead: ""
date: 2021-04-30T10:40:00+00:00
lastmod: 2021-04-30T10:40:00+00:00
draft: false
images: []
menu:
  docs:
    parent: "design"
weight: 100
toc: true
---

There needs to be a suggestion engine to suggest potential queries to compose links between signals.

Each signal must have cluster, namespace and pod. Container is a nice to have.

## Metrics to Logs, Traces

### Metric to Log

Alert fires:

sum(rate(http_requests_total{code="5.."}[5m])) / sum(rate(http_requests_total[5m])) > X

This query gives us a fleet wide aggregation of percentage of 5xx errors. We want to see logs from a container that produced such errors.

Problem: The aggregation lost the required metadata identifying the workload.

topk(10, sum by(cluster, namespace, pod, container) (rate(http_requests_total{code="5.."}[5m])) / sum by(cluster, namespace, pod, container) (rate(http_requests_total[5m])))

With the result of the above query we have enough metadata to jump to logs within the violated -10m - time(alert fired). This significantly narrows down the amount of data that needs to be browsed to continue troubleshooting.

### Metric to trace

Alert fires:

histogram_quantile(X, sum(rate(http_requests_duration_seconds[5m]))) > Y

This query gives us a fleet wide aggregation of tail percentile latency. We want to see an example trace from a container that reports responses of such latency.

Problem: The aggregation lost the required metadata identifying the workload.

topk(10, histogram_quantile(X, sum by(cluster, namespace, pod, container) (rate(http_requests_duration_seconds[5m]))))

With the result of the above query we have enough metadata to jump to a trace of a requests of a container within the violated -10m - time(alert fired). This significantly narrows down the amount of data that needs to be browsed to continue troubleshooting. Potentially we can filter traces with the >Y filter.

## Logs to Traces, Metrics

TBD

## Traces to Logs, Metrics

TBD
