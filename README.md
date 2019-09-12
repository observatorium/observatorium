# Configuration

This repository contains the configuration for the Observatorium instances that the team runs.

[![Build Status](https://cloud.drone.io/api/badges/observatorium/configuration/status.svg)](https://cloud.drone.io/observatorium/configuration)

## Getting started

* To get all the required tools, run `make install-tools`
* Bring the dependencies by running `make vendor`
* To generate all the manifests, run `make build`. The generated manifest files are located under `environments/**/manifests`
* Verify that the jsonnet files are properly formated: `make check`
* To simulate what the CI does, `make ci`
