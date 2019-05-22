### Installing the Chart

To install the chart with the release name `kibana-config-controller` in namespace `logging`:

```console
$ helm upgrade kibana-config-controller charts/kibana-config-controller --namespace logging --install
```
The command deploys kibana-config-controller on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

### Uninstalling the Chart

To uninstall/delete the `kibana-config-controller` deployment:

```console
$ helm delete kibana-config-controller --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
The following table lists the configurable parameters of the kibana-controller chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of pod replicas | `1`
`image.repository` | kibana-config-controller container image repository | `dockerregistry/kibana-config-controller`
`image.tag` | kibana-config-controller container image tag | `1.1.0`
`url` | The url to access kibana | `http://localhost:5601`
`id` | The id to specify kibana | `0`
`logLevel` | The log-level of kibana-config-controller | `info`
`logFormat` | The log-format of kibana-config-controller | `json`