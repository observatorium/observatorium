# Observatorium

Observatorium is a unified service to alert and troubleshoot infrastructure. It combines observability signals such as metrics, logs, traces, profiles and more in a single experience to reduce mean time to resolution.

## Technologies

Observatorium makes use of existing technologies in opinionated and multi-tenant ways, providing a consistent and reliable experience.

At the highest layer, Observatorium offers an API aggregator, exposing known upstream APIs for ingesting and querying observability signals. Each signal then has its own specialized backend:

* [Metrics](design/metrics.md)
* [Logs](design/logs.md)
* [Traces](design/traces.md)
* [Profiles](design/profiles.md)

It extends the individual experience by also offering functionality to [correlate](design/correlation.md) signals among each other.
