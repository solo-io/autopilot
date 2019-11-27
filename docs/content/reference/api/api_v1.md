---
title: "Autopilot Config Files"
weight: 5
---

<a name="top"></a>


## Table of Contents

- [autopilot.proto](#autopilot.proto)
    - [AutopilotProject](#autopilot.AutopilotProject)
    - [Input](#autopilot.Input)
    - [MetricsQuery](#autopilot.MetricsQuery)
    - [Output](#autopilot.Output)
    - [Phase](#autopilot.Phase)
    - [Resource](#autopilot.Resource)
    - [ResourceParameter](#autopilot.ResourceParameter)
    - [ThirdPartyResource](#autopilot.ThirdPartyResource)
  
  
  
  

- [autopilot-operator.proto](#autopilot-operator.proto)
    - [AutopilotOperator](#autopilot.AutopilotOperator)
  
    - [MeshProvider](#autopilot.MeshProvider)
  
  
  




<a name="autopilot.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## autopilot.proto
The following Schema defines the structure of the `autopilot.yaml` configuration file.

This file is used to generate and re-generate the project structure, as well
as execute tasks related to build and deployment. It can be consumed
both via the `ap` CLI as well as in `codegen` packages.


<a name="autopilot.AutopilotProject"></a>

### AutopilotProject
The AutopilotProject file is the root configuration file for the project itself.

This file will be used to build and deploy the autopilot operator.
It is loaded automatically by the autopilot CLI. Its
default location is 'autopilot.yaml'


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| operatorName | [string](#string) |  | the name of the Operator this is used to name and label loggers, k8s resources, and metrics exposed by the operator. Should be [a valid Kube resource name](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names). |
| resources | [][Resource](#autopilot.Resource) | repeated | the set of Top-Level Custom Resources that are managed by this Operator. the Operator will run a [Controller](https://kubernetes.io/docs/concepts/architecture/controller/) loop for each resource. To add CRDs without creating a controller, set enableController: false on the resource. |
| thirdPartyResources | [][ThirdPartyResource](#autopilot.ThirdPartyResource) | repeated | Third-party CRDs which can be used as parameters. Extends Autopilot's builtin types |
| queries | [][MetricsQuery](#autopilot.MetricsQuery) | repeated | custom Queries which extend Autopilot's builtin metrics queries |






<a name="autopilot.Input"></a>

### Input
Input represents an input parameter type
These can either be a k8s resource,
a metric, or a webhook.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [ResourceParameter](#autopilot.ResourceParameter) |  | a kubernetes resource. this can include builtin and custom resources. <br> Only one of `resource`, `metric`, or `webhook` should be set. |
| metric | [string](#string) |  | a named metric query. this can reference either a built-in query, or a custom query defined in the root of the `autopilot.yaml` the names of built-in queries are listed here: https://docs.solo.io/autopilot/latest/reference/queries |
| webhook | [string](#string) |  | the name of a webhook defined in the `autopilot.yaml`. Phase inputs will contain a queue of unprocessed of payloads received by the webhook. |






<a name="autopilot.MetricsQuery"></a>

### MetricsQuery
MetricsQueries extend the query options available to workers.
MetricsQueries are accessible to workers via generated client code
that lives in <project root>/pkg/metrics


The following MetricsQuery:

```
name: success-rate
parameters:
- Name
- Namespace
- Interval
queryTemplate: |
    sum(
        rate(
            envoy_cluster_upstream_rq{
                kubernetes_namespace="{{ .Namespace }}",
                kubernetes_pod_name=~"{{ .Name }}-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)",
                envoy_response_code!~"5.*"
            }[{{ .Interval }}]
        )
    )
    /
    sum(
        rate(
            envoy_cluster_upstream_rq{
                kubernetes_namespace="{{ .Namespace }}",
                kubernetes_pod_name=~"{{ .Name }}-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)"
            }[{{ .Interval }}]
        )
    )
    * 100
```

would produce the following `metrics` Interface:

```go
type CanaryDeploymentMetrics interface {
    metrics.Client
    GetIstioSuccessRate(ctx context.Context, Namespace, Name, Interval string) (*metrics.QueryResult, error)
    GetIstioRequestDuration(ctx context.Context, Namespace, Name, Interval string) (*metrics.QueryResult, error)
    GetEnvoySuccessRate(ctx context.Context, Namespace, Name, Interval string) (*metrics.QueryResult, error)
    GetEnvoyRequestDuration(ctx context.Context, Namespace, Name, Interval string) (*metrics.QueryResult, error)
}
```


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| queryTemplate | [string](#string) |  |  |
| parameters | [][string](#string) | repeated |  |






<a name="autopilot.Output"></a>

### Output
Output represents an output parameter type
Currently, these can only be a k8s resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [ResourceParameter](#autopilot.ResourceParameter) |  | a kubernetes resource. this can include builtin and custom resources. <br> Only `resource` can currently be set. |






<a name="autopilot.Phase"></a>

### Phase
MeshProviders provide an interface to monitoring and managing a specific
mesh.

Autopilot does not abstract the mesh API - Autopilot developers must
still reason able about Provider-specific CRDs. Autopilot's job is to
abstract operational concerns such as discovering control plane configuration
and monitoring metrics.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name of the phase. must be unique |
| description | [string](#string) |  | description of the phase. used for comments and docs |
| initial | [bool](#bool) |  | indicates whether this is the initial phase of the system. exactly one phase must be the initial phase |
| final | [bool](#bool) |  | indicates whether this is a "final" or "resting" phase of the system. when the CRD is in the final phase, no more processing will be done on it |
| inputs | [][Input](#autopilot.Input) | repeated | The set of inputs for this phase. The inputs will be retrieved by the scheduler and passed to the worker as input parameters. |
| outputs | [][Output](#autopilot.Output) | repeated | the set of outputs for this phase the inputs will be propagated to k8s storage (etcd) by the scheduler.

custom outputs can be defined in the autopilot.yaml |






<a name="autopilot.Resource"></a>

### Resource
An Autopilot Resource is a Custom Resource.
Autopilot will generate Go code for


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | the name (kubernetes Kind) of the Custom Resource e.g. "MyResource" |
| group | [string](#string) |  | the Api Group of the top-level CRD for the operator e.g. "mycompany.io" |
| version | [string](#string) |  | e.g. "v1" |
| phases | [][Phase](#autopilot.Phase) | repeated | Each phase represents a different stage in the lifecycle of the CRD (e.g. Pending/Succeeded/Failed). <br> Each phase specifies a unique name and its own set of inputs and outputs. <br> If a controller is generated for this Resource, each phase will define the inputs/outputs and work function the controller will run. |
| enableController | [google.protobuf.BoolValue](#google.protobuf.BoolValue) |  | Generate and run a controller to manage this resource. This is set to 'true' by default. Set this to 'false' to create the resource without generating or running a controller for it. |
| enableFinalizer | [bool](#bool) |  | enable use of a Finalizer to handle object deletion. only applies if enableController is not set to false |






<a name="autopilot.ResourceParameter"></a>

### ResourceParameter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | resource Api Kind |
| group | [string](#string) |  | resource Api Group. leave empty for core resources |
| version | [string](#string) |  | resource Api Version |
| list | [bool](#bool) |  | parameter should be a list of resources (in one or all namespaces) if set to false (default) |






<a name="autopilot.ThirdPartyResource"></a>

### ThirdPartyResource
ThirdPartyCustomResource allow code to be generated
for input/output CRDs that are not built-in to Autopilot.
These types must be Kubernetes-compatible Go structs.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [string](#string) |  | the singular CamelCased name of the resource equivalent to Kind |
| group | [string](#string) |  | Kubernetes API group for the resource e.g. "networking.istio.io" |
| version | [string](#string) |  | Kubernetes API Version for the resource e.g. "v1beta3" |
| pluralKind | [string](#string) |  | the plural CamelCased name of the resource equivalent to the pluralized form of Kind |
| goPackage | [string](#string) |  | go package (import path) containing the go struct for the resource |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="autopilot-operator.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## autopilot-operator.proto
autopilot-operator.proto defines the API Schema for the autopilot-operator.yaml configuration file.
this file provides the bootstrap configuration that is loaded to the
operator at boot-time/runtime


<a name="autopilot.AutopilotOperator"></a>

### AutopilotOperator
The AutopilotOperator file is the bootstrap
Configuration file for the Operator.
It is stored and mounted to the operator as a Kubernetes ConfigMap.
The Operator will hot-reload when the configuration file changes.
Default name is 'autopilot-operator.yaml' and should be stored in the project root.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [string](#string) |  | version of the operator used for logging and metrics default is "0.0.1" |
| meshProvider | [MeshProvider](#autopilot.MeshProvider) |  | meshProvider determines how the operator will connect to a service mesh Default is "SMI" |
| controlPlaneNs | [string](#string) |  | controlPlaneNs is the namespace the control plane lives in Default is "istio-system" |
| workInterval | [google.protobuf.Duration](#google.protobuf.Duration) |  | workInterval to sets the interval at which CRD workers resync. Default is 5s |
| metricsAddr | [string](#string) |  | Serve metrics on this address. Set to empty string to disable metrics defaults to ":9091" |
| enableLeaderElection | [bool](#bool) |  | Enable leader election. This will prevent more than one operator from running at a time defaults to true |
| watchNamespace | [string](#string) |  | if non-empty, watchNamespace will restrict the Operator to watching resources in a single namespace if empty (default), the Operator must have Cluster-scope RBAC permissions (ClusterRole/Binding) can also be set via the WATCH_NAMESPACE environment variable |
| leaderElectionNamespace | [string](#string) |  | The namespace to use for Leader Election (requires read/write ConfigMap permissions) defaults to the watchNamespace |
| logLevel | [google.protobuf.UInt32Value](#google.protobuf.UInt32Value) |  | Log level for the operator's logger values: 0 - Debug 1 - Info 2 - Warn 3 - Error 4 - DPanic 5 - Panic 6 - Fatal Defaults to Info |





 <!-- end messages -->


<a name="autopilot.MeshProvider"></a>

### MeshProvider
MeshProviders provide an interface to monitoring and managing a specific
mesh.
Autopilot does not abstract the mesh API - Autopilot developers must
still reason able about Provider-specific CRDs. Autopilot's job is to
abstract operational concerns such as discovering control plane configuration
and monitoring metrics.

| Name | Number | Description |
| ---- | ------ | ----------- |
| Istio | 0 | the Operator will utilize Istio mesh for metrics and configuration |
| Custom | 1 | the Operator will utilize a locally deployed Prometheus instance for metrics (Currently unimplemented) |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->


