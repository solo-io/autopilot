---
title: "Why Autopilot"
description: "An introduction to Autopilot"
weight: 2
---

This session will introduce and expand upon Autopilot, a new framework for building Kubernetes Operators that gives special attention to monitoring and configuring a Service Mesh.

This session will provide a brief overview of the current ecosystem of Kubernetes Operators and the frameworks available today including the Red Hat Operator SDK and Kubebuilder. 

Next, we will walk through a brief sample of Operator code that demonstrates the additional heavy lifting required to integrate with service mesh components such as Istio/Linkerd, Prometheus, and Jaeger.

Finally, Autopilot will be introduced with a brief demonstration of how we can greatly simplify the above, which will automate watching a mesh-driven metric and reconciling mesh configuration to a desired state.




Autopilot is an open-source framework for developing Kubernetes Operators (inspired by the KubeBuilder operator sdk) that takes an opinionated approach to integrating with service meshes such as Istio and Linkerd, and monitoring components such as Prometheus and Jaeger. 

Autopilot an SDK for Go that provides hooks for building operators that operate both the  Kubernetes and Service Mesh control planes. 

Autopilot operators can monitor and react to events from pluggable backends, such as Prometheus metrics, Jaeger traces, CloudEvents, and Cloud-based webhooks such as GitHub, making it easy to build operators that monitor and react to a plethora of information sources. 

Autopilot operators can publish config changes to Git repos and monitor GitHub for events, making Autopilot ideal for building CI/CD/GitOps solutions. By managing configuration via a Git-based pipeline, Autopilot integrates well with existing tooling such as Weaveworks' Flux.

As users and vendors continue to build custom extensions to Kubernetes, they are already experiencing the pains of integrating with the growing service mesh ecosystem. Autopilot provides users with the ability to write custom operators without having to re-learn and re-implement the logic for operating a mesh.

