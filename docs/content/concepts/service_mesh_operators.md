---
title: "Service Mesh Operators"
description: "Introduction to Service Mesh Operators"
weight: 1
---

# Concepts - Service Mesh Operators

Scale and complexity have made infrastructure configuration impossible via Helm charts and BASH scripts alone. 
As a result, we have seen an explosion of [Kubernetes Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) 
and tooling for building them. Operators effectively extend the API of Kubernetes, 
automating away operational complexity and extending the core API with additional features. 
This has proven vital, for both vendors building Kubernetes-native solutions, 
as well as Kubernetes end-users who require customized automation for their environments.

Simply put, the future of the Kubernetes ecosystem is in Kubernetes Operators. The future of the Service Mesh ecosystem is in Service Mesh 
Operators.

A *Service Mesh Operator* is a  Kubernetes Operator specifically designed to leverage features offered by service meshes - observability, security, load balancing, 
and more. They may also leverage core Kubernetes features such as deployments, services, secrets, and other Kubernetes primitives.

SDKs such as the Operator Framework by RedHat and the *Kubebuilder SDK* Kubernetes-SIG have sprung up to simplify and accelerate
the development of Kubernetes Operators. However, their domain knowledge ends with vanilla Kubernetes, providing no out-of-the-box
integration points with service meshes such as Istio. The work of doing so falls upon the developer. 

For this reason, Solo.io has built Autopilot, an SDK for building *Service Mesh Operators*. 

By treating the mesh as a first-class concept, **Autopilot** makes it easy to build Service Mesh Operators 
that automate and extend meshes similar to how basic Kubernetes Operators do for Kubernetes.

