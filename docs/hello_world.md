# HelloWorld with Autopilot - AutoRouter

This guide will walk you through building a minimalistic "HelloWorld" Operator with Autopilot. 
This operator will automatically generate ingress routes for discovered deployments.

The Operator will process CRDs of kind `AutoRoute` in API group `example.io/v1`.

The `AutoRoute` CRD will have three phases: `Initializing`, `Syncing`, and `Ready`.

In the `Initializing` phase, the `AutoRoute` will be marked as `Syncing` to confirm processing has begun.

In the `Syncing` phase, the `AutoRoute` ensure Istio provides ingress routes to any deployments matching the `spec.selector` on the `AutoRoute`. 
 The `AutoRoute` will ensure that Istio has an ingress route for each selected deployment. Once the routes are created, the 
 state will change to the `Ready` phase to mark that the `AutoRoute` is ready to serve traffic.

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

## Initialize the `autorouter` operator

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
INFO[0001] Generating Deepcopy types for autorouter.example.io/pkg/apis/autoroutes/v1
INFO[0001] Generating Deepcopy code for API: &args.GeneratorArgs{InputDirs:[]string{"./pkg/apis/autoroutes/v1"}, OutputBase:"", OutputPackagePath:"./pkg/apis/autoroutes/v1", OutputFileBaseName:"zz_generated.deepcopy", GoHeaderFilePath:"hack/boilerplate/boilerplate.go.txt", GeneratedByCommentTemplate:"// Code generated by GENERATOR_NAME. DO NOT EDIT.", VerifyOnly:false, IncludeTestFiles:false, GeneratedBuildTag:"ignore_autogenerated", CustomArgs:(*generators.CustomArgs)(0xc00069d1c0), defaultCommandLineFlags:false}
INFO[0005] Finished generating examples.io/v1.AutoRoute
```

## Test the deployment

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
ap deploy <image tag>
```

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


## Update the API Spec
## Re-Generate the code
## Write an example CRD
## Update the Initializing Worker
## Update the Replying Worker
## Update the Resting Worker
## Redeploy
## Try it out!
