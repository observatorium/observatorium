---
toc: true
title: 2022-02 Additional Metrics Write Endpoints
---

# Additional Metrics Write Endpoints Proposal

* **Owners:**:
  * `@marcolan018`

* **Related Tickets:**
  * `<JIRA, GH Issues>`

* **Other docs:**
  * [Exporting Metrics To Additional Consumers in RedHat Advanced Cluster Management for Kubernetes](https://docs.google.com/document/d/1t7GodF4s_V5MtFtjTeiT0UFeQuu7cuSzz9-7FEENDsQ) (internal Red Hat link)

## TLDR

We propose to support additional metrics write endpoints. The metrics write requests can be forwarded to additional backend endpoints, which support Prometheus remote write protocol, besides Thanos receivers.

## Why
Currently the Observatorium API always forwards the metrics to Thanos. Given that users have multiple consumers of metric data, they have a requirement for the collected metrics to be made available to these additional consumers. We need a way to let Observatorium API to forward the metrics to more than one targets.
There are some other options to fulfill this requirement. But all of them have problems/restrictions:
1. The clients send metrics to the additional consumers directly. This option is not workable in some scenarios. e.g. there is no network connection between the clients and the additional consumers. Also, some clients have very limited resources and this option will lead to more consumption of cpu and bandwidth.
2. The users of the additional consumers can pull the metrics from Thanos side directly, but that's not a realtime solution. Also it will lead to heavy workload on Thanos if users consistently query massive metrics from it. 
3. Use other tool, such as [Vector](https://vector.dev/). Observatorium API will continue to send metrics to single write endpoint, the Vector endpoint. Then Vector is responsible to forward the metrics to multiple backend write endpoints, including Thanos receiver. Firstly, this option needs to introduce a new component and increase the complexity of existing topology. Furthermore, Vector does not support the Prometheus remote write spec fully due to the [compliance test result](https://prometheus.io/blog/2021/05/04/prometheus-conformance-remote-write-compliance/). It means some type of metrics data cannot be forwarded to write endpoints.

## Goals

* Observatorium API can pushes data to configured additional metrics write endpoints, secured or non-secured

## Non-Goals

* Other types of observability data(e.g. logs) to multiple endpoints not considered. We might explore this in the future

## How

Currently Observatorium API only supports one metrics write endpoint. We propose an update in Observatorium API, to add a new handler RemoteWrite Proxy, to clone the request body of incoming metrics write requests, then forward to additional write endpoints, which support Prometheus remote write protocol.
For each additional metrics write endpoint, the RemoteWrite Proxy will clone the request body and send it to the endpoint based on its' configurations. A sample additional write endpoints configuration with two endpoints is like below:
```yaml
endpints:
  - name: remotewrite
    url: http://xyz/write
  - name: secured_remotewrite
    url: https://xyz:8443/secured_write
    tlsConfig:
      ca: /var/certs/ca.crt
      cert: /var/certs/tls.crt
      key: /var/certs/tls.key
``` 
The proposed sequence diagrame for incoming metrics write request is as below: ![sequence-diagram](../../assets/additional-write-endpoints.png)

The client will send metrics write request to Observatorium API, the request will reach the RemoteWrite Proxy after some middlewares handling. RemoteWrite Proxy will return once it receives the request, to notifify the client that the request has reached, already passed some necessary checking such as authentication, and ready to be forwarded to backend write endpoints. In the meantime, it will send the request to the backend write endpoints, including Thanos receiver. If any request fails finally, the RemoteWrite Proxy will record the errors in the logs. Also, there will be a new metrics named remote_write_requests_count to expose the status for the requests. To make the scenario more robust, we might add retry so that the RemoteWrite Proxy will try to forward the metrics to the write endpoints if run into some temporary errors.

## Action Plan

* Iterate and finalise this design document.
* Implement the Observatorium API to push metrics to additional write endpoints.
* Update Observatorium operator to support additional write endpoints.
