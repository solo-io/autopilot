---
title: "Getting Started Pt 1"
weight: 1
description: Initialize a simple operator and deploy it to Kubernetes
---

## Getting Started with Autopilot Part 1 - Initializing the Operator

This is the first part in a series of guides that will walk you through building a minimalistic "AutoRouter" Operator with Autopilot. 

This operator will automatically provide ingress access for selected Kubernetes deployments.

Part 1 of the tutorial introduces the operator and the steps 
to initialize the operator and deploy it to Kubernetes.

In part 2, we'll implement business logic for the operator in the form of "workers" that process the `AutoRoute` CRD.

Finally, in part 3 we'll redeploy the operator with our changes. Then we create an AutoRoute and a deployment, and test the Operator by sending ingress traffic to the deployed container through Istio.

To see the completed code for this tutorial, check out https://github.com/solo-io/autorouter

## About the `AutoRoute` Example Operator

The Operator we will build processes CRDs of kind `AutoRoute` in API group `examples.io/v1`. The `AutoRoute` specifies a `selector` which routes to deployments based on their labels.

The `AutoRoute` Operator will have three phases: `Initializing`, `Syncing`, and `Ready`.

In the `Initializing` phase, the `AutoRoute` will be marked as `Syncing` to confirm processing has begun.

In the `Syncing` phase, the `AutoRoute` ensure Istio provides ingress routes to any deployments matching the `spec.selector` on the `AutoRoute`. 

The `AutoRoute` will ensure that Istio has an ingress route for each selected deployment. Once the routes are created, the state will change to the `Ready` phase to mark that the `AutoRoute` is ready to serve traffic.

In the `Ready` phase, the Operator watch for new deployments. If the list of deployments falls out of sync with the configured routes,
the `AutoRoute` will go back to the `Syncing` phase.

The purpose of this application is to demonstrate basic functionality in Autopilot, as well as provide an example of how to get
started.

# Tutorial

## Prerequisites

- `go`, `docker`, `kubectl` installed locally,
- Istio v1.1 or higher installed in your cluster. This tutorial depends on the `istio-ingressgateway` Gateways/Pods being deployed as well.

## Install the Autopilot CLI

Download the latest release of the Autopilot CLI from the GitHub releases page: https://github.com/solo-io/autopilot/releases

or install from the terminal:

```bash
curl -sL https://run.solo.io/autopilot/install | sh
export PATH=$HOME/.autopilot/bin:$PATH
```

## Initialize the `autorouter` project

Run the following to create the operator directory in your workspace. Can be inside or outside `GOPATH` (uses Go modules):

```bash
ap init autorouter --kind AutoRoute --group examples.io --version v1 
```

> Note: if running outside your $GOPATH, you will need to add the flag
> `--module autorouter.examples.io` 

This should produce the output:

```
INFO[0000] Creating Project Config: kind:"AutoRoute" apiVersion:"examples.io/v1" operatorName:"autoroute-operator" phases:<name:"Initializing" description:"AutoRoute has begun initializing" initial:true outputs:"virtualservices" > phases:<name:"Processing" description:"AutoRoute has begun processing" inputs:"metrics" outputs:"virtualservices" > phases:<name:"Finished" description:"AutoRoute has finished" final:true >
go: creating new go.mod: module autorouter.examples.io
```

If you run `tree` on the newly created `autorouter` directory, you'll see the following files were created:

```
.
└── autorouter
    ├── autopilot-operator.yaml
    ├── autopilot.yaml
    └── go.mod
``` 

* `autopilot.yaml`: the project configuration file. Use this to generate the operator code, build, and deployment files.
* `autopilot-operator.yaml`: the runtime configuration for the Operator.
* `go.mod` - a generated `go.mod` file with some `replace` statements necessary to import Autopilot and its dependencies.

## Update the `autopilot.yaml` file

Let's take a look at the generated `autopilot.yaml` file:

```yaml
apiVersion: examples.io/v1
kind: AutoRoute
operatorName: autoroute-operator
phases:
- description: AutoRoute has begun initializing
  initial: true
  name: Initializing
  outputs:
  - virtualservices
- description: AutoRoute has begun processing
  inputs:
  - metrics
  name: Processing
  outputs:
  - virtualservices
- description: AutoRoute has finished
  final: true
  name: Finished
```

This is a default template and meant to be edited. Let's replace the contents with the following:

```yaml
apiVersion: examples.io/v1
kind: AutoRoute
operatorName: autoroute-operator
phases:

- description: AutoRoute is pending processing
  # initial: true indicates this state is the initial state of the CRD
  # initial can only be true for one phase
  initial: true
  name: Initializing
  
- description: AutoRoute is syncing with deployments in the cluster
  name: Syncing
  inputs:
  - deployments
  outputs:
  - services
  - virtualservices
  - gateways

- description: AutoRoute is in-sync and ready to serve traffic
  inputs:
  - deployments
  name: Ready
```

Our `autopilot.yaml` specifies each of the *phases* the `AutoRoute` can be in, as well as the 
inputs and outputs used by the system during the given phase.

