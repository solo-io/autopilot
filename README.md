
<h1 align="center">
    <img src="https://github.com/solo-io/autopilot/blob/master/docs/content/img/logo.png?raw=true" alt="Autopilot" width="260" height="242">
  <br>
  The Service Mesh SDK
</h1>

**Autopilot** is an SDK and toolkit for developing and deploying [service mesh operators](docs/content/concepts/service_mesh_operators.md). 

**Autopilot** generates scaffolding, builds, and deploys Operators which run against a local or remote Kubernetes cluster installed with a Service Mesh. 

**Autopilot** generated code and libraries provide an easy way to automate configuration and monitoring of a service mesh (and other Kubernetes/infra resources) via the [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) pattern.

[**Installation**](https://autopilot.solo.io/installation/) &nbsp; |
&nbsp; [**Documentation**](https://autopilot.solo.io) &nbsp; |
&nbsp; [**Blog**](TODO LINK) &nbsp; |
&nbsp; [**Slack**](https://slack.solo.io) &nbsp; |
&nbsp; [**Twitter**](https://twitter.com/soloio_inc)

<BR><center><img src=""https://github.com/solo-io/autopilot/blob/master/docs/content/img/architecture.png?raw=true" alt="Autopilot Architecture" width="906"></center>

# How does it work?

Developers define an `autopilot.yaml` and `autopilot-operator.yaml` which specify the skeleton and configuration of an *Autopilot Operator*.

Autopilot makes use of these files to (re-)generate the project skeleton, build, deploy, and manage the lifecycle of the operator via the `ap` CLI.

Users place their API in a generated `spec.go` file, and business logic in generated `worker.go` files. Once these files have been modified, they will not be overwritten by `ap generate`.

# How is it different from SDKs like Operator Framework and Kubebuilder?

The [Operator Framework](https://github.com/operator-framework) and [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) are open-ended SDKs that take a far less opinionated approach to building Kubernetes software.

**Autopilot** provides a more opinionated control loop via a generated *scheduler* that implements the [Controller-Runtime Reconciler interface](https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/reconcile/reconcile.go#L80), for which users write stateless Work functions for various states of their top-level CRD. State information is stored
 on the *status* of the CRD, promoting a stateless design for Autopilot operators.
 
**Autopilot** additionally provides primitives, generated code, and helper functions for interacting with a variery of service meshes. While Autopilot can be used to build operators that do not configure or monitor a mesh, much of *Autopilot*'s design has been oriented to facilitate easy integration with popular service meshes.

Finally, **Autopilot** favors simplicity over flexibility, though it is the intention of the project to support the vast majority of DevOps workflows built on top of Kubernetes+Service mesh.

## Next Steps
- Join us on our Slack channel: [https://slack.solo.io/](https://slack.solo.io/)
- Follow us on Twitter: [https://twitter.com/soloio_inc](https://twitter.com/soloio_inc)
- Check out the docs: [https://gloo.solo.io](https://gloo.solo.io)
- Check out the code and contribute: [Contribution Guide](CONTRIBUTING.md)
- Contribute to the [Docs](https://github.com/solo-io/solo-docs)

### Thanks

**Gloo** would not be possible without the valuable open-source work of projects in the community. We would like to extend a special thank-you to [Envoy](https://www.envoyproxy.io).





# Requirements

- in $GOPATH
    - k8s.io/gengo/boilerplate/boilerplate.go.txt

# Hello World



# Roadmap
- Support for managing multiple (remote) clusters.

## scrap

Autopilot provides an opinionated structure 
for executing an operator's 
workflow. Read more about the 
[Autopilot Architecture]() to learn about 
how Autopilot Operators schedule and execute work.

Code generation can also be invoked from Go code using the `codegen` package. 

Autopilot is composed of 3 components:
- `cli`
- `codegen` package
- `pkg` libraries



# todo

# cleanup
- example
- docs 
- improve docs generation template
- bake templates into cli
- clean up CLI messages

- idempotent generation of rbac yaml (rule ordering not idempotent)

## test
- e2e metrics test with istio

## features
- git ops
- define custom metrics queries in autopilot.yaml
- validate method for project config
    - check operatorName is kube compliant
    - apiVerson, kind, phases are correct
    - customParameters
    - final phase with i/o
- add user config to configmap with config settings
- curl script to download
- builders
- ap undeploy
- label everything for easy deletion/listing
- expose garbage collection func to workers
    - rollback the phase when something ensure fails? (option in config)
- multiple crds

## punt
- schema generation
- interactive cli
- automatic metrics for worker syncs
- automatic traces for worker syncs
- option to make workers persistent






# docs todos:
- architecture description. how does my Autopilot operator work?
    - user-project directory structure 
- how does autopilot generate code? when do i regenerate? when do i redeploy?
- autopilot libraries/pkg directory structure
- e2e hello world guide
    - tour-through-your-hello-world package-by-package
- tutorial through e2e



# done 
* works across namespaces..


# code guide:

explain where all the existing things are - the templated queries and clients

folder -> what does it do