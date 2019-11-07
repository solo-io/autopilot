# AutoPilot - The Service Mesh Automator

**AutoPilot** is an SDK and toolkit for developing and deploying [service mesh operators](). 

**AutoPilot** generates scaffolding, builds, and deploys Operators which run against a local or remote Kubernetes cluster installed with a Service Mesh. 

**AutoPilot** generated code and libraries provide an easy way to automate configuration and monitoring of a service mesh (and other Kubernetes/infra resources) via the [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) pattern.

# How does it work?

Developers define an `autopilot.yaml` and `autopilot-operator.yaml` which specify the skeleton and configuration of an *AutoPilot Operator*.

AutoPilot makes use of these files to (re-)generate the project skeleton, build, deploy, and manage the lifecycle of the operator via the `ap` CLI.

# How is it different from SDKs like Operator Framework and Kubebuilder?

The [Operator Framework](https://github.com/operator-framework) and [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) are open-ended SDKs that take a far less opinionated approach to building Kubernetes software.

**AutoPilot** provides a *scheduler* on top of the 
where users can write stateless Work functions for various states of their top-level CRD. State information is stored
 on the *status* of the CRD, promoting a stateless design for AutoPilot operators.
 
**AutoPilot** additionally provides primitives, generated code, and helper functions for interacting with a variery of service meshes. While AutoPilot can be used to build operators that do not configure or monitor a mesh, much of *AutoPilot*'s design has been oriented to facilitate easy integration with popular service meshes, as well as the Service Mesh Interface (SMI).

Finally, **AutoPilot** favors simplicity over flexibility, though it is the intention of the project to support the vast majority of DevOps workflows built on top of Kubernetes+Service mesh.

# Roadmap
- Support for managing multiple (remote) clusters.



The `ap` CLI is used to generate code, 

The `ap` CLI is also designed to build and deploy the operator, but this can be done with standard tooling (`go`, `docker`, `kubectl`, etc.). 

AutoPilot provides an opinionated structure 
for executing an operator's 
workflow. Read more about the 
[AutoPilot Architecture]() to learn about 
how AutoPilot Operators schedule and execute work.

Code generation can also be invoked from Go code using the `codegen` package. 

AutoPilot is composed of 3 components:
- `cli`
- `codegen` package
- `pkg` libraries



# todo

- e2e metrics test with istio
- automated e2e with test repo

- configmap with config settings
    - generated from yaml file
    - main func watches for the file, restarts 

- do docs generation for autopilot api
- schema generation


* bake templates into cli
* curl script to download
* add many more aliases/parameter types

- types from config files # partial
    - istio types

- builders

- generic metrics interface / define in autopilot.yaml

- git ops?

- fix the metrics at boot thing

- interactive cli


quarantine <=> operator
|
3 traffic split


deployment => pods


# done 
* works across namespaces..
