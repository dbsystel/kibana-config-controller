:warning: The project has been archived and is no longer maintained!

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Build Status](https://travis-ci.org/dbsystel/kibana-config-controller.svg)](https://travis-ci.org/dbsystel/kibana-config-controller)

# Config Controller for Kibana
This Controller is based on the [Grafana Operator](https://github.com/tsloughter/grafana-operator). The Config Controller should be run within [Kubernetes](https://github.com/kubernetes/kubernetes) as a sidecar with [Kibana](https://github.com/elastic/kibana).

It watches for new/updated/deleted *ConfigMaps* and if they define the specified annotations as `true` it will `POST` each resource from ConfigMap to Kibana via the [Saved Objects API](https://www.elastic.co/guide/en/kibana/current/saved-objects-api.html). This requires Kibana 5.x (or newer).

## Annotations

Currently it supports index-patterns, searches, visualizations and dashboards. The ConfigMap has to have the following annotations:

**Saved object**

`kibana.net/savedobject` with values: `"true"` or `"false"`

(**Id**)

`kibana.net/id` with values: `"0"` ... `"n"`

In case of multiple Kibana *Deployments* in same Kubernetes Cluster the ConfigMaps have to be mapped to the right Kibana instance.
So each *ConfigMap* can be additionaly annotated with the `kibana.net/id` (if not, the default `id` will be `"0"`)

**Note**

Mentioned `"true"` values can be also specified with: `"1", "t", "T", "true", "TRUE", "True"`

Mentioned `"false"` values can be also specified with: `"0", "f", "F", "false", "FALSE", "False"`

**ConfigMap examples can be found [here](configmap-examples).**

## Usage
```
--log-level # desired log level, one of: [debug, info, warn, error]
--log-format # desired log format, one of: [json, logfmt]
--run-outside-cluster # Uses ~/.kube/config rather than in cluster configuration
--kibana-url # Sets the URL and authentication to use to access the Kibana API
--id # Sets the ID, so the Controller knows which ConfigMaps should be watched
```

## Development
### Build
```sh
make build
# to run the linter, tests and build the binary, run
make ci
```

To build a docker image out of it, look at provided [Dockerfile](Dockerfile) example which expects the `kibana-config-controller` binary in the same directory.

## Deployment
Our preferred way to install kibana-config-controller is [Helm](https://helm.sh/). See example installation at our [Helm directory](helm) within this repo.

## Scripts
If you want to export kibana saved-objects into json files you can use the provided [scripts](scripts) within this repo.