## Generate the code

We can now generate our project structure by issuing the following command:

```bash
cd autorouter
ap generate
```

The output from `ap` should verify that all files generated successfully:

```
INFO[0000] skippinng file cmd/autoroute-operator/main.go because it exists
INFO[0000] Writing pkg/scheduler/scheduler.go
INFO[0001] Writing pkg/parameters/parameters.go
INFO[0001] Writing pkg/apis/autoroutes/v1/doc.go
INFO[0001] Writing pkg/apis/autoroutes/v1/phases.go
INFO[0001] Writing pkg/apis/autoroutes/v1/register.go
INFO[0001] Writing pkg/apis/autoroutes/v1/spec.go
INFO[0001] Writing pkg/apis/autoroutes/v1/types.go
INFO[0001] Writing build/Dockerfile
INFO[0001] Writing build/bin/user_setup
INFO[0001] Writing build/bin/entrypoint
INFO[0001] Writing deploy/crd.yaml
INFO[0001] Writing deploy/deployment-single-namespace.yaml
INFO[0001] Writing deploy/deployment-all-namespaces.yaml
INFO[0001] Writing deploy/configmap.yaml
INFO[0001] Writing deploy/role.yaml
INFO[0001] Writing deploy/rolebinding.yaml
INFO[0001] Writing deploy/clusterrole.yaml
INFO[0001] Writing deploy/clusterrolebinding.yaml
INFO[0001] Writing deploy/service_account.yaml
INFO[0001] skippinng file hack/create_cr_yaml.go because it exists
INFO[0001] Writing deploy/autoroute_example.yaml
INFO[0001] skippinng file hack/boilerplate/boilerplate.go.txt because it exists
INFO[0001] Writing .gitignore
INFO[0001] Writing pkg/workers/initializing/inputs_outputs.go
INFO[0001] Writing pkg/workers/initializing/worker.go
INFO[0001] Writing pkg/workers/syncing/inputs_outputs.go
INFO[0001] Writing pkg/workers/syncing/worker.go
INFO[0001] Writing pkg/workers/ready/inputs_outputs.go
INFO[0001] Writing pkg/workers/ready/worker.go
INFO[0001] Generating Deepcopy types for autorouter.examples.io/pkg/apis/autoroutes/v1
INFO[0001] Generating Deepcopy code for API: &args.GeneratorArgs{InputDirs:[]string{"./pkg/apis/autoroutes/v1"}, OutputBase:"", OutputPackagePath:"./pkg/apis/autoroutes/v1", OutputFileBaseName:"zz_generated.deepcopy", GoHeaderFilePath:"hack/boilerplate/boilerplate.go.txt", GeneratedByCommentTemplate:"// Code generated by GENERATOR_NAME. DO NOT EDIT.", VerifyOnly:false, IncludeTestFiles:false, GeneratedBuildTag:"ignore_autogenerated", CustomArgs:(*generators.CustomArgs)(0xc00069d1c0), defaultCommandLineFlags:false}
INFO[0005] Finished generating examples.io/v1.AutoRoute
```

## Build and Deploy the Operator

Our Operator should already be ready to build and deploy to Kubernetes! Let's try it out!

First, to build:

```bash
ap build <image tag>
```

Should yield the output:

```
INFO[0004] Building OCI image <image tag>
Sending build context to Docker daemon  43.88MB
Step 1/7 : FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
 ---> 8c980b20fbaa
Step 2/7 : ENV OPERATOR=/usr/local/bin/autoroute-operator USER_UID=1001 USER_NAME=autoroute-operator
 ---> Running in 290f289d161f
Removing intermediate container 290f289d161f
 ---> 85a24b29a46a
Step 3/7 : COPY build/_output/bin/autoroute-operator ${OPERATOR}
 ---> 79b2a88a152a
Step 4/7 : COPY build/bin /usr/local/bin
 ---> 1381ca562c4d
Step 5/7 : RUN  /usr/local/bin/user_setup
 ---> Running in 646e6e1b2447
+ mkdir -p /root
+ chown 1001:0 /root
+ chmod ug+rwx /root
+ chmod g+rw /etc/passwd
+ rm /usr/local/bin/user_setup
Removing intermediate container 646e6e1b2447
 ---> acd518dd2c4c
Step 6/7 : ENTRYPOINT ["/usr/local/bin/entrypoint"]
 ---> Running in 35b6a656d5ba
Removing intermediate container 35b6a656d5ba
 ---> e3146307a863
Step 7/7 : USER ${USER_UID}
 ---> Running in a702c45db3dc
Removing intermediate container a702c45db3dc
 ---> 7156f513ea13
Successfully built 7156f513ea13
Successfully tagged <image tag>
INFO[0012] Operator build complete.
```

Next, we'll deploy:

```bash
ap deploy <image tag> -p
```

> Note: Omit the `-p` flag if you wish to skip the `docker push` step.

Should yield the output:

