---
title: "Operator Architecture"
description: "Overvriew of the Architecture of Operators built with Autopilot"
weight: 3
---

The architecture of Autopilot Operators is determined by Autopilot and enforced in generated code. While generated code can be manually modified to customize this architecture, it is not currently recommended.

To better understand the architecture of an Autopilot Operator, 
let's use the [Example Canary Operator](https://github.com/solo-io/autopilot/tree/master/test/e2e) as an example.

Consider the following diagram of our Canary. The Canary Operator has the following architecture:

{{<mermaid align="left">}}
 graph LR;
   
    subgraph legend
    userDefined[User-Defined Code]
    generated[Generated Code]
    k8s[Imported Kubernetes Libraries]
    end
    
    subgraph Canary Operator
    w1[Initializing<br>Worker]-->|scheduled when <br>Canary CRD<br> is in Initializing phase|s[Canary Scheduler]
    w2[Waiting<br>Worker]-->|scheduled when <br>Canary CRD<br> is in Waiting phase|s
    w3[Evaluating<br>Worker]-->|scheduled when <br>Canary CRD<br> is in Evaluating phase|s
    w4[Promoting<br>Worker]-->|scheduled when <br>Canary CRD<br> is in Promoting phase|s
    w5[Rollback<br>Worker]-->|scheduled when <br>Canary CRD<br> is in Rollback phase|s
    s-->|Scheduler is called <br>by Manager on <br>CRD change or input output event|m[Manager]
    spec[Canary API Spec]-->|Watch Canary API Types|c[Kubernetes API Client]
    c-->|communcation with Kubernetes API|m
    end
    
   classDef userDefined fill:#0DDF00,stroke:#233,stroke-width:4px;
   class w1,w2,w3,w4,w5,spec,userDefined userDefined;
   classDef generated fill:#fae100,stroke:#233,stroke-width:4px;
   class s,generated generated;
   classDef k8s fill:#eb7b26,stroke:#233,stroke-width:4px;
   class m,c,k8s k8s;

{{< /mermaid >}}

For reference, code for each component can be found in the following table:

Component               | Source Code | Location
------------------------|-------------|-----
Initializing Worker     | [initializing/worker.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/workers/initializing/worker.go) | User Project (user-defined)
Waiting Worker          | [waiting/worker.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/workers/waiting/worker.go) | User Project (user-defined)
Evaluating Worker       | [evaluating/worker.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/workers/evaluating/worker.go) | User Project (user-defined)
Promoting Worker        | [promoting/worker.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/workers/promoting/worker.go) | User Project (user-defined)
Rollback Worker         | [rollback/worker.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/workers/rollback/worker.go) | User Project (user-defined)
Canary Scheduler        | [scheduler/sceduler.go](https://github.com/solo-io/autopilot/blob/master/test/e2e/canary/pkg/scheduler/scheduler.go) | User Project (generated)
Manager                 | [manager.go](https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/manager/manager.go) | Imported from [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
K8s Client              | [client/interfaces.go](https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/client/interfaces.go) | Imported from [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)

The `Manager` connects our operator to the underlying Kubernetes watches that trigger the operator to resync on a change.

The `Manager` calls into our generated `Scheduler` each time a top-level Custom Resource is created, updated, or deleted. A custom scheduler, generated for each Operator, then executes a resync function each time it is called by the operator.

The main control loop of the application looks like the following:

{{<mermaid align="left">}}
 graph TD;
    subgraph main control loop
    m[Manager]-->|receive k8s event|s
    s[Scheduler]-->receive[Receive Resync request]
    receive-->retrieve[Retrieve Canary <br>from Cache]
    retrieve-->eventType[Handle Event Type]
    eventType-->|create/update event|phase[Determine Canary Phase]
    eventType-->|delete event|finalizer[Call Optional User-defined Finalizer]
    phase-->inputs[Read inputs for Phase]
    inputs-->worker[Construct Worker corresponding to Phase]
    worker-->sync[Call worker.Sync function with inputs]
    sync-->|return error|backoff[Retry with Backoff]
    sync-->|return outputs|ensure[Ensure all outputs are written to cluster]
    sync-->|return next phase|next[update the Canary with the new Phase]
    next-->m
     end
     
     
   classDef userDefined fill:#0DDF00,stroke:#233,stroke-width:4px;
   class sync userDefined;
   classDef generated fill:#fae100,stroke:#233,stroke-width:4px;
   class s,receive,retrieve,eventType,phase,finalizer,inputs,worker,backoff,ensure,next generated;
   
   classDef k8s fill:#eb7b26,stroke:#233,stroke-width:4px;
   class m k8s;
{{< /mermaid >}}

The main control loop is started by Autopilot's [`run.Run`](https://github.com/solo-io/autopilot/blob/master/pkg/run/run.go).

Users are expected to modify and update their `spec.go` and `worker.go` files. When the `autopilot.yaml` file changes, 
the project should be regenerated. `worker.go` files will not be overwritten by a regenerate, allowing users to iteratively regenerate the scheduler for their operator. Users can then update their workers as necessary.