```
INFO[0000] Deploying Operator with image <image tag>
INFO[0000] Pushing image <image tag>
The push refers to repository [<image tag>]
427b6a62465c: Pushed
cc5cff924b4a: Pushed
0a1eccf95daf: Pushed
b6f081e4b2b6: Mounted from ilackarms/example
d8e1f35641ac: Mounted from ilackarms/example
latest: digest: sha256:cd9c51c70df98176b76574642656050716f432404eec0f2b2d09b476e1e5a87d size: 1363
namespace/autoroute-operator created
INFO[0012] Deploying crd.yaml
customresourcedefinition.apiextensions.k8s.io/autoroutes.examples.io created
INFO[0013] Deploying configmap.yaml
configmap/autoroute-operator created
INFO[0013] Deploying service_account.yaml
serviceaccount/autoroute-operator created
INFO[0014] Deploying clusterrole.yaml
clusterrole.rbac.authorization.k8s.io/autoroute-operator created
INFO[0014] Deploying clusterrolebinding.yaml
clusterrolebinding.rbac.authorization.k8s.io/autoroute-operator created
INFO[0015] Deploying deployment-all-namespaces.yaml
deployment.apps/autoroute-operator created
INFO[0015] Operator deployment complete.
```

If everything worked correctly, we should see the `autoroute-operator` namespace created in Kubernetes:

```bash
kubectl get ns
```

```
NAME                 STATUS   AGE
autoroute-operator   Active   19s
default              Active   5h29m
kube-public          Active   5h29m
kube-system          Active   5h29m
```

Let's see that our operator is running:

```bash
kubectl get pod -n autoroute-operator
```

```
NAME                                 READY   STATUS    RESTARTS   AGE
autoroute-operator-ccbd545c6-ll7xh   1/1     Running   0          5s
```

We can tail logs from the pod as well (`Ctrl^C` to exit):

```
ap logs -f
```

```json
{"level":"info","ts":1573768902.359966,"msg":"Starting Operator with config","config":"version:\"0.0.1\" controlPlaneNs:\"istio-system\" workInterval:<seconds:5 > metricsAddr:\":9091\" enableLeaderElection:true logLevel:<value:1 > "}
{"level":"info","ts":1573768902.3601103,"msg":"Warning: Flushing Operator Metrics!"}
{"level":"info","ts":1573768903.0675852,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":":9091"}
{"level":"info","ts":1573768903.0686636,"msg":"Registering watch for primary resource AutoRoute"}
{"level":"info","ts":1573768903.0687194,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573768903.0688958,"msg":"Registering watch for output resource Services"}
{"level":"info","ts":1573768903.0689237,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573768903.0690174,"msg":"Registering watch for output resource VirtualServices"}
{"level":"info","ts":1573768903.0690434,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573768903.069132,"msg":"Registering watch for output resource Gateways"}
{"level":"info","ts":1573768903.0691571,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"autoRoute-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1573768903.0693312,"logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
^C
```

## Test the Operator with an AutoRoute Resource

Cool! We've already gotten our first Service Mesh Operator built and deployed! But it's not doing anything yet, right? 

Let's see what happens when we create an `AutoRoute` resource (an example of which Autopilot conveniently generates in the `deploy/` directory):

```bash
kubectl create ns example
kubectl apply -n example -f deploy/autoroute_example.yaml
```

```bash
autoroute.examples.io/example created
```

Now we'll notice the pod is crashing:

```bash
kubectl get pod -n autoroute-operator
```
```bash
autoroute-operator-ccbd545c6-sj6tx   0/1     CrashLoopBackOff   2          4m1s
```
```bash
ap logs
```
```bash
...
E1114 22:06:10.089481       1 runtime.go:67] Observed a panic: implement me!
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/runtime/runtime.go:76
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/runtime/runtime.go:65
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/runtime/runtime.go:51
/usr/local/Cellar/go/1.13.3/libexec/src/runtime/panic.go:679
/Users/ilackarms/workspace/goprojects/autorouter/pkg/workers/initializing/worker.go:20
/Users/ilackarms/workspace/goprojects/autorouter/pkg/scheduler/scheduler.go:145
/Users/ilackarms/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.3.0/pkg/internal/controller/controller.go:216
/Users/ilackarms/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.3.0/pkg/internal/controller/controller.go:192
/Users/ilackarms/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.3.0/pkg/internal/controller/controller.go:171
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/wait/wait.go:152
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/wait/wait.go:153
/Users/ilackarms/go/pkg/mod/k8s.io/apimachinery@v0.0.0-20190404173353-6a84e37a896d/pkg/util/wait/wait.go:88
/usr/local/Cellar/go/1.13.3/libexec/src/runtime/asm_amd64.s:1357
panic: implement me! [recovered]
	panic: implement me!

```

We haven't implemented anything yet, and our Operator is letting us know by panicking. Let's go ahead and start implementing our 
operator!

Continue to [part 2]({{< versioned_link_path fromRoot="/tutorial_code/getting_started_2">}}) to start implementing 
the business logic for our service mesh operator.

